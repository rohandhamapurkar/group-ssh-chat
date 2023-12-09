package main

import (
	"group-ssh-chat/auth"
	"group-ssh-chat/sshserver"
	"log"

	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()

	sshAuth := auth.New()
	sshServer := sshserver.New(sshAuth)
	
	log.Println("SSH server is listening for incoming connections.")
	sshServer.AcceptConnections()

}
