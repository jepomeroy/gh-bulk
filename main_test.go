package main

import (
	"strings"
	"testing"

	"github.com/jepomeroy/gh-bulk/internal/commit"
	"github.com/jepomeroy/gh-bulk/internal/execute"
	"github.com/jepomeroy/gh-bulk/internal/repo"
)

func TestMakeDescription(t *testing.T) {
	cmd := execute.Command{CommandValue: "go mod tidy"}
	c := commit.Commit{
		BranchName:       "fix/deps",
		PullRequestTitle: "Fix dependencies",
		CommitMessage:    "Update go.mod and go.sum",
	}
	repos := []repo.Repository{
		{Name: "repo-a"},
		{Name: "repo-b"},
	}

	got := makeDescription(cmd, c, repos)

	for _, want := range []string{
		"go mod tidy", "fix/deps", "Fix dependencies", "Update go.mod and go.sum", "repo-a", "repo-b",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("description missing %q\ngot:\n%s", want, got)
		}
	}
}
