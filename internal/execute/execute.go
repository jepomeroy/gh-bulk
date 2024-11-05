package execute

import (
	"fmt"
	"os/exec"

	"github.com/charmbracelet/huh"
)

type Command struct {
	CommandValue string
}

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

func (c Command) Execute() error {
	out, err := exec.Command("sh", "-c", c.CommandValue).CombinedOutput()
	if err != nil {
		fmt.Println(string(out))
		return err
	}

	return nil
}
