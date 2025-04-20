package auth

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"net"
	"os"
	"testing"

	"golang.org/x/crypto/ssh"
)

// MockConnMetadata implements ssh.ConnMetadata for testing
type MockConnMetadata struct {
	username string
}

func (m *MockConnMetadata) User() string {
	return m.username
}

func (m *MockConnMetadata) SessionID() []byte {
	return []byte("mock-session-id")
}

func (m *MockConnMetadata) ClientVersion() []byte {
	return []byte("SSH-2.0-MockClient")
}

func (m *MockConnMetadata) ServerVersion() []byte {
	return []byte("SSH-2.0-MockServer")
}

func (m *MockConnMetadata) RemoteAddr() net.Addr {
	return &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 1234}
}

func (m *MockConnMetadata) LocalAddr() net.Addr {
	return &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 22}
}

func TestHandlePublicKeyLogin(t *testing.T) {
	// Create a temporary authorized_keys file for testing
	tempAuthKeysFile, err := os.CreateTemp("", "auth_test_keys")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tempAuthKeysFile.Name())

	// Create a temporary private key file for testing
	tempPrivKeyFile, err := os.CreateTemp("", "auth_test_privkey")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tempPrivKeyFile.Name())

	// Generate a test key
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("Failed to generate test key: %v", err)
	}

	// Convert to SSH private key
	signer, err := ssh.NewSignerFromKey(privateKey)
	if err != nil {
		t.Fatalf("Failed to create signer: %v", err)
	}

	// Write the private key to the temp file in PEM format
	privKeyDer := x509.MarshalPKCS1PrivateKey(privateKey)
	privKeyBlock := pem.Block{
		Type:    "RSA PRIVATE KEY",
		Headers: nil,
		Bytes:   privKeyDer,
	}
	privKeyPem := pem.EncodeToMemory(&privKeyBlock)
	err = os.WriteFile(tempPrivKeyFile.Name(), privKeyPem, 0600)
	if err != nil {
		t.Fatalf("Failed to write private key: %v", err)
	}

	// Create authorized_keys content with the public key
	publicKey := signer.PublicKey()
	authorizedKey := ssh.MarshalAuthorizedKey(publicKey)
	// Add a comment (username) to the authorized key
	authorizedKeyWithComment := append(authorizedKey[:len(authorizedKey)-1], []byte(" testuser\n")...)

	err = os.WriteFile(tempAuthKeysFile.Name(), authorizedKeyWithComment, 0600)
	if err != nil {
		t.Fatalf("Failed to write authorized keys: %v", err)
	}

	// Set environment variables for testing
	os.Setenv("HOST_SSH_PRIVATE_KEY_PATH", tempPrivKeyFile.Name())
	os.Setenv("AUTHORIZED_KEYS_PATH", tempAuthKeysFile.Name())

	// Create a new SSHAuth instance
	auth := New()

	// Test cases
	tests := []struct {
		name     string
		username string
		pubKey   ssh.PublicKey
		wantErr  bool
	}{
		{
			name:     "Valid key",
			username: "testuser",
			pubKey:   publicKey,
			wantErr:  false,
		},
		{
			name:     "Invalid key",
			username: "testuser",
			pubKey:   nil, // This will be replaced with a different key
			wantErr:  true,
		},
	}

	// Generate a different key for the invalid test case
	invalidPrivateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("Failed to generate invalid test key: %v", err)
	}

	// Convert to SSH key
	invalidSigner, err := ssh.NewSignerFromKey(invalidPrivateKey)
	if err != nil {
		t.Fatalf("Failed to create invalid signer: %v", err)
	}

	invalidKey := invalidSigner

	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Use the invalid key for the second test case
			if i == 1 {
				tt.pubKey = invalidKey.PublicKey()
			}

			conn := &MockConnMetadata{username: tt.username}
			perms, err := auth.HandlePublicKeyLogin(conn, tt.pubKey)

			if (err != nil) != tt.wantErr {
				t.Errorf("HandlePublicKeyLogin() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && perms == nil {
				t.Errorf("HandlePublicKeyLogin() returned nil permissions for valid key")
			}

			if !tt.wantErr && perms.Extensions["pubkey-fp"] == "" {
				t.Errorf("HandlePublicKeyLogin() returned empty fingerprint")
			}
		})
	}
}
