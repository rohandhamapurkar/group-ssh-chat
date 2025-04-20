package commands

import (
	"strings"
	"testing"
)

func TestUsersCommand(t *testing.T) {
	// Test users to return
	testUsers := []string{"user1", "user2", "user3"}
	
	// Track system messages
	var receivedMessage string
	systemMessageFunc := func(message string) {
		receivedMessage = message
	}
	
	// Create the users command
	usersCmd := NewUsersCommand(
		func() []string {
			return testUsers
		},
		systemMessageFunc,
	)
	
	// Test GetCommandHelp
	helpText := usersCmd.GetCommandHelp()
	if helpText != "Shows a list of online users" {
		t.Errorf("GetCommandHelp() = %s, want 'Shows a list of online users'", helpText)
	}
	
	// Test HandleCommand
	handled, err := usersCmd.HandleCommand("users", []string{}, "sender")
	
	if !handled {
		t.Error("HandleCommand() returned handled=false, want true")
	}
	
	if err != nil {
		t.Errorf("HandleCommand() returned err=%v, want nil", err)
	}
	
	// Check that the system message was sent with the user list
	if receivedMessage == "" {
		t.Error("HandleCommand() did not send a system message")
	}
	
	// Check that the message contains the expected user count
	if !strings.Contains(receivedMessage, "Online users (3)") {
		t.Errorf("System message does not contain correct user count, got: %s", receivedMessage)
	}
	
	// Check that the message contains all users
	for _, user := range testUsers {
		if !strings.Contains(receivedMessage, user) {
			t.Errorf("System message does not contain user %s, got: %s", user, receivedMessage)
		}
	}
	
	// Test with empty user list
	emptyUsersCmd := NewUsersCommand(
		func() []string {
			return []string{}
		},
		systemMessageFunc,
	)
	
	receivedMessage = ""
	handled, err = emptyUsersCmd.HandleCommand("users", []string{}, "sender")
	
	if !handled {
		t.Error("HandleCommand() with empty users returned handled=false, want true")
	}
	
	if err != nil {
		t.Errorf("HandleCommand() with empty users returned err=%v, want nil", err)
	}
	
	// Check that the system message indicates no users
	if !strings.Contains(receivedMessage, "No users are currently online") {
		t.Errorf("System message for empty users list incorrect, got: %s", receivedMessage)
	}
}
