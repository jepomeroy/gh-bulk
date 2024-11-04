package main

import (
	"context"
	"fmt"
	"os"
	"path"

	"github.com/charmbracelet/huh"
	"github.com/cli/go-gh/v2/pkg/api"
	"github.com/jonepom/gh-bulk/internal/commit"
	"github.com/jonepom/gh-bulk/internal/execute"
	"github.com/jonepom/gh-bulk/internal/repo"
)

func main() {
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Println(err)
		os.Exit(0)
		return
	}

	client, err := api.DefaultRESTClient()
	if err != nil {
		fmt.Println("Error creating API client:", err)
		os.Exit(0)
		return
	}

	err = client.Get("user", &repo.UserAuth)
	if err != nil {
		fmt.Println(err)
		os.Exit(0)
		return
	}

	ctx := context.Background()
	ctx = context.WithValue(ctx, "auth", repo.UserAuth)

	repoOptions, err := repo.FilterReposOptions(client, ctx)
	if err != nil {
		fmt.Println(err)
		os.Exit(0)
		return
	}

	repos, err := repo.SelectRepositories(repoOptions)
	if err != nil {
		fmt.Println(err)
		os.Exit(0)
		return
	}

	commit, err := commit.NewCommit()
	if err != nil {
		fmt.Println(err)
		os.Exit(0)
		return
	}

	command, err := execute.GetCommand()
	if err != nil {
		fmt.Println(err)
		os.Exit(0)
		return
	}

	if validate(command, commit, repos) == false {
		fmt.Println("Aborting...")
		os.Exit(0)
		return
	}

	processRepos(cwd, repos, command, commit)
}

func processRepos(cwd string, repos []repo.Repository, command execute.Command, commit commit.Commit) {
	for _, r := range repos {
		tempDir := path.Join(os.TempDir(), r.Name)
		err := r.Clone(tempDir)
		if err != nil {
			fmt.Println("Error cloning repository:", tempDir, err)
			clean(cwd, r)
			continue
		}

		err = r.CreateBranch(commit)
		if err != nil {
			fmt.Println("Error creating branch:", err)
			clean(cwd, r)
			continue
		}

		err = command.Execute()
		if err != nil {
			fmt.Println("Error executing command:", err)
			clean(cwd, r)
			continue
		}

		err = r.CommitAndPush(commit)
		if err != nil {
			fmt.Println("Error committing and pushing:", err)
			clean(cwd, r)
			continue
		}

		err = r.CreatePR(commit)
		if err != nil {
			fmt.Println("Error creating PR:", err)
			clean(cwd, r)
			continue
		}

		clean(cwd, r)
	}
}

func clean(cwd string, r repo.Repository) {
	os.Chdir(cwd)
	err := r.Clean()

	if err != nil {
		fmt.Printf("Error cleaning %s, %s\n", r.Name, err)
	}
}
func makeDescription(command execute.Command, commit commit.Commit, selectedRepos []repo.Repository) string {
	description := fmt.Sprintf("%-20s %s\n%-20s %s\n%-20s %s\n%-20s %s\n\n",
		"command:",
		command.CommandValue,
		"branch name:",
		commit.BranchName,
		"pull request title:",
		commit.PullRequestTitle,
		"commit message:",
		commit.CommitMessage,
	)

	description += "Repositories:\n"
	for _, r := range selectedRepos {
		description += fmt.Sprintf("  %s\n", r.Name)
	}

	return description
}
func validate(command execute.Command, commit commit.Commit, selectedRepos []repo.Repository) bool {
	var confirm bool
	description := makeDescription(command, commit, selectedRepos)

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title("Verify Info").
				Description(description).
				Affirmative("Correct").
				Negative("Abort").
				Value(&confirm),
		),
	)

	err := form.Run()
	if err != nil {
		return false
	}

	return confirm
}
