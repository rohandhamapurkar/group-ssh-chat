package commands

import (
	"fmt"
	"strings"
)

// WhisperCommand allows sending private messages to other users
type WhisperCommand struct {
	whisperFunc func(string, string, string)
	userExistsFunc func(string) bool
	systemMessageFunc func(string)
}

// NewWhisperCommand creates a new whisper command
func NewWhisperCommand(
	whisperFunc func(string, string, string),
	userExistsFunc func(string) bool,
	systemMessageFunc func(string),
) *WhisperCommand {
	return &WhisperCommand{
		whisperFunc: whisperFunc,
		userExistsFunc: userExistsFunc,
		systemMessageFunc: systemMessageFunc,
	}
}

// HandleCommand handles the whisper command
func (c *WhisperCommand) HandleCommand(command string, args []string, sender string) (bool, error) {
	if len(args) < 2 {
		c.systemMessageFunc("Usage: /whisper <username> <message>")
		return true, nil
	}
	
	recipient := args[0]
	message := strings.Join(args[1:], " ")
	
	// Check if the recipient exists
	if !c.userExistsFunc(recipient) {
		c.systemMessageFunc(fmt.Sprintf("User '%s' not found", recipient))
		return true, nil
	}
	
	// Send the whisper
	c.whisperFunc(sender, recipient, message)
	return true, nil
}

// GetCommandHelp returns help text for the whisper command
func (c *WhisperCommand) GetCommandHelp() string {
	return "Sends a private message to another user"
}
