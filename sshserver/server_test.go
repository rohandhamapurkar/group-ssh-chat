package sshserver

import (
	"crypto/rand"
	"crypto/rsa"
	"group-ssh-chat/auth"
	"os"
	"testing"

	"golang.org/x/crypto/ssh"
)

// Setup test environment
func setupTestEnv(t *testing.T) *SSHServer {
	// Create a temporary private key for testing
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("Failed to generate test key: %v", err)
	}
	
	// Convert to SSH signer
	signer, err := ssh.NewSignerFromKey(privateKey)
	if err != nil {
		t.Fatalf("Failed to create signer: %v", err)
	}

	// Create a real auth instance
	realAuth := &auth.SSHAuth{
		HostSSHPrivateKey: signer,
	}

	// Set environment variables for testing
	os.Setenv("SSH_SERVER_HOST", "127.0.0.1")
	os.Setenv("SSH_SERVER_PORT", "0") // Use port 0 to let the OS assign a free port

	// Create a new server
	server := New(realAuth)

	return server
}

func TestNew(t *testing.T) {
	server := setupTestEnv(t)

	if server == nil {
		t.Fatal("New() returned nil")
	}

	if server.activeClientsMap == nil {
		t.Error("New() did not initialize activeClientsMap")
	}

	if server.sshServerConfig == nil {
		t.Error("New() did not initialize sshServerConfig")
	}

	if server.tcpListener == nil {
		t.Error("New() did not initialize tcpListener")
	}

	if server.commandManager == nil {
		t.Error("New() did not initialize commandManager")
	}
}

func TestRegisterCommands(t *testing.T) {
	server := setupTestEnv(t)

	// Check that commands were registered
	if len(server.commandManager.GetHelpText()) == 0 {
		t.Error("registerCommands() did not register any commands")
	}
}
