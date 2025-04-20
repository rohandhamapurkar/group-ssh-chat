package ui

import (
	"testing"
)

func TestNewChatUI(t *testing.T) {
	// Test creating a new ChatUI
	ui := NewChatUI("testuser")
	
	if ui == nil {
		t.Fatal("NewChatUI() returned nil")
	}
	
	if ui.username != "testuser" {
		t.Errorf("NewChatUI() set username to %s, want 'testuser'", ui.username)
	}
	
	if ui.inputFunc != nil {
		t.Error("NewChatUI() should initialize inputFunc to nil")
	}
}

func TestSetInputHandler(t *testing.T) {
	// Create a test UI
	ui := NewChatUI("testuser")
	
	// Create a test handler
	handlerCalled := false
	testHandler := func(message string) {
		handlerCalled = true
	}
	
	// Set the handler
	ui.SetInputHandler(testHandler)
	
	// Check that the handler was set
	if ui.inputFunc == nil {
		t.Fatal("SetInputHandler() did not set the input handler")
	}
	
	// Call the handler and check that it works
	ui.inputFunc("test message")
	
	if !handlerCalled {
		t.Error("Input handler was not called when invoked")
	}
}

func TestPlaceholderMethods(t *testing.T) {
	// Create a test UI
	ui := NewChatUI("testuser")
	
	// Test Run method
	err := ui.Run()
	if err != nil {
		t.Errorf("Run() returned error: %v", err)
	}
	
	// Test Stop method (should not panic)
	ui.Stop()
	
	// Test GetApplication method
	app := ui.GetApplication()
	if app != nil {
		t.Errorf("GetApplication() returned %v, want nil", app)
	}
}
