package commit

import (
	"errors"

	"github.com/charmbracelet/huh"
)

type Commit struct {
	BranchName       string
	PullRequestTitle string
	CommitMessage    string
}

func NewCommit() (Commit, error) {
	var branchName string
	var prTitle string
	var commitMessage string

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Branch name: ").
				Value(&branchName).
				CharLimit(80).
				Validate(func(s string) error {
					// check if branch name is empty
					if len(s) == 0 {
						return errors.New("Branch name required")
					}

					// check if branch name is valid, only a-z A-Z 0-9 - _ . and / are allowed
					for _, c := range s {
						if !((c >= 'a' && c <= 'z') ||
							(c >= 'A' && c <= 'Z') ||
							(c >= '0' && c <= '9') ||
							c == '-' ||
							c == '_' ||
							c == '.' ||
							c == '/') {
							return errors.New("Branch name can only contain a-z A-Z 0-9 - _ . /")
						}
					}

					return nil
				}),
			huh.NewInput().
				Title("Pull Request title: ").
				Value(&prTitle).
				CharLimit(80).
				Validate(func(s string) error {
					// check if commit title is empty
					if len(s) == 0 {
						return errors.New("Pull Request title required")
					}

					return nil
				}),
			huh.NewText().
				Title("Commit message: ").
				Value(&commitMessage).
				CharLimit(400),
		),
	).WithTheme(huh.ThemeCatppuccin())

	err := form.Run()
	if err != nil {
		return Commit{}, err
	}

	return Commit{
		BranchName:       branchName,
		PullRequestTitle: prTitle,
		CommitMessage:    commitMessage,
	}, nil
}
