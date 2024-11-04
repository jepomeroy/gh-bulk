package repo

import (
	"context"
	"fmt"
	"log"
	"os"

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
	// Clone repository
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

	err = os.Chdir(r.tmpDir)
	if err != nil {
		fmt.Printf("Error changing directory to %s: %s\n", r.tmpDir, err)
		return err
	}

	return nil
}

func (r *Repository) Clean() error {
	// Clean up repository
	fmt.Printf("Cleaning up repository %s\n", r.Name)

	err := os.RemoveAll(r.tmpDir)
	if err != nil {
		fmt.Printf("Error cleaning up repository %s: %s\n", r.Name, err)
		return err
	}

	return nil
}

func (r Repository) CreateBranch(commit commit.Commit) error {
	// Commit and Push changes
	println("Committing and pushing changes")
	// Create a new branch
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

func (r Repository) CommitAndPush(commit commit.Commit) error {
	// Commit changes to the new branch
	w, err := r.gitRepo.Worktree()
	if err != nil {
		fmt.Println("Failed to get worktree:", err)
		return err
	}

	_, err = w.Add(".")
	if err != nil {
		fmt.Println("Error added in changes for commit: ", err)
		return err
	}

	_, err = w.Commit("Add new feature", &git.CommitOptions{Author: &object.Signature{}})
	// 	Author: &object.Signature{
	// 		Name:  "Your Name",
	// 		Email: "your.email@example.com",
	// 		When:  time.Now(),
	// 	},
	// })
	if err != nil {
		fmt.Println("Failed to commit changes: ", err)
		return err
	}

	// Push the new branch to the remote repository
	pushOptions := &git.PushOptions{
		RemoteName: "origin",
	}
	err = r.gitRepo.Push(pushOptions)
	if err != nil {
		fmt.Println("Failed to push branch: ", err)
		return err
	}

	fmt.Printf("Branch %s pushed successfully!\n", commit.BranchName)
	return nil
}

func (r Repository) CreatePR(commit commit.Commit) error {
	if len(commit.CommitMessage) > 0 {
		_, stdErr, err := gh.Exec("pr", "create", "--title", commit.CommitTitle, "--body", commit.CommitMessage)
		if err != nil {
			fmt.Println(stdErr)
			return err
		}
	} else {
		_, stdErr, err := gh.Exec("pr", "create", "--title", commit.CommitTitle, "--body", "")
		if err != nil {
			fmt.Println(stdErr)
			return err
		}
	}

	return nil
}

func FilterReposOptions(client *api.RESTClient, ctx context.Context) []Repository {
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

	ierr := form.Run()

	if ierr != nil {
		fmt.Println(ierr)
		os.Exit(0)
	}

	fmt.Println("Fetching repositories...")
	repos := []Repository{}
	page := 1

	for {
		user := ctx.Value("auth").(Auth).Login
		queryParams := fmt.Sprintf("%s+user:%s&page=%d&sort=name&order=asc", searchQuery, user, page)

		var result map[string]interface{}
		err := client.Get("search/repositories?q="+queryParams, &result)

		if err != nil {
			log.Panic("Error fetching repositories:", err)
			os.Exit(0)
		}

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

	return repos
}

func SelectRepositories(repos []Repository) []Repository {
	var selections []string

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
		os.Exit(0)
	}

	selectedRepos := []Repository{}
	for _, repo := range repos {
		for _, selection := range selections {
			if repo.Name == selection {
				selectedRepos = append(selectedRepos, repo)
			}
		}
	}

	return selectedRepos
}
