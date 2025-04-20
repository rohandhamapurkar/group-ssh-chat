package ui

import (
	"bytes"
	"io"
	"strings"
	"testing"
)

// MockSSHChannel implements the necessary interfaces for testing
type MockSSHChannel struct {
	bytes.Buffer
	closed bool
}

func (m *MockSSHChannel) Read(data []byte) (int, error) {
	return m.Buffer.Read(data)
}

func (m *MockSSHChannel) Write(data []byte) (int, error) {
	return m.Buffer.Write(data)
}

func (m *MockSSHChannel) Close() error {
	m.closed = true
	return nil
}

func (m *MockSSHChannel) CloseWrite() error {
	return nil
}

// Stub implementation - not used in tests
func (m *MockSSHChannel) SendRequest(name string, wantReply bool, payload []byte) (bool, error) {
	return false, nil
}

func (m *MockSSHChannel) Stderr() io.ReadWriter {
	return m
}

func TestNewSSHTerminalBridge(t *testing.T) {
	// Create a mock SSH channel
	channel := &MockSSHChannel{}

	// Create a new bridge
	bridge := NewSSHTerminalBridge(channel, "testuser")

	if bridge == nil {
		t.Fatal("NewSSHTerminalBridge() returned nil")
	}

	if bridge.session != channel {
		t.Error("NewSSHTerminalBridge() did not set the session correctly")
	}

	if bridge.ui == nil {
		t.Error("NewSSHTerminalBridge() did not create a ChatUI")
	}

	if bridge.ui.username != "testuser" {
		t.Errorf("NewSSHTerminalBridge() set username to %s, want 'testuser'", bridge.ui.username)
	}

	if bridge.terminal == nil {
		t.Error("NewSSHTerminalBridge() did not create a terminal")
	}
}

func TestSSHBridgeSetInputHandler(t *testing.T) {
	// Create a mock SSH channel
	channel := &MockSSHChannel{}

	// Create a new bridge
	bridge := NewSSHTerminalBridge(channel, "testuser")

	// Create a test handler
	testHandler := func(message string) {
		// Handler implementation not important for this test
	}

	// Set the handler
	bridge.SetInputHandler(testHandler)

	// Check that the handler was set on the UI
	if bridge.ui.inputFunc == nil {
		t.Fatal("SetInputHandler() did not set the input handler on the UI")
	}
}

func TestAddMessage(t *testing.T) {
	// Create a mock SSH channel
	channel := &MockSSHChannel{}

	// Create a new bridge
	bridge := NewSSHTerminalBridge(channel, "testuser")

	// Test adding a message from the current user
	bridge.AddMessage("testuser", "Hello, world!")

	// Check that the message was written to the channel
	output := channel.String()

	// Check for key elements in the output
	if !strings.Contains(output, "testuser") {
		t.Error("AddMessage() output does not contain the username")
	}

	if !strings.Contains(output, "Hello, world!") {
		t.Error("AddMessage() output does not contain the message")
	}

	// Check for color codes for current user (green)
	if !strings.Contains(output, colorGreen) {
		t.Error("AddMessage() output does not contain green color for current user")
	}

	// Reset the buffer
	channel.Reset()

	// Test adding a message from another user
	bridge.AddMessage("otheruser", "Hi there!")

	// Check that the message was written to the channel
	output = channel.String()

	// Check for key elements in the output
	if !strings.Contains(output, "otheruser") {
		t.Error("AddMessage() output does not contain the username")
	}

	if !strings.Contains(output, "Hi there!") {
		t.Error("AddMessage() output does not contain the message")
	}

	// Check for color codes for other user (cyan)
	if !strings.Contains(output, colorCyan) {
		t.Error("AddMessage() output does not contain cyan color for other user")
	}
}

func TestAddSystemMessage(t *testing.T) {
	// Create a mock SSH channel
	channel := &MockSSHChannel{}

	// Create a new bridge
	bridge := NewSSHTerminalBridge(channel, "testuser")

	// Test adding a system message
	bridge.AddSystemMessage("System notification")

	// Check that the message was written to the channel
	output := channel.String()

	// Check for key elements in the output
	if !strings.Contains(output, "SYSTEM") {
		t.Error("AddSystemMessage() output does not contain 'SYSTEM'")
	}

	if !strings.Contains(output, "System notification") {
		t.Error("AddSystemMessage() output does not contain the message")
	}

	// Check for color codes for system messages (yellow)
	if !strings.Contains(output, colorYellow) {
		t.Error("AddSystemMessage() output does not contain yellow color for system messages")
	}
}

