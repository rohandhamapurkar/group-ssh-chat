package commands

import (
	"strings"
	"testing"
)

func TestWhisperCommand(t *testing.T) {
	// Track whisper messages
	var lastSender, lastRecipient, lastMessage string
	whisperFunc := func(sender, recipient, message string) {
		lastSender = sender
		lastRecipient = recipient
		lastMessage = message
	}
	
	// Track system messages
	var systemMessage string
	systemMessageFunc := func(message string) {
		systemMessage = message
	}
	
	// User existence check
	userExistsFunc := func(username string) bool {
		return username == "validuser"
	}
	
	// Create the whisper command
	whisperCmd := NewWhisperCommand(
		whisperFunc,
		userExistsFunc,
		systemMessageFunc,
	)
	
	// Test GetCommandHelp
	helpText := whisperCmd.GetCommandHelp()
	if helpText != "Sends a private message to another user" {
		t.Errorf("GetCommandHelp() = %s, want 'Sends a private message to another user'", helpText)
	}
	
	// Test HandleCommand with insufficient arguments
	handled, err := whisperCmd.HandleCommand("whisper", []string{}, "sender")
	
	if !handled {
		t.Error("HandleCommand() with insufficient args returned handled=false, want true")
	}
	
	if err != nil {
		t.Errorf("HandleCommand() with insufficient args returned err=%v, want nil", err)
	}
	
	// Check that a usage message was sent
	if !strings.Contains(systemMessage, "Usage:") {
		t.Errorf("System message for insufficient args incorrect, got: %s", systemMessage)
	}
	
	// Test HandleCommand with non-existent user
	systemMessage = ""
	handled, err = whisperCmd.HandleCommand("whisper", []string{"invaliduser", "test message"}, "sender")
	
	if !handled {
		t.Error("HandleCommand() with invalid user returned handled=false, want true")
	}
	
	if err != nil {
		t.Errorf("HandleCommand() with invalid user returned err=%v, want nil", err)
	}
	
	// Check that an error message was sent
	if !strings.Contains(systemMessage, "not found") {
		t.Errorf("System message for invalid user incorrect, got: %s", systemMessage)
	}
	
	// Test HandleCommand with valid user and message
	systemMessage = ""
	lastSender = ""
	lastRecipient = ""
	lastMessage = ""
	
	handled, err = whisperCmd.HandleCommand("whisper", []string{"validuser", "Hello", "there!"}, "sender")
	
	if !handled {
		t.Error("HandleCommand() with valid args returned handled=false, want true")
	}
	
	if err != nil {
		t.Errorf("HandleCommand() with valid args returned err=%v, want nil", err)
	}
	
	// Check that the whisper function was called with correct parameters
	if lastSender != "sender" {
		t.Errorf("Whisper sender = %s, want 'sender'", lastSender)
	}
	
	if lastRecipient != "validuser" {
		t.Errorf("Whisper recipient = %s, want 'validuser'", lastRecipient)
	}
	
	if lastMessage != "Hello there!" {
		t.Errorf("Whisper message = %s, want 'Hello there!'", lastMessage)
	}
}
