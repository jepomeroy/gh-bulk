// Package commit provides the Commit type for capturing git branch, commit, and pull request metadata.
package commit

import (
	"errors"

	"github.com/charmbracelet/huh"
)

// Commit holds the metadata needed to create a branch, commit changes, and open a pull request.
type Commit struct {
	BranchName       string
	PullRequestTitle string
	CommitMessage    string
}

// NewCommit prompts the user interactively for branch name, pull request title, and commit message.
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
					if len(s) == 0 {
						return errors.New("Branch name required")
					}

					// only a-z A-Z 0-9 - _ . / are valid branch name characters
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
