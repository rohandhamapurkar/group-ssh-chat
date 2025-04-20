package ui

// ChatUI manages the terminal UI components
type ChatUI struct {
	inputFunc func(string)
	username  string
}

// NewChatUI creates a new chat UI instance
func NewChatUI(username string) *ChatUI {
	ui := &ChatUI{
		username: username,
	}
	return ui
}

// SetInputHandler sets the function to be called when the user submits a message
func (ui *ChatUI) SetInputHandler(handler func(string)) {
	ui.inputFunc = handler
}

// Run is a placeholder for compatibility
func (ui *ChatUI) Run() error {
	// This is a no-op in the simple terminal mode
	return nil
}

// Stop is a placeholder for compatibility
func (ui *ChatUI) Stop() {
	// This is a no-op in the simple terminal mode
}

// GetApplication is a placeholder for compatibility
func (ui *ChatUI) GetApplication() interface{} {
	return nil
}
