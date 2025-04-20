package sshserver

import (
	"errors"
	"fmt"
	"group-ssh-chat/auth"
	"group-ssh-chat/commands"
	"group-ssh-chat/ui"
	"log"
	"net"
	"os"
	"sync"
	"syscall"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/ssh"
	"golang.org/x/term"
)

// SSHServer manages SSH connections, client sessions, and message broadcasting
type SSHServer struct {
	activeClientsMap   map[string][]clientSSHSession // Maps usernames to their active sessions
	activeClientsMutex sync.Mutex                    // Mutex to protect concurrent access to the map
	sshServerConfig    *ssh.ServerConfig              // SSH server configuration
	tcpListener        net.Listener                   // TCP listener for incoming connections
	commandManager     *commands.CommandManager        // Command manager for handling slash commands
}

// clientSSHSession represents an active client terminal session
// Each SSH connection can have multiple sessions
type clientSSHSession struct {
	terminal   *term.Terminal       // Terminal interface for reading/writing to the session
	connection *ssh.ServerConn      // The SSH connection this session belongs to
	id         string               // Unique identifier for this session
	ui         *ui.SSHTerminalBridge // UI bridge for tview integration
	channel    ssh.Channel          // SSH channel for this session
}

// New creates and initializes a new SSH server instance
// It sets up the server configuration with the provided authentication handler
// and initializes the TCP listener based on environment variables
func New(sauth *auth.SSHAuth) *SSHServer {
	ss := &SSHServer{
		activeClientsMap: make(map[string][]clientSSHSession),
		sshServerConfig: &ssh.ServerConfig{
			// Comment below to disable password auth.
			// PasswordCallback: sauth.HandlePasswordLogin,

			PublicKeyCallback: sauth.HandlePublicKeyLogin,
		},
		commandManager: commands.NewCommandManager(),
	}

	ss.sshServerConfig.AddHostKey(sauth.HostSSHPrivateKey)
	ss.initListener()

	// Register commands
	ss.registerCommands()

	return ss
}

// Initializes a tcp listener on host and port
func (ss *SSHServer) initListener() {
	svrAddress := fmt.Sprintf("%s:%s", os.Getenv("SSH_SERVER_HOST"), os.Getenv("SSH_SERVER_PORT"))
	listener, err := net.Listen("tcp", svrAddress)
	if err != nil {
		log.Fatal("failed to listen for connection: ", err)
	}

	ss.tcpListener = listener
}

// Accepts tcp connections and makes the ssh handshake
func (ss *SSHServer) AcceptConnections() {
	for {
		nConn, err := ss.tcpListener.Accept()
		if err != nil {
			// Handle specific error types that are known to be temporary
			if errors.Is(err, syscall.EINTR) || errors.Is(err, syscall.EAGAIN) || errors.Is(err, syscall.EWOULDBLOCK) {
				log.Printf("temporary error accepting connection: %v", err)
				continue
			}
			log.Fatalf("fatal error accepting connection: %v", err)
		}

		// Before use, a handshake must be performed on the incoming
		// net.Conn.
		conn, chans, reqs, err := ssh.NewServerConn(nConn, ss.sshServerConfig)
		if err != nil {
			log.Printf("failed to handshake: %v", err)
			nConn.Close()
			continue
		}
		log.Printf("user %s logged in with key %s", conn.User(), conn.Permissions.Extensions["pubkey-fp"])
		go ss.handleConnection(conn, chans, reqs)
	}
}

