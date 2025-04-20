package commands

import (
	"errors"
	"strings"
	"testing"
)

// MockCommandHandler implements CommandHandler for testing
type MockCommandHandler struct {
	helpText      string
	shouldHandle  bool
	shouldError   bool
	handledCalled bool
	lastCommand   string
	lastArgs      []string
	lastSender    string
}

func (m *MockCommandHandler) HandleCommand(command string, args []string, sender string) (bool, error) {
	m.handledCalled = true
	m.lastCommand = command
	m.lastArgs = args
	m.lastSender = sender
	
	if m.shouldError {
		return m.shouldHandle, errors.New("mock error")
	}
	return m.shouldHandle, nil
}

func (m *MockCommandHandler) GetCommandHelp() string {
	return m.helpText
}

func TestNewCommandManager(t *testing.T) {
	cm := NewCommandManager()
	if cm == nil {
		t.Fatal("NewCommandManager() returned nil")
	}
	
	if cm.commands == nil {
		t.Fatal("NewCommandManager() did not initialize commands map")
	}
}

func TestRegisterCommand(t *testing.T) {
	cm := NewCommandManager()
	handler := &MockCommandHandler{helpText: "Test command"}
	
	cm.RegisterCommand("test", handler)
	
	if _, exists := cm.commands["test"]; !exists {
		t.Fatal("RegisterCommand() did not register the command")
	}
}

func TestHandleCommand(t *testing.T) {
	cm := NewCommandManager()
	
	// Test non-command input
	handled, err := cm.HandleCommand("not a command", "sender")
	if handled || err != nil {
		t.Errorf("HandleCommand() with non-command input returned handled=%v, err=%v, want handled=false, err=nil", handled, err)
	}
	
	// Test unknown command
	handled, err = cm.HandleCommand("/unknown", "sender")
	if !handled {
		t.Error("HandleCommand() with unknown command returned handled=false, want true")
	}
	if err == nil || !strings.Contains(err.Error(), "unknown command") {
		t.Errorf("HandleCommand() with unknown command returned err=%v, want error containing 'unknown command'", err)
	}
	
	// Test successful command handling
	handler := &MockCommandHandler{shouldHandle: true, helpText: "Test command"}
	cm.RegisterCommand("test", handler)
	
	handled, err = cm.HandleCommand("/test arg1 arg2", "sender")
	if !handled {
		t.Error("HandleCommand() with valid command returned handled=false, want true")
	}
	if err != nil {
		t.Errorf("HandleCommand() with valid command returned err=%v, want nil", err)
	}
	if !handler.handledCalled {
		t.Error("HandleCommand() did not call handler's HandleCommand method")
	}
	if handler.lastCommand != "test" {
		t.Errorf("Handler received command=%s, want 'test'", handler.lastCommand)
	}
	if len(handler.lastArgs) != 2 || handler.lastArgs[0] != "arg1" || handler.lastArgs[1] != "arg2" {
		t.Errorf("Handler received args=%v, want ['arg1', 'arg2']", handler.lastArgs)
	}
	if handler.lastSender != "sender" {
		t.Errorf("Handler received sender=%s, want 'sender'", handler.lastSender)
	}
	
	// Test command with error
	errorHandler := &MockCommandHandler{shouldHandle: true, shouldError: true, helpText: "Error command"}
	cm.RegisterCommand("error", errorHandler)
	
	handled, err = cm.HandleCommand("/error", "sender")
	if !handled {
		t.Error("HandleCommand() with error command returned handled=false, want true")
	}
	if err == nil {
		t.Error("HandleCommand() with error command returned err=nil, want error")
	}
}

func TestGetHelpText(t *testing.T) {
	cm := NewCommandManager()
	
	// Register some test commands
	cm.RegisterCommand("test1", &MockCommandHandler{helpText: "Test command 1"})
	cm.RegisterCommand("test2", &MockCommandHandler{helpText: "Test command 2"})
	
	helpText := cm.GetHelpText()
	
	// Check that the help text contains the expected information
	if !strings.Contains(helpText, "Available commands:") {
		t.Error("GetHelpText() output does not contain 'Available commands:'")
	}
	if !strings.Contains(helpText, "/test1 - Test command 1") {
		t.Error("GetHelpText() output does not contain help for test1")
	}
	if !strings.Contains(helpText, "/test2 - Test command 2") {
		t.Error("GetHelpText() output does not contain help for test2")
	}
}
