package main

import (
	"fmt"
	"group-ssh-chat/auth"
	"group-ssh-chat/sshserver"
	"log"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()

	sshAuth := auth.New()
	sshServer := sshserver.New(sshAuth)

	host := os.Getenv("SSH_SERVER_HOST")
	port := os.Getenv("SSH_SERVER_PORT")

	fmt.Println("╔════════════════════════════════════════════════════════════╗")
	fmt.Println("║                   Group SSH Chat Server                    ║")
	fmt.Println("╠════════════════════════════════════════════════════════════╣")
	fmt.Println("║ Features:                                                  ║")
	fmt.Println("║  • Enhanced terminal UI with tview                         ║")
	fmt.Println("║  • Split screen with chat history and user list            ║")
	fmt.Println("║  • Color-coded usernames and timestamps                    ║")
	fmt.Println("║  • Real-time user list updates                             ║")
	fmt.Println("╠════════════════════════════════════════════════════════════╣")
	fmt.Printf("║ Server listening on %s:%s                          ║\n", host, port)
	fmt.Println("╚════════════════════════════════════════════════════════════╝")

	log.Printf("SSH server is listening for connections on %s:%s", host, port)
	sshServer.AcceptConnections()
}