// Handles a single ssh connection and manages the channels from the connection
func (ss *SSHServer) handleConnection(conn *ssh.ServerConn, chans <-chan ssh.NewChannel, reqs <-chan *ssh.Request) {
	// Handle connection closure
	defer func() {
		log.Printf("User %s disconnected", conn.User())
		conn.Close()
	}()

	// Discard global requests
	go ssh.DiscardRequests(reqs)

	// Service the incoming Channel channels.
	for channelReq := range chans {
		// Channels have a type, depending on the application level
		// protocol intended. In the case of a shell or pty-req, the type is
		// "session"
		if channelReq.ChannelType() != "session" {
			log.Printf("Rejecting channel type: %s", channelReq.ChannelType())
			channelReq.Reject(ssh.UnknownChannelType, "unknown channel type")
			continue
		}

		sessionChannel, sshRequests, err := channelReq.Accept()
		if err != nil {
			log.Printf("Could not accept channel: %v", err)
			continue
		}

		// Create a basic terminal for fallback
		termSession := term.NewTerminal(sessionChannel, "> ")

		// Create a unique session ID
		sessionID := uuid.New().String()
		log.Printf("New session created for user %s with ID %s", conn.User(), sessionID)

		// Create the UI bridge
		uiBridge := ui.NewSSHTerminalBridge(sessionChannel, conn.User())

		// Add client session to active clients map
		ss.activeClientsMutex.Lock()
		clientsess := clientSSHSession{
			terminal:   termSession,
			connection: conn,
			id:         sessionID,
			ui:         uiBridge,
			channel:    sessionChannel,
		}
		_, ok := ss.activeClientsMap[conn.User()]
		if !ok {
			ss.activeClientsMap[conn.User()] = make([]clientSSHSession, 0, 1)
		}
		ss.activeClientsMap[conn.User()] = append(
			ss.activeClientsMap[conn.User()],
			clientsess,
		)

		// Update the user list for all clients
		ss.updateAllUserLists()
		ss.activeClientsMutex.Unlock()

		// Set up the input handler for the UI
		uiBridge.SetInputHandler(func(message string) {
			// Store a reference to the current session for use in the closure
			currentSession := clientsess

			// Check if the message is a command
			handled, err := ss.commandManager.HandleCommand(message, conn.User())
			if err != nil {
				// Send error message to the user
				currentSession.ui.AddSystemMessage(fmt.Sprintf("Error: %s", err.Error()))
				return
			}

			// If not a command, broadcast the message
			if !handled {
				ss.broadcastMessage(conn.User(), message)
			}
		})

		// Broadcast user joined message to other users
		ss.broadcastSystemMessageExcept(fmt.Sprintf("%s has joined the chat", conn.User()), conn.User())

		// We'll handle welcome messages in the handleUISession method

		// Handle session in a separate goroutine
		go ss.handleUISession(conn.User(), &clientsess)

		// Sessions have out-of-band requests such as "shell",
		// "pty-req" and "env".
		go ss.handleSSHRequests(sshRequests)
	}
}

// Handles a UI session
func (ss *SSHServer) handleUISession(user string, clientsess *clientSSHSession) {
	defer func() {
		log.Printf("User %s disconnected", user)
		clientsess.connection.Close()
		ss.removeClientSession(clientsess.id, true)
		// Notify other users that this user has left
		ss.broadcastSystemMessage(fmt.Sprintf("%s has left the chat", user))
	}()

	// Add a small delay before sending the first message to ensure terminal is ready
	time.Sleep(100 * time.Millisecond)

	// Add welcome messages as system messages
	clientsess.ui.AddSystemMessage("Welcome to Group SSH Chat!")

	// Add a small delay between messages
	time.Sleep(100 * time.Millisecond)

	clientsess.ui.AddSystemMessage(fmt.Sprintf("You are logged in as %s", user))

	// Add a small delay before sending the user list
	time.Sleep(200 * time.Millisecond)

	// Get a list of all usernames
	ss.activeClientsMutex.Lock()
	usernames := make([]string, 0, len(ss.activeClientsMap))
	for username := range ss.activeClientsMap {
		usernames = append(usernames, username)
	}
	ss.activeClientsMutex.Unlock()

	// Update the user list for this client
	clientsess.ui.UpdateUserList(usernames)

	// Start the UI - this will block until the session ends
	err := clientsess.ui.Start()
	if err != nil {
		log.Printf("UI error for user %s: %v", user, err)
	}
}

// broadcastMessage sends a message from a user to all connected clients
func (ss *SSHServer) broadcastMessage(user string, message string) {
	// Log the message to the server console
	log.Printf("%s: %s", user, message)

	// Get a snapshot of sessions to message while holding the lock
	ss.activeClientsMutex.Lock()
	var sessionsToMessage []clientSSHSession
	for _, sessions := range ss.activeClientsMap {
		sessionsToMessage = append(sessionsToMessage, sessions...)
	}
	ss.activeClientsMutex.Unlock()

	// Send messages and track failed sessions
	var failedSessionIDs []string
	for _, cs := range sessionsToMessage {
		// Try to add the message, if it fails, track the session for removal
		if cs.ui == nil {
			failedSessionIDs = append(failedSessionIDs, cs.id)
			continue
		}
		cs.ui.AddMessage(user, message)
	}

	// Remove failed sessions outside the message loop
	if len(failedSessionIDs) > 0 {
		for _, sessionID := range failedSessionIDs {
			ss.removeClientSession(sessionID, true)
		}
	}
}

