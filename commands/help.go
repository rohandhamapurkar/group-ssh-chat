package commands

// HelpCommand provides help information about available commands
type HelpCommand struct {
	manager *CommandManager
	systemMessageFunc func(string)
}

// NewHelpCommand creates a new help command
func NewHelpCommand(manager *CommandManager, systemMessageFunc func(string)) *HelpCommand {
	return &HelpCommand{
		manager: manager,
		systemMessageFunc: systemMessageFunc,
	}
}

// HandleCommand handles the help command
func (c *HelpCommand) HandleCommand(command string, args []string, sender string) (bool, error) {
	helpText := c.manager.GetHelpText()
	c.systemMessageFunc(helpText)
	return true, nil
}

// GetCommandHelp returns help text for the help command
func (c *HelpCommand) GetCommandHelp() string {
	return "Shows this help message"
}
