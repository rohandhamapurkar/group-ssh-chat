package sshserver

import (
	"fmt"
	"group-ssh-chat/auth"
	"log"
	"net"
	"os"
	"sync"

	"github.com/google/uuid"
	"golang.org/x/crypto/ssh"
	"golang.org/x/term"
)

// An SSHServer is represented by custom struct
type SSHServer struct {
	activeClientsMap   map[string][]clientSSHSession
	activeClientsMutex sync.Mutex
	sshServerConfig    *ssh.ServerConfig
	tcpListener        net.Listener
}

type clientSSHSession struct {
	terminal   *term.Terminal
	connection *ssh.ServerConn
	id         string
}

// Returns new instance of the ssh server
func New(sauth *auth.SSHAuth) *SSHServer {
	ss := &SSHServer{
		activeClientsMap: make(map[string][]clientSSHSession),
		sshServerConfig: &ssh.ServerConfig{
			// Comment below to disable password auth.
			// PasswordCallback: sauth.HandlePasswordLogin,

			PublicKeyCallback: sauth.HandlePublicKeyLogin,
		},
	}

	ss.sshServerConfig.AddHostKey(sauth.HostSSHPrivateKey)
	ss.initListener()

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
			log.Printf("failed to accept incoming connection: %q", err)
			continue
		}

		// Before use, a handshake must be performed on the incoming
		// net.Conn.
		conn, chans, reqs, err := ssh.NewServerConn(nConn, ss.sshServerConfig)
		if err != nil {
			log.Printf("failed to handshake: %q", err)
			continue
		}
		log.Printf("logged in with key %s", conn.Permissions.Extensions["pubkey-fp"])
		go ss.handleConnection(conn, chans, reqs)

	}
}

// Handles a single ssh connection and manages the channels from the connection
func (ss *SSHServer) handleConnection(conn *ssh.ServerConn, chans <-chan ssh.NewChannel, reqs <-chan *ssh.Request) {
	go ssh.DiscardRequests(reqs)

	// Service the incoming Channel channels.
	for channelReq := range chans {
		// Channels have a type, depending on the application level
		// protocol intended. In the case of a shell or pty-req, the type is
		// "session"
		if channelReq.ChannelType() != "session" {
			channelReq.Reject(ssh.UnknownChannelType, "unknown channel type")
			continue
		}

		sessionChannel, sshRequests, err := channelReq.Accept()
		if err != nil {
			log.Printf("Could not accept channel: %v", err)
			continue
		}

		termSession := term.NewTerminal(sessionChannel, "> ")

		ss.activeClientsMutex.Lock()
		clientsess := clientSSHSession{
			terminal:   termSession,
			connection: conn,
			id:         uuid.New().String(),
		}
		_, ok := ss.activeClientsMap[conn.User()]
		if !ok {
			ss.activeClientsMap[conn.User()] = make([]clientSSHSession, 0)
		}
		ss.activeClientsMap[conn.User()] = append(
			ss.activeClientsMap[conn.User()],
			clientsess,
		)
		ss.activeClientsMutex.Unlock()

		go ss.handleSessionInput(conn.User(), &clientsess)

		// Sessions have out-of-band requests such as "shell",
		// "pty-req" and "env".
		go ss.handleSSHRequests(sshRequests)
	}
}

// Handles text input from the client session channel
func (ss *SSHServer) handleSessionInput(user string, clientsess *clientSSHSession) {
	defer clientsess.connection.Close()
	for {
		line, err := clientsess.terminal.ReadLine()
		if err != nil {
			if err.Error() != "EOF" {
				log.Println("Read error:", err)
			}
			ss.removeClientSession(clientsess.id, true)
			break
		}
		ss.activeClientsMutex.Lock()
		for _, sessions := range ss.activeClientsMap {

			for _, cs := range sessions {
				_, err := cs.terminal.Write([]byte(fmt.Sprintf("%s said: %q\n", user, line)))
				if err != nil {
					if err.Error() != "EOF" {
						log.Println("Write error:", err)
					}
					ss.removeClientSession(cs.id, false)
				}
			}

		}
		ss.activeClientsMutex.Unlock()
	}
}

// removes the client session based on the
func (ss *SSHServer) removeClientSession(sessionId string, lock bool) {
	if lock {
		ss.activeClientsMutex.Lock()
	}
	for user, sessions := range ss.activeClientsMap {
		var updatedSessions []clientSSHSession
		for _, session := range sessions {
			if session.id != sessionId {
				updatedSessions = append(updatedSessions, session)
			}
		}

		// Update the map with the filtered sessions
		ss.activeClientsMap[user] = updatedSessions
		log.Println("Removed Session:", sessionId)
		if len(ss.activeClientsMap[user]) == 0 {
			delete(ss.activeClientsMap, user)
			log.Println("Removed all channels for:", user)
		}
	}

	if lock {
		ss.activeClientsMutex.Unlock()
	}

}

// Handles ssh requests and replies to them to service the ssh connection
func (ss *SSHServer) handleSSHRequests(sshRequests <-chan *ssh.Request) {
	for req := range sshRequests {
		if req.Type == "pty-req" {
			termLen := req.Payload[3]
			term := string(req.Payload[4 : termLen+4])
			log.Printf("PTY requested: %s", term)
			if req.WantReply {
				req.Reply(true, nil)
			}
		}
		if req.Type == "shell" {
			req.Reply(true, nil)
		}
	}
}