// broadcastSystemMessage sends a system message to all connected clients
func (ss *SSHServer) broadcastSystemMessage(message string) {
	// Get a snapshot of sessions to message while holding the lock
	ss.activeClientsMutex.Lock()
	var sessionsToMessage []clientSSHSession
	for _, sessions := range ss.activeClientsMap {
		sessionsToMessage = append(sessionsToMessage, sessions...)
	}
	ss.activeClientsMutex.Unlock()

	// Log the system message to the server console
	log.Printf("SYSTEM: %s", message)

	// Send messages and track failed sessions
	var failedSessionIDs []string
	for _, cs := range sessionsToMessage {
		// Try to add the message, if it fails, track the session for removal
		if cs.ui == nil {
			failedSessionIDs = append(failedSessionIDs, cs.id)
			continue
		}
		cs.ui.AddSystemMessage(message)
	}

	// Remove failed sessions outside the message loop
	if len(failedSessionIDs) > 0 {
		for _, sessionID := range failedSessionIDs {
			ss.removeClientSession(sessionID, true)
		}
	}
}

// broadcastSystemMessageExcept sends a system message to all connected clients except the specified username
func (ss *SSHServer) broadcastSystemMessageExcept(message string, exceptUsername string) {
	// Get a snapshot of sessions to message while holding the lock
	ss.activeClientsMutex.Lock()
	var sessionsToMessage []clientSSHSession
	for username, sessions := range ss.activeClientsMap {
		if username != exceptUsername {
			sessionsToMessage = append(sessionsToMessage, sessions...)
		}
	}
	ss.activeClientsMutex.Unlock()

	// Log the system message to the server console
	log.Printf("SYSTEM: %s", message)

	// Send messages and track failed sessions
	var failedSessionIDs []string
	for _, cs := range sessionsToMessage {
		// Try to add the message, if it fails, track the session for removal
		if cs.ui == nil {
			failedSessionIDs = append(failedSessionIDs, cs.id)
			continue
		}
		cs.ui.AddSystemMessage(message)
	}

	// Remove failed sessions outside the message loop
	if len(failedSessionIDs) > 0 {
		for _, sessionID := range failedSessionIDs {
			ss.removeClientSession(sessionID, true)
		}
	}
}

// removes the client session based on the session ID
// lock parameter determines whether to acquire the mutex lock (should be true if called from outside a locked context)
func (ss *SSHServer) removeClientSession(sessionId string, lock bool) {
	if lock {
		ss.activeClientsMutex.Lock()
		defer ss.activeClientsMutex.Unlock()
	}

	for user, sessions := range ss.activeClientsMap {
		var updatedSessions []clientSSHSession
		var sessionRemoved bool

		for _, session := range sessions {
			if session.id != sessionId {
				updatedSessions = append(updatedSessions, session)
			} else {
				sessionRemoved = true
			}
		}

		// Only update and log if we actually removed a session
		if sessionRemoved {
			// Update the map with the filtered sessions
			ss.activeClientsMap[user] = updatedSessions
			log.Println("Removed Session:", sessionId)

			if len(ss.activeClientsMap[user]) == 0 {
				delete(ss.activeClientsMap, user)
				log.Println("Removed all channels for:", user)
			}

			// Update user lists for all remaining clients
			ss.updateAllUserLists()

			// We found and removed the session, no need to continue searching
			break
		}
	}
}

// Handles ssh requests and replies to them to service the ssh connection
func (ss *SSHServer) handleSSHRequests(sshRequests <-chan *ssh.Request) {
	for req := range sshRequests {
		switch req.Type {
		case "pty-req":
			// Parse terminal type from payload
			if len(req.Payload) >= 4 {
				termLen := req.Payload[3]
				if len(req.Payload) >= int(termLen)+4 {
					// term := string(req.Payload[4 : termLen+4])
					// log.Printf("PTY requested: %s", term)
				} else {
					log.Printf("Invalid PTY request: payload too short")
				}
			} else {
				log.Printf("Invalid PTY request: payload too short")
			}

			if req.WantReply {
				req.Reply(true, nil)
			}

		case "shell":
			// log.Printf("Shell requested")
			if req.WantReply {
				req.Reply(true, nil)
			}

		case "window-change":
			// Window size change request
			if req.WantReply {
				req.Reply(true, nil)
			}

		default:
			log.Printf("Received request of type: %s", req.Type)
			if req.WantReply {
				req.Reply(false, nil)
			}
		}
	}
}

