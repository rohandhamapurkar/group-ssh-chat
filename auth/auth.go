package auth

import (
	"fmt"
	"log"
	"os"

	"golang.org/x/crypto/ssh"
)

// Used for managing SSH authentication
type SSHAuth struct {
	authorizedKeysMap map[string]string
	HostSSHPrivateKey ssh.Signer
}

// Returns new ssh auth manager struct reference
func New() *SSHAuth {
	sam := &SSHAuth{
		authorizedKeysMap: map[string]string{},
	}
	sam.initHostSSHPrivateKey()
	sam.initAuthorizedKeys()

	return sam
}

// Handles the public authorized key login for a user
func (sam SSHAuth) HandlePublicKeyLogin(c ssh.ConnMetadata, pubKey ssh.PublicKey) (*ssh.Permissions, error) {
	// Check if the key is in our authorized keys map for any user
	pubKeyStr := string(pubKey.Marshal())
	for _, authorizedKey := range sam.authorizedKeysMap {
		if authorizedKey == pubKeyStr {
			return &ssh.Permissions{
				// Record the public key used for authentication.
				Extensions: map[string]string{
					"pubkey-fp": ssh.FingerprintSHA256(pubKey),
				},
			}, nil
		}
	}
	return nil, fmt.Errorf("unknown public key for %q", c.User())
}

// handles password based login
// func (sam SSHAuth) HandlePasswordLogin(c ssh.ConnMetadata, pass []byte) (*ssh.Permissions, error) {
// 	// Should use constant-time compare (or better, salt+hash) in
// 	// a production setting.
// 	if c.User() == "testuser" && string(pass) == "tiger" {
// 		return &ssh.Permissions{
// 			// Record the public key used for authentication.
// 			Extensions: map[string]string{
// 				"pubkey-fp": c.User(),
// 			},
// 		}, nil
// 	}
// 	return nil, fmt.Errorf("password rejected for %q", c.User())
// }

// Reads the host ssh server private key and parses it
func (sam *SSHAuth) initHostSSHPrivateKey() {
	pkBytes, err := os.ReadFile(os.Getenv("HOST_SSH_PRIVATE_KEY_PATH"))
	if err != nil {
		log.Fatal("Failed to load private key: ", err)
	}

	pk, err := ssh.ParsePrivateKey(pkBytes)
	if err != nil {
		log.Fatal("Failed to parse private key: ", err)
	}

	sam.HostSSHPrivateKey = pk
}

// Public key authentication is done by comparing the public key of a received connection
func (sam *SSHAuth) initAuthorizedKeys() {
	authorizedKeysBytes, err := os.ReadFile(os.Getenv("AUTHORIZED_KEYS_PATH"))
	if err != nil {
		log.Fatalf("Failed to load authorized_keys, err: %v", err)
	}

	for len(authorizedKeysBytes) > 0 {
		pubKey, comment, _, rest, err := ssh.ParseAuthorizedKey(authorizedKeysBytes)
		if err != nil {
			log.Fatal(err)
		}

		sam.authorizedKeysMap[comment] = string(pubKey.Marshal())
		authorizedKeysBytes = rest
	}
}
