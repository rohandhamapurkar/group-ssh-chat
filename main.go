package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"sync"

	"golang.org/x/crypto/ssh"
	"golang.org/x/term"
)

var authorizedKeysMap map[string]bool

var sshServerConfig *ssh.ServerConfig = &ssh.ServerConfig{
	// Remove to disable password auth.
	PasswordCallback: handlePasswordLogin,
	// Remove to disable public key auth.
	PublicKeyCallback: handlePublicKeyLogin,
}

func main() {
	// Public key authentication is done by comparing
	// the public key of a received connection
	// with the entries in the authorized_keys file.
	// authorizedKeysBytes, err := os.ReadFile("authorized_keys")
	// if err != nil {
	// 	log.Fatalf("Failed to load authorized_keys, err: %v", err)
	// }

	// authorizedKeysMap := map[string]bool{}
	// for len(authorizedKeysBytes) > 0 {
	// 	pubKey, _, _, rest, err := ssh.ParseAuthorizedKey(authorizedKeysBytes)
	// 	if err != nil {
	// 		log.Fatal(err)
	// 	}

	// 	authorizedKeysMap[string(pubKey.Marshal())] = true
	// 	authorizedKeysBytes = rest
	// }

	// An SSH server is represented by a ServerConfig, which holds
	// certificate details and handles authentication of ServerConns.
	sshServerConfig.AddHostKey(getSvrPrivateKey())

	log.Println("SSH server is listening for incoming connections.")

	listenForNewConnections()

	
}

func getSvrPrivateKey() ssh.Signer {
	privateBytes, err := os.ReadFile("ssh_server_key")
	if err != nil {
		log.Fatal("Failed to load private key: ", err)
	}

	private, err := ssh.ParsePrivateKey(privateBytes)
	if err != nil {
		log.Fatal("Failed to parse private key: ", err)
	}

	return private
}

func handlePublicKeyLogin(c ssh.ConnMetadata, pubKey ssh.PublicKey) (*ssh.Permissions, error) {
	if authorizedKeysMap[string(pubKey.Marshal())] {
		return &ssh.Permissions{
			// Record the public key used for authentication.
			Extensions: map[string]string{
				"pubkey-fp": ssh.FingerprintSHA256(pubKey),
			},
		}, nil
	}
	return nil, fmt.Errorf("unknown public key for %q", c.User())
}

func handlePasswordLogin(c ssh.ConnMetadata, pass []byte) (*ssh.Permissions, error) {
	// Should use constant-time compare (or better, salt+hash) in
	// a production setting.
	if c.User() == "testuser" && string(pass) == "tiger" {
		return &ssh.Permissions{
			// Record the public key used for authentication.
			Extensions: map[string]string{
				"pubkey-fp": c.User(),
			},
		}, nil
	}
	return nil, fmt.Errorf("password rejected for %q", c.User())
}

func listenForNewConnections() {
	// Once a ServerConfig has been configured, connections can be
	// accepted.
	listener, err := net.Listen("tcp", "0.0.0.0:2022")
	if err != nil {
		log.Fatal("failed to listen for connection: ", err)
	}
	for {
		nConn, err := listener.Accept()
		if err != nil {
			log.Printf("failed to accept incoming connection: %q", err)
			continue
		}

		// Before use, a handshake must be performed on the incoming
		// net.Conn.
		conn, chans, reqs, err := ssh.NewServerConn(nConn, sshServerConfig)
		if err != nil {
			log.Printf("failed to handshake: %q", err)
			continue
		}
		log.Printf("logged in with key %s", conn.Permissions.Extensions["pubkey-fp"])
		go handleSingleConn(conn, chans, reqs)

	}
}

func handleSingleConn(conn *ssh.ServerConn, chans <-chan ssh.NewChannel, reqs <-chan *ssh.Request) {
	var wg sync.WaitGroup
	defer wg.Wait()

	// The incoming Request channel must be serviced.
	wg.Add(1)
	go func() {
		ssh.DiscardRequests(reqs)
		wg.Done()
	}()

	// Service the incoming Channel channel.
	for newChannel := range chans {
		// Channels have a type, depending on the application level
		// protocol intended. In the case of a shell, the type is
		// "session" and ServerShell may be used to present a simple
		// terminal interface.
		if newChannel.ChannelType() != "session" {
			newChannel.Reject(ssh.UnknownChannelType, "unknown channel type")
			continue
		}
		channel, requests, err := newChannel.Accept()
		if err != nil {
			log.Fatalf("Could not accept channel: %v", err)
		}

		// Sessions have out-of-band requests such as "shell",
		// "pty-req" and "env".  Here we handle only the
		// "shell" request.
		wg.Add(1)
		go func(inChan <-chan *ssh.Request) {
			for req := range inChan {
				if req.Type == "pty-req" {
					termLen := req.Payload[3]
					term := string(req.Payload[4 : termLen+4])
					log.Printf("PTY requested: %s", term)
					if req.WantReply {
						req.Reply(true, nil)
					}
					return
				}
				if req.Type == "shell" {
					req.Reply(true, nil)
					return
				}
			}
			wg.Done()
		}(requests)

		term := term.NewTerminal(channel, "> ")

		wg.Add(1)
		go func() {
			defer func() {
				channel.Close()
				wg.Done()
			}()
			for {
				line, err := term.ReadLine()
				if err != nil {
					break
				}
				fmt.Println("Out:" + line)
			}
		}()
	}
}
