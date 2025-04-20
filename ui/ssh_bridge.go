package ui

import (
	"fmt"
	"io"
	"log"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"
	"golang.org/x/term"
)

// ANSI color codes
const (
	// Text colors
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorPurple = "\033[35m"
	colorCyan   = "\033[36m"
	colorGray   = "\033[90m"

	// Text styles
	colorBold      = "\033[1m"
	colorUnderline = "\033[4m"
)

// SSHTerminalBridge bridges between SSH terminal and chat functionality
type SSHTerminalBridge struct {
	session  ssh.Channel
	ui       *ChatUI
	terminal *term.Terminal
}

// NewSSHTerminalBridge creates a new SSH terminal bridge
func NewSSHTerminalBridge(session ssh.Channel, username string) *SSHTerminalBridge {
	ui := NewChatUI(username)

	// Create a terminal with a colored prompt
	terminal := term.NewTerminal(session, colorGreen + "> " + colorReset)

	bridge := &SSHTerminalBridge{
		session:  session,
		ui:       ui,
		terminal: terminal,
	}

	return bridge
}

// Start starts the terminal bridge
func (bridge *SSHTerminalBridge) Start() error {
	// Main input loop - this will block until the session ends
	for {
		line, err := bridge.terminal.ReadLine()
		if err != nil {
			if err != io.EOF {
				log.Printf("Error reading from SSH for user %s: %v", bridge.ui.username, err)
			}
			break
		}

		// Process the input
		if bridge.ui.inputFunc != nil && line != "" {
			bridge.ui.inputFunc(line)
		}
	}

	return nil
}

// SetInputHandler sets the function to be called when the user submits a message
func (bridge *SSHTerminalBridge) SetInputHandler(handler func(string)) {
	bridge.ui.SetInputHandler(handler)
}

// AddMessage adds a message to the chat view
func (bridge *SSHTerminalBridge) AddMessage(username, message string) {
	timestamp := time.Now().Format("15:04:05")

	// Format the message with timestamp and colors
	var formattedMsg string
	if username == bridge.ui.username {
		// Current user's messages in green
		formattedMsg = fmt.Sprintf("\r\n%s[%s]%s %s%s%s: %s\r\n",
			colorGray, timestamp, colorReset,
			colorGreen + colorBold, username, colorReset,
			message)
	} else {
		// Other users' messages in cyan
		formattedMsg = fmt.Sprintf("\r\n%s[%s]%s %s%s%s: %s\r\n",
			colorGray, timestamp, colorReset,
			colorCyan + colorBold, username, colorReset,
			message)
	}

	bridge.session.Write([]byte(formattedMsg))

	// Write the prompt again to ensure it appears after the message
	bridge.terminal.SetPrompt(colorGreen + "> " + colorReset)
}

// AddSystemMessage adds a system message to the chat view
func (bridge *SSHTerminalBridge) AddSystemMessage(message string) {
	timestamp := time.Now().Format("15:04:05")

	// Format the system message with timestamp and colors
	formattedMsg := fmt.Sprintf("\r\n%s[%s]%s %s%sSYSTEM:%s %s\r\n",
		colorGray, timestamp, colorReset,
		colorYellow, colorBold, colorReset,
		message)

	bridge.session.Write([]byte(formattedMsg))

	// Write the prompt again to ensure it appears after the message
	bridge.terminal.SetPrompt(colorGreen + "> " + colorReset)
}

// UpdateUserList updates the list of online users
func (bridge *SSHTerminalBridge) UpdateUserList(users []string) {
	// Format the user list with a header and separator
	var sb strings.Builder

	sb.WriteString("\r\n")
	sb.WriteString(colorBlue + "----------------------------------------" + colorReset + "\r\n")
	sb.WriteString(colorBlue + colorBold + "Users online:" + colorReset + "\r\n")
	sb.WriteString(colorBlue + "----------------------------------------" + colorReset + "\r\n")

	// Add current user at the top with green color
	sb.WriteString(fmt.Sprintf("  %s%s%s (you)\r\n",
		colorGreen + colorBold, bridge.ui.username, colorReset))

	// Add other users with cyan color
	for _, user := range users {
		if user != bridge.ui.username {
			sb.WriteString(fmt.Sprintf("  %s%s%s\r\n",
				colorCyan, user, colorReset))
		}
	}

	sb.WriteString(colorBlue + "----------------------------------------" + colorReset + "\r\n")
	sb.WriteString("\r\n")

	// Send the user list
	bridge.session.Write([]byte(sb.String()))

	// Write the prompt again to ensure it appears after the user list
	bridge.terminal.SetPrompt(colorGreen + "> " + colorReset)
}

// UpdateStatus updates the status bar text
func (bridge *SSHTerminalBridge) UpdateStatus(status string) {
	// Send the status message with proper formatting and colors
	statusMsg := fmt.Sprintf("\r\n%s----------------------------------------%s\r\n%s%sStatus:%s %s\r\n%s----------------------------------------%s\r\n\r\n",
		colorPurple, colorReset,
		colorPurple, colorBold, colorReset, status,
		colorPurple, colorReset)

	bridge.session.Write([]byte(statusMsg))

	// Write the prompt again to ensure it appears after the status message
	bridge.terminal.SetPrompt(colorGreen + "> " + colorReset)
}

// Stop stops the UI application
func (bridge *SSHTerminalBridge) Stop() {
	bridge.ui.Stop()
}

// GetUI returns the underlying ChatUI
func (bridge *SSHTerminalBridge) GetUI() *ChatUI {
	return bridge.ui
}

// WriteWelcomeMessage writes the welcome message to the terminal
func (bridge *SSHTerminalBridge) WriteWelcomeMessage() {
	// Create a welcome message with proper line endings
	var sb strings.Builder

	sb.WriteString("\r\n")
	sb.WriteString("=== Welcome to Group SSH Chat! ===\r\n")
	sb.WriteString("\r\n")
	sb.WriteString(fmt.Sprintf("You are logged in as %s\r\n", bridge.ui.username))
	sb.WriteString("\r\n")

	// Send the welcome message
	bridge.session.Write([]byte(sb.String()))
}

// ClearScreen clears the terminal screen
func (bridge *SSHTerminalBridge) ClearScreen() {
	// ANSI escape sequence to clear screen and move cursor to top-left
	bridge.session.Write([]byte("\033[2J\033[H"))

	// Write the prompt again
	bridge.terminal.SetPrompt(colorGreen + "> " + colorReset)
}
