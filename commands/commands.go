package commands

import (
	"fmt"
	"strings"
)

// CommandHandler defines the interface for handling commands
type CommandHandler interface {
	// HandleCommand processes a command and returns whether it was handled and any error
	HandleCommand(command string, args []string, sender string) (bool, error)
	
	// GetCommandHelp returns help text for the command
	GetCommandHelp() string
}

// CommandManager manages all available commands
type CommandManager struct {
	commands map[string]CommandHandler
}

// NewCommandManager creates a new command manager
func NewCommandManager() *CommandManager {
	return &CommandManager{
		commands: make(map[string]CommandHandler),
	}
}

// RegisterCommand registers a command handler
func (cm *CommandManager) RegisterCommand(name string, handler CommandHandler) {
	cm.commands[name] = handler
}

// HandleCommand processes a command string and routes it to the appropriate handler
func (cm *CommandManager) HandleCommand(input string, sender string) (bool, error) {
	// Check if the input is a command (starts with /)
	if !strings.HasPrefix(input, "/") {
		return false, nil
	}
	
	// Split the input into command and arguments
	parts := strings.SplitN(input[1:], " ", 2)
	command := parts[0]
	
	var args []string
	if len(parts) > 1 {
		args = strings.Split(parts[1], " ")
	}
	
	// Find the command handler
	handler, exists := cm.commands[command]
	if !exists {
		return true, fmt.Errorf("unknown command: %s", command)
	}
	
	// Handle the command
	return handler.HandleCommand(command, args, sender)
}

// GetHelpText returns help text for all registered commands
func (cm *CommandManager) GetHelpText() string {
	var sb strings.Builder
	
	sb.WriteString("Available commands:\n")
	
	for name, handler := range cm.commands {
		sb.WriteString(fmt.Sprintf("/%s - %s\n", name, handler.GetCommandHelp()))
	}
	
	return sb.String()
}
