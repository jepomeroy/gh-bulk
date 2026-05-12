// Package execute provides types for capturing and running shell commands against repositories.
package execute

import (
	"fmt"
	"os/exec"

	"github.com/charmbracelet/huh"
)

// Command holds the shell command string to run on each repository.
type Command struct {
	CommandValue string
}

// GetCommand prompts the user to enter the shell command to execute on each repository.
func GetCommand() (Command, error) {
	var command string

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Command").
				Prompt("Enter command: ").
				Description("Enter the command to run on each repository").
				Value(&command),
		),
	).WithTheme(huh.ThemeCatppuccin())

	err := form.Run()
	if err != nil {
		return Command{}, err
	}

	return Command{CommandValue: command}, nil
}

// Execute runs the command in a shell and returns any error.
func (c Command) Execute() error {
	out, err := exec.Command("sh", "-c", c.CommandValue).CombinedOutput()
	if err != nil {
		fmt.Println(string(out))
		return err
	}

	return nil
}
