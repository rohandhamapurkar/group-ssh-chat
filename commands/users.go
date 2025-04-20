package commands

import (
	"fmt"
	"strings"
)

// UsersCommand shows a list of online users
type UsersCommand struct {
	getUsersFunc func() []string
	systemMessageFunc func(string)
}

// NewUsersCommand creates a new users command
func NewUsersCommand(
	getUsersFunc func() []string,
	systemMessageFunc func(string),
) *UsersCommand {
	return &UsersCommand{
		getUsersFunc: getUsersFunc,
		systemMessageFunc: systemMessageFunc,
	}
}

// HandleCommand handles the users command
func (c *UsersCommand) HandleCommand(command string, args []string, sender string) (bool, error) {
	users := c.getUsersFunc()
	
	if len(users) == 0 {
		c.systemMessageFunc("No users are currently online")
		return true, nil
	}
	
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Online users (%d):\n", len(users)))
	
	for _, user := range users {
		sb.WriteString(fmt.Sprintf("- %s\n", user))
	}
	
	c.systemMessageFunc(sb.String())
	return true, nil
}

// GetCommandHelp returns help text for the users command
func (c *UsersCommand) GetCommandHelp() string {
	return "Shows a list of online users"
}
