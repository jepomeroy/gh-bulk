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
	"github.com/jonepom/gh-bulk/internal/commit"
)

var (
	UserAuth Auth
)

type Auth struct {
	Login string
}

type Repository struct {
	Name    string
	SSHURL  string
	tmpDir  string
	gitRepo *git.Repository
}

func (r *Repository) Clone(tempDir string) error {
	fmt.Printf("Cloning repository %s\n", r.Name)

	// Clone repository
	_, stdErr, err := gh.Exec("repo", "clone", r.SSHURL, tempDir)
	if err != nil {
		fmt.Printf("Error cloning repository %s: %s\n", r.Name, err)
		fmt.Printf("Output: %s\n", stdErr.String())
		return err
	}

	// Set the tempdir and open the repos
	r.tmpDir = tempDir
	gitRepo, err := git.PlainOpen(tempDir)
	if err != nil {
		fmt.Println(err)
		return err
	}

	// Store the repo reference
	r.gitRepo = gitRepo

	// set the current working dir to the tempDir
	// everything is relative to the tempDir
	err = os.Chdir(r.tmpDir)
	if err != nil {
		fmt.Printf("Error changing directory to %s: %s\n", r.tmpDir, err)
		return err
	}

	return nil
}

func (r *Repository) Clean() error {
	fmt.Printf("Cleaning up repository %s\n", r.Name)

	// Clean up repository
	err := os.RemoveAll(r.tmpDir)
	if err != nil {
		fmt.Printf("Error cleaning up repository %s: %s\n", r.Name, err)
		return err
	}

	return nil
}

func (r Repository) CreateBranch(commit commit.Commit) error {
	println("Committing and pushing changes")

	w, err := r.gitRepo.Worktree()
	if err != nil {
		fmt.Println("Failed to get worktree:", err)
		return err
	}

	// create a new branch and check it out
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

func (r Repository) CommitAndPush(commit commit.Commit) error {
	w, err := r.gitRepo.Worktree()
	if err != nil {
		fmt.Println("Failed to get worktree:", err)
		return err
	}

	// stage all the changes from commmand execution
	_, err = w.Add(".")
	if err != nil {
		fmt.Println("Error added in changes for commit:", err)
		return err
	}

	// commit the changes
	_, err = w.Commit(commit.CommitMessage, &git.CommitOptions{Author: &object.Signature{
		Name: "GH Bulk Extension",
		When: time.Now(),
	}})
	if err != nil {
		fmt.Println("Failed to commit changes:", err)
		return err
	}

	// Push the new branch to the remote repository
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

func (r Repository) CreatePR(commit commit.Commit) error {
	// commit the PR
	_, stdErr, err := gh.Exec("pr", "create", "--title", commit.PullRequestTitle, "--body", commit.CommitMessage)
	if err != nil {
		fmt.Println(stdErr.String())
		return err
	}

	return nil
}

// Filter the repositories based on the search query. The query is a simple string match
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

	// Retrieve all repostories for the authenticated user
	for {
		// get the authenticated user
		user := ctx.Value("auth").(Auth).Login
		// build the query parameters
		queryParams := fmt.Sprintf("%s+user:%s&page=%d&sort=name&order=asc", searchQuery, user, page)

		// Fetch repositories
		var result map[string]interface{}
		err := client.Get("search/repositories?q="+queryParams, &result)
		if err != nil {
			fmt.Println("Error fetching repositories:", err)
			return []Repository{}, err
		}

		// Parese the result and append to the repos slice
		if items, ok := result["items"].([]interface{}); ok {
			for _, item := range items {
				if repo, ok := item.(map[string]interface{}); ok {
					name := repo["name"].(string)
					sshURL := repo["ssh_url"].(string)
					repos = append(repos, Repository{Name: name, SSHURL: sshURL})
				}
			}
		}

		// Break if there are no more pages
		total_count := int(result["total_count"].(float64))
		if total_count == len(repos) {
			break
		}

		// Fetch next page
		page++
	}

	return repos, nil
}

// Using the result of the FilterReposOptions function, select the repositories to process
func SelectRepositories(repos []Repository) ([]Repository, error) {
	var selections []string

	// Create a multi-select form from the repository names
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
	).
		WithTheme(huh.ThemeCatppuccin())

	err := form.Run()
	if err != nil {
		return []Repository{}, err
	}

	// Get the selected repositories and return them
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