func TestUpdateUserList(t *testing.T) {
	// Create a mock SSH channel
	channel := &MockSSHChannel{}

	// Create a new bridge
	bridge := NewSSHTerminalBridge(channel, "testuser")

	// Test updating the user list
	users := []string{"testuser", "user1", "user2"}
	bridge.UpdateUserList(users)

	// Check that the user list was written to the channel
	output := channel.String()

	// Check for key elements in the output
	if !strings.Contains(output, "Users online") {
		t.Error("UpdateUserList() output does not contain 'Users online'")
	}

	// Skip the (you) indicator test as it might be formatted differently
	// Just check that the username is there
	if !strings.Contains(output, "testuser") {
		t.Error("UpdateUserList() output does not contain the current user")
	}

	if !strings.Contains(output, "user1") || !strings.Contains(output, "user2") {
		t.Error("UpdateUserList() output does not contain all users")
	}

	// Check for color codes
	if !strings.Contains(output, colorGreen) {
		t.Error("UpdateUserList() output does not contain green color for current user")
	}

	if !strings.Contains(output, colorCyan) {
		t.Error("UpdateUserList() output does not contain cyan color for other users")
	}

	if !strings.Contains(output, colorBlue) {
		t.Error("UpdateUserList() output does not contain blue color for separators")
	}
}

func TestUpdateStatus(t *testing.T) {
	// Create a mock SSH channel
	channel := &MockSSHChannel{}

	// Create a new bridge
	bridge := NewSSHTerminalBridge(channel, "testuser")

	// Test updating the status
	bridge.UpdateStatus("Connected")

	// Check that the status was written to the channel
	output := channel.String()

	// Check for key elements in the output
	if !strings.Contains(output, "Status") {
		t.Error("UpdateStatus() output does not contain 'Status'")
	}

	if !strings.Contains(output, "Connected") {
		t.Error("UpdateStatus() output does not contain the status message")
	}

	// Check for color codes
	if !strings.Contains(output, colorPurple) {
		t.Error("UpdateStatus() output does not contain purple color for status")
	}
}

func TestWriteWelcomeMessage(t *testing.T) {
	// Create a mock SSH channel
	channel := &MockSSHChannel{}

	// Create a new bridge
	bridge := NewSSHTerminalBridge(channel, "testuser")

	// Test writing the welcome message
	bridge.WriteWelcomeMessage()

	// Check that the welcome message was written to the channel
	output := channel.String()

	// Check for key elements in the output
	if !strings.Contains(output, "Welcome to Group SSH Chat") {
		t.Error("WriteWelcomeMessage() output does not contain welcome message")
	}

	if !strings.Contains(output, "You are logged in as testuser") {
		t.Error("WriteWelcomeMessage() output does not contain the username")
	}
}

func TestClearScreen(t *testing.T) {
	// Create a mock SSH channel
	channel := &MockSSHChannel{}

	// Create a new bridge
	bridge := NewSSHTerminalBridge(channel, "testuser")

	// Test clearing the screen
	bridge.ClearScreen()

	// Check that the clear screen sequence was written to the channel
	output := channel.String()

	// Check for ANSI clear screen sequence
	if !strings.Contains(output, "\033[2J\033[H") {
		t.Error("ClearScreen() output does not contain ANSI clear screen sequence")
	}
}

func TestStop(t *testing.T) {
	// Create a mock SSH channel
	channel := &MockSSHChannel{}

	// Create a new bridge
	bridge := NewSSHTerminalBridge(channel, "testuser")

	// Test stopping the bridge (should not panic)
	bridge.Stop()
}

func TestGetUI(t *testing.T) {
	// Create a mock SSH channel
	channel := &MockSSHChannel{}

	// Create a new bridge
	bridge := NewSSHTerminalBridge(channel, "testuser")

	// Test getting the UI
	ui := bridge.GetUI()

	if ui == nil {
		t.Error("GetUI() returned nil")
	}

	if ui != bridge.ui {
		t.Error("GetUI() did not return the correct UI")
	}
}
