package commands

// ClearCommand clears the chat screen
type ClearCommand struct {
	clearFunc func()
	systemMessageFunc func(string)
}

// NewClearCommand creates a new clear command
func NewClearCommand(
	clearFunc func(),
	systemMessageFunc func(string),
) *ClearCommand {
	return &ClearCommand{
		clearFunc: clearFunc,
		systemMessageFunc: systemMessageFunc,
	}
}

// HandleCommand handles the clear command
func (c *ClearCommand) HandleCommand(command string, args []string, sender string) (bool, error) {
	c.clearFunc()
	c.systemMessageFunc("Screen cleared")
	return true, nil
}

// GetCommandHelp returns help text for the clear command
func (c *ClearCommand) GetCommandHelp() string {
	return "Clears the chat screen"
}