// updateAllUserLists updates the user list for all connected clients
func (ss *SSHServer) updateAllUserLists() {
	// Get a list of all usernames
	usernames := make([]string, 0, len(ss.activeClientsMap))
	for username := range ss.activeClientsMap {
		usernames = append(usernames, username)
	}

	// Update each client's user list
	for _, sessions := range ss.activeClientsMap {
		for _, session := range sessions {
			session.ui.UpdateUserList(usernames)
		}
	}
}

// registerCommands registers all available commands with the command manager
func (ss *SSHServer) registerCommands() {
	// Help command
	helpCmd := commands.NewHelpCommand(ss.commandManager, func(message string) {
		// Broadcast to the current user only
		ss.activeClientsMutex.Lock()
		defer ss.activeClientsMutex.Unlock()

		// Find the current user's sessions
		for _, sessions := range ss.activeClientsMap {
			for _, session := range sessions {
				session.ui.AddSystemMessage(message)
			}
		}
	})
	ss.commandManager.RegisterCommand("help", helpCmd)

	// Users command
	usersCmd := commands.NewUsersCommand(
		func() []string {
			// Get a list of all usernames
			ss.activeClientsMutex.Lock()
			defer ss.activeClientsMutex.Unlock()

			usernames := make([]string, 0, len(ss.activeClientsMap))
			for username := range ss.activeClientsMap {
				usernames = append(usernames, username)
			}
			return usernames
		},
		func(message string) {
			// Broadcast to the current user only
			ss.activeClientsMutex.Lock()
			defer ss.activeClientsMutex.Unlock()

			// Find the current user's sessions
			for _, sessions := range ss.activeClientsMap {
				for _, session := range sessions {
					session.ui.AddSystemMessage(message)
				}
			}
		},
	)
	ss.commandManager.RegisterCommand("users", usersCmd)

	// Whisper command
	whisperCmd := commands.NewWhisperCommand(
		func(sender, recipient, message string) {
			// Send private message
			ss.activeClientsMutex.Lock()
			defer ss.activeClientsMutex.Unlock()

			// Find the sender's sessions
			for _, session := range ss.activeClientsMap[sender] {
				session.ui.AddSystemMessage(fmt.Sprintf("To %s: %s", recipient, message))
			}

			// Find the recipient's sessions
			for _, session := range ss.activeClientsMap[recipient] {
				session.ui.AddSystemMessage(fmt.Sprintf("From %s: %s", sender, message))
			}
		},
		func(username string) bool {
			// Check if user exists
			ss.activeClientsMutex.Lock()
			defer ss.activeClientsMutex.Unlock()

			_, exists := ss.activeClientsMap[username]
			return exists
		},
		func(message string) {
			// Broadcast to the current user only
			ss.activeClientsMutex.Lock()
			defer ss.activeClientsMutex.Unlock()

			// Find the current user's sessions
			for _, sessions := range ss.activeClientsMap {
				for _, session := range sessions {
					session.ui.AddSystemMessage(message)
				}
			}
		},
	)
	ss.commandManager.RegisterCommand("whisper", whisperCmd)

	// Clear command
	clearCmd := commands.NewClearCommand(
		func() {
			// Clear the screen for the current user
			ss.activeClientsMutex.Lock()
			defer ss.activeClientsMutex.Unlock()

			// Find the current user's sessions
			for _, sessions := range ss.activeClientsMap {
				for _, session := range sessions {
					session.ui.ClearScreen()
				}
			}
		},
		func(message string) {
			// Broadcast to the current user only
			ss.activeClientsMutex.Lock()
			defer ss.activeClientsMutex.Unlock()

			// Find the current user's sessions
			for _, sessions := range ss.activeClientsMap {
				for _, session := range sessions {
					session.ui.AddSystemMessage(message)
				}
			}
		},
	)
	ss.commandManager.RegisterCommand("clear", clearCmd)
}
