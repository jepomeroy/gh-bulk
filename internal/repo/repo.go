// Package repo provides types and functions for discovering, cloning, and automating changes across GitHub repositories.
package repo

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/charmbracelet/huh"
	"github.com/cli/go-gh/v2"
	"github.com/cli/go-gh/v2/pkg/api"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/jepomeroy/gh-bulk/internal/commit"
)

type AuthUserKey string

// Repository represents a GitHub repository with its name, SSH URL, and local clone state.
type Repository struct {
	Name    string
	SSHURL  string
	tmpDir  string
	gitRepo *git.Repository
}

// Clone clones r to tempDir using the GitHub CLI and changes the working directory to tempDir.
func (r *Repository) Clone(tempDir string) error {
	fmt.Printf("Cloning repository %s\n", r.Name)

	_, stdErr, err := gh.Exec("repo", "clone", r.SSHURL, tempDir)
	if err != nil {
		fmt.Printf("Error cloning repository %s: %s\n", r.Name, err)
		fmt.Printf("Output: %s\n", stdErr.String())
		return err
	}

	r.tmpDir = tempDir
	gitRepo, err := git.PlainOpen(tempDir)
	if err != nil {
		fmt.Println(err)
		return err
	}

	r.gitRepo = gitRepo

	// Shell commands executed via Execute run relative to the cloned repository.
	err = os.Chdir(r.tmpDir)
	if err != nil {
		fmt.Printf("Error changing directory to %s: %s\n", r.tmpDir, err)
		return err
	}

	return nil
}

// Clean removes the cloned repository from the temporary directory.
func (r *Repository) Clean() error {
	fmt.Printf("Cleaning up repository %s\n", r.Name)

	err := os.RemoveAll(r.tmpDir)
	if err != nil {
		fmt.Printf("Error cleaning up repository %s: %s\n", r.Name, err)
		return err
	}

	return nil
}

// CreateBranch creates and checks out a new branch named by commit.BranchName.
func (r Repository) CreateBranch(commit commit.Commit) error {
	println("Committing and pushing changes")

	w, err := r.gitRepo.Worktree()
	if err != nil {
		fmt.Println("Failed to get worktree:", err)
		return err
	}

	newBranch := plumbing.NewBranchReferenceName(commit.BranchName)
	err = w.Checkout(&git.CheckoutOptions{
		Branch: newBranch,
		Create: true,
		Force:  true,
	})
	if err != nil {
		fmt.Println("Failed to checkout new branch:", err)
		return err
	}

	return nil
}

// CommitAndPush stages all changes, commits with commit.CommitMessage, and pushes to origin.
func (r Repository) CommitAndPush(commit commit.Commit) error {
	w, err := r.gitRepo.Worktree()
	if err != nil {
		fmt.Println("Failed to get worktree:", err)
		return err
	}

	_, err = w.Add(".")
	if err != nil {
		fmt.Println("Error added in changes for commit:", err)
		return err
	}

	_, err = w.Commit(commit.CommitMessage, &git.CommitOptions{Author: &object.Signature{
		Name: "GH Bulk Extension",
		When: time.Now(),
	}})
	if err != nil {
		fmt.Println("Failed to commit changes:", err)
		return err
	}

	pushOptions := &git.PushOptions{
		RemoteName: "origin",
	}
	err = r.gitRepo.Push(pushOptions)
	if err != nil {
		fmt.Println("Failed to push branch:", err)
		return err
	}

	fmt.Printf("Branch %s pushed successfully!\n", commit.BranchName)
	return nil
}

// CreatePR opens a pull request using the commit's title and message as body.
func (r Repository) CreatePR(commit commit.Commit) error {
	_, stdErr, err := gh.Exec("pr", "create", "--title", commit.PullRequestTitle, "--body", commit.CommitMessage)
	if err != nil {
		fmt.Println(stdErr.String())
		return err
	}

	return nil
}

// FilterReposOptions prompts for a search filter and returns matching non-archived repositories.
func FilterReposOptions(client *api.RESTClient, ctx context.Context) ([]Repository, error) {
	var searchQuery string

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Search").
				Prompt("filter: ").
				Description("Empty query will return all repositories").
				Value(&searchQuery),
		),
	).WithTheme(huh.ThemeCatppuccin())

	err := form.Run()
	if err != nil {
		return []Repository{}, err
	}

	fmt.Println("Fetching repositories...")
	repos := []Repository{}
	page := 1

	for {
		user := ctx.Value(AuthUserKey("auth"))
		queryParams := fmt.Sprintf("%s+user:%s+archived:false&page=%d&sort=name&order=asc", searchQuery, user, page)

		var result map[string]any
		err := client.Get("search/repositories?q="+queryParams, &result)
		if err != nil {
			fmt.Println("Error fetching repositories:", err)
			return []Repository{}, err
		}

		if items, ok := result["items"].([]any); ok {
			for _, item := range items {
				if repo, ok := item.(map[string]any); ok {
					name := repo["name"].(string)
					sshURL := repo["ssh_url"].(string)
					repos = append(repos, Repository{Name: name, SSHURL: sshURL})
				}
			}
		}

		totalCount := int(result["total_count"].(float64))
		if totalCount == len(repos) {
			break
		}

		page++
	}

	return repos, nil
}

// SelectRepositories presents a multi-select prompt and returns the chosen repositories from repos.
func SelectRepositories(repos []Repository) ([]Repository, error) {
	var selections []string
	customKeyMap := huh.NewDefaultKeyMap()

	// customize keymap help
	customKeyMap.MultiSelect.SelectAll.SetHelp("ctrl+shift+a", "select all")
	customKeyMap.MultiSelect.SelectNone.SetHelp("ctrl+shift+a", "select none")

	repoOptions := []huh.Option[string]{}

	for _, repo := range repos {
		repoOptions = append(repoOptions, huh.NewOption(repo.Name, repo.Name))
	}

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewMultiSelect[string]().
				Title("Select Repositories to Process").
				Options(repoOptions...).
				Filterable(false).
				Value(&selections).
				Height(20),
		),
	).WithKeyMap(customKeyMap).WithTheme(huh.ThemeCatppuccin())

	err := form.Run()
	if err != nil {
		return []Repository{}, err
	}

	selectedRepos := []Repository{}
	for _, repo := range repos {
		for _, selection := range selections {
			if repo.Name == selection {
				selectedRepos = append(selectedRepos, repo)
			}
		}
	}

	return selectedRepos, nil
}
