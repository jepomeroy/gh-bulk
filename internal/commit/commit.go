package commit

import (
	"errors"

	"github.com/charmbracelet/huh"
)

type Commit struct {
	BranchName    string
	CommitTitle   string
	CommitMessage string
}

func NewCommit() (Commit, error) {
	var branchName string
	var commitTitle string
	var commitMessage string

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Branch name: ").
				Value(&branchName).
				CharLimit(80).
				Validate(func(s string) error {
					if len(s) == 0 {
						return errors.New("Branch name required")
					}

					return nil
				}),
			huh.NewInput().
				Title("Commit title: ").
				Value(&commitTitle).
				CharLimit(80).
				Validate(func(s string) error {
					if len(s) == 0 {
						return errors.New("Commit title required")
					}

					return nil
				}),
			huh.NewText().
				Title("Commit message: ").
				Value(&commitMessage).
				CharLimit(400),
		),
	)

	err := form.Run()
	if err != nil {
		return Commit{}, err
	}

	return Commit{
		BranchName:    branchName,
		CommitTitle:   commitTitle,
		CommitMessage: commitMessage,
	}, nil
}
