package commands

import (
	"testing"
)

func TestClearCommand(t *testing.T) {
	// Track if clear was called
	clearCalled := false
	clearFunc := func() {
		clearCalled = true
	}
	
	// Track system messages
	var systemMessage string
	systemMessageFunc := func(message string) {
		systemMessage = message
	}
	
	// Create the clear command
	clearCmd := NewClearCommand(
		clearFunc,
		systemMessageFunc,
	)
	
	// Test GetCommandHelp
	helpText := clearCmd.GetCommandHelp()
	if helpText != "Clears the chat screen" {
		t.Errorf("GetCommandHelp() = %s, want 'Clears the chat screen'", helpText)
	}
	
	// Test HandleCommand
	handled, err := clearCmd.HandleCommand("clear", []string{}, "sender")
	
	if !handled {
		t.Error("HandleCommand() returned handled=false, want true")
	}
	
	if err != nil {
		t.Errorf("HandleCommand() returned err=%v, want nil", err)
	}
	
	// Check that the clear function was called
	if !clearCalled {
		t.Error("HandleCommand() did not call the clear function")
	}
	
	// Check that a confirmation message was sent
	if systemMessage != "Screen cleared" {
		t.Errorf("System message = %s, want 'Screen cleared'", systemMessage)
	}
}
