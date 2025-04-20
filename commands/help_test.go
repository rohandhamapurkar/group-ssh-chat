package commands

import (
	"testing"
)

func TestHelpCommand(t *testing.T) {
	// Create a mock command manager
	cm := NewCommandManager()
	cm.RegisterCommand("test", &MockCommandHandler{helpText: "Test command"})
	
	// Track system messages
	var receivedMessage string
	systemMessageFunc := func(message string) {
		receivedMessage = message
	}
	
	// Create the help command
	helpCmd := NewHelpCommand(cm, systemMessageFunc)
	
	// Test GetCommandHelp
	helpText := helpCmd.GetCommandHelp()
	if helpText != "Shows this help message" {
		t.Errorf("GetCommandHelp() = %s, want 'Shows this help message'", helpText)
	}
	
	// Test HandleCommand
	handled, err := helpCmd.HandleCommand("help", []string{}, "sender")
	
	if !handled {
		t.Error("HandleCommand() returned handled=false, want true")
	}
	
	if err != nil {
		t.Errorf("HandleCommand() returned err=%v, want nil", err)
	}
	
	// Check that the system message was sent with the help text
	if receivedMessage == "" {
		t.Error("HandleCommand() did not send a system message")
	}
	
	// Check that the help text contains information about the test command
	if receivedMessage != cm.GetHelpText() {
		t.Errorf("System message = %s, want %s", receivedMessage, cm.GetHelpText())
	}
}
