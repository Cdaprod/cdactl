// pkg/git/git.go
package git

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/Cdaprod/repocate/pkg/utils"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
)

// Auth holds authentication credentials for Git operations.
type Auth struct {
	Username string
	Password string
	Token    string // Optional: For token-based authentication
}

// Repository encapsulates details about a Git repository and its associated Docker container.
type Repository struct {
	URL         string         // Git repository URL
	Name        string         // Repository name
	Path        string         // Local filesystem path
	Auth        *Auth          // Authentication credentials
	ContainerID string         // Associated Docker container ID
	ImageTag    string         // Docker image tag for the repository
	Instance    *git.Repository // go-git repository instance
	Branch      string         // Current active branch
	Logger      *utils.Logger  // Logger for logging operations
}

// GitClient defines the abstraction for Git operations.
type GitClient interface {
	InitRepo(repo *Repository) error
	CloneRepo(repo *Repository) error
	AddFile(repo *Repository, filePath string) error
	CommitChanges(repo *Repository, message string) (plumbing.Hash, error)
	PushChanges(repo *Repository, branch string) error
	PullChanges(repo *Repository, branch string) error
	GetFileContent(repo *Repository, branch, filePath string) ([]byte, error)
	CheckoutBranch(repo *Repository, branch string) error
	CreateBranch(repo *Repository, branch string) error
	MergeBranch(repo *Repository, sourceBranch, targetBranch string) error
	ListBranches(repo *Repository) ([]string, error)
	AssociateContainer(repo *Repository, containerID string)
	GetAssociatedContainer(repo *Repository) string
}

// gitClient is the concrete implementation of GitClient using go-git.
type gitClient struct {
	logger *utils.Logger
}

// NewGitClient creates a new instance of GitClient.
func NewGitClient(logger *utils.Logger) GitClient {
	return &gitClient{
		logger: logger,
	}
}

// InitRepo initializes a new Git repository at the specified path.
func (gc *gitClient) InitRepo(repo *Repository) error {
	gc.logger.LogInfo(fmt.Sprintf("Initializing Git repository at %s", repo.Path))
	repoInstance, err := git.PlainInit(repo.Path, false)
	if err != nil {
		gc.logger.LogError(fmt.Sprintf("Failed to initialize Git repository: %v", err))
		return err
	}
	repo.Instance = repoInstance
	gc.logger.LogInfo("Git repository initialized successfully.")
	return nil
}

// CloneRepo clones the Git repository to the repository's local path.
func (gc *gitClient) CloneRepo(repo *Repository) error {
	gc.logger.LogInfo(fmt.Sprintf("Cloning repository from %s to %s", repo.URL, repo.Path))
	cloneOptions := &git.CloneOptions{
		URL:      repo.URL,
		Progress: os.Stdout, // Optional: Stream clone progress to stdout
	}

	// Handle authentication if provided
	if repo.Auth != nil && repo.Auth.Token != "" {
		cloneOptions.Auth = &http.BasicAuth{
			Username: repo.Auth.Username, // Can be anything except an empty string
			Password: repo.Auth.Token,
		}
	} else if repo.Auth != nil {
		cloneOptions.Auth = &http.BasicAuth{
			Username: repo.Auth.Username,
			Password: repo.Auth.Password,
		}
	}

	repoInstance, err := git.PlainClone(repo.Path, false, cloneOptions)
	if err != nil {
		gc.logger.LogError(fmt.Sprintf("Failed to clone repository: %v", err))
		return err
	}
	repo.Instance = repoInstance

	// Set the default branch (assuming 'main' exists)
	repo.Branch = "main"
	gc.logger.LogInfo(fmt.Sprintf("Repository cloned successfully to %s", repo.Path))
	return nil
}

// AddFile stages a specified file for commit.
func (gc *gitClient) AddFile(repo *Repository, filePath string) error {
	gc.logger.LogInfo(fmt.Sprintf("Adding file %s to repository %s", filePath, repo.Name))
	wt, err := repo.Instance.Worktree()
	if err != nil {
		gc.logger.LogError(fmt.Sprintf("Failed to get worktree: %v", err))
		return err
	}

	_, err = wt.Add(filePath)
	if err != nil {
		gc.logger.LogError(fmt.Sprintf("Failed to add file %s: %v", filePath, err))
		return err
	}

	gc.logger.LogInfo(fmt.Sprintf("File %s added successfully.", filePath))
	return nil
}

// CommitChanges commits staged changes with a commit message.
func (gc *gitClient) CommitChanges(repo *Repository, message string) (plumbing.Hash, error) {
	gc.logger.LogInfo(fmt.Sprintf("Committing changes in repository %s with message: %s", repo.Name, message))
	wt, err := repo.Instance.Worktree()
	if err != nil {
		gc.logger.LogError(fmt.Sprintf("Failed to get worktree: %v", err))
		return plumbing.ZeroHash, err
	}

	commitHash, err := wt.Commit(message, &git.CommitOptions{
		Author: &object.Signature{
			Name:  "Repocate",
			Email: "repocate@example.com",
			When:  time.Now(),
		},
	})
	if err != nil {
		gc.logger.LogError(fmt.Sprintf("Failed to commit changes: %v", err))
		return plumbing.ZeroHash, err
	}

	gc.logger.LogInfo(fmt.Sprintf("Changes committed successfully with hash %s.", commitHash.String()))
	return commitHash, nil
}

// PushChanges pushes committed changes to the specified branch on the remote repository.
func (gc *gitClient) PushChanges(repo *Repository, branch string) error {
	gc.logger.LogInfo(fmt.Sprintf("Pushing changes to remote repository %s on branch %s", repo.Name, branch))

	pushOptions := &git.PushOptions{
		RemoteName: "origin",
		RefSpecs: []git.ConfigRefSpec{
			git.ConfigRefSpec(fmt.Sprintf("refs/heads/%s:refs/heads/%s", branch, branch)),
		},
	}

	// Handle authentication if provided
	if repo.Auth != nil && repo.Auth.Token != "" {
		pushOptions.Auth = &http.BasicAuth{
			Username: repo.Auth.Username, // Can be anything except an empty string
			Password: repo.Auth.Token,
		}
	} else if repo.Auth != nil {
		pushOptions.Auth = &http.BasicAuth{
			Username: repo.Auth.Username,
			Password: repo.Auth.Password,
		}
	}

	err := repo.Instance.Push(pushOptions)
	if err != nil && err != git.NoErrAlreadyUpToDate {
		gc.logger.LogError(fmt.Sprintf("Failed to push changes: %v", err))
		return err
	}

	gc.logger.LogInfo("Changes pushed to remote repository successfully.")
	return nil
}

// PullChanges pulls and merges changes from the specified branch of the remote repository.
func (gc *gitClient) PullChanges(repo *Repository, branch string) error {
	gc.logger.LogInfo(fmt.Sprintf("Pulling changes from remote repository %s on branch %s", repo.Name, branch))

	pullOptions := &git.PullOptions{
		RemoteName:    "origin",
		ReferenceName: plumbing.NewBranchReferenceName(branch),
		SingleBranch:  true,
		Force:         true,
	}

	// Handle authentication if provided
	if repo.Auth != nil && repo.Auth.Token != "" {
		pullOptions.Auth = &http.BasicAuth{
			Username: repo.Auth.Username, // Can be anything except an empty string
			Password: repo.Auth.Token,
		}
	} else if repo.Auth != nil {
		pullOptions.Auth = &http.BasicAuth{
			Username: repo.Auth.Username,
			Password: repo.Auth.Password,
		}
	}

	err := repo.Instance.Pull(pullOptions)
	if err != nil && err != git.NoErrAlreadyUpToDate {
		gc.logger.LogError(fmt.Sprintf("Failed to pull changes: %v", err))
		return err
	}

	gc.logger.LogInfo("Changes pulled from remote repository successfully.")
	return nil
}

// GetFileContent retrieves the content of a specific file from a specified branch.
func (gc *gitClient) GetFileContent(repo *Repository, branch, filePath string) ([]byte, error) {
	gc.logger.LogInfo(fmt.Sprintf("Retrieving content of file %s from branch %s in repository %s", filePath, branch, repo.Name))

	// Get the reference for the branch
	ref, err := repo.Instance.Reference(plumbing.NewBranchReferenceName(branch), true)
	if err != nil {
		gc.logger.LogError(fmt.Sprintf("Failed to get reference for branch '%s': %v", branch, err))
		return nil, err
	}

	// Get the commit object
	commit, err := repo.Instance.CommitObject(ref.Hash())
	if err != nil {
		gc.logger.LogError(fmt.Sprintf("Failed to get commit object: %v", err))
		return nil, err
	}

	// Get the file
	file, err := commit.File(filePath)
	if err != nil {
		gc.logger.LogError(fmt.Sprintf("Failed to get file '%s' from commit: %v", filePath, err))
		return nil, err
	}

	// Read the file content
	reader, err := file.Reader()
	if err != nil {
		gc.logger.LogError(fmt.Sprintf("Failed to get reader for file '%s': %v", filePath, err))
		return nil, err
	}
	defer reader.Close()

	content, err := io.ReadAll(reader)
	if err != nil {
		gc.logger.LogError(fmt.Sprintf("Failed to read content of file '%s': %v", filePath, err))
		return nil, err
	}

	gc.logger.LogInfo(fmt.Sprintf("Content of file '%s' retrieved successfully.", filePath))
	return content, nil
}

// CheckoutBranch checks out to the specified branch in the repository.
func (gc *gitClient) CheckoutBranch(repo *Repository, branch string) error {
	gc.logger.LogInfo(fmt.Sprintf("Checking out to branch %s in repository %s", branch, repo.Name))

	wt, err := repo.Instance.Worktree()
	if err != nil {
		gc.logger.LogError(fmt.Sprintf("Failed to get worktree: %v", err))
		return err
	}

	err = wt.Checkout(&git.CheckoutOptions{
		Branch: plumbing.NewBranchReferenceName(branch),
		Force:  true,
	})
	if err != nil {
		gc.logger.LogError(fmt.Sprintf("Failed to checkout to branch %s: %v", branch, err))
		return err
	}

	repo.Branch = branch
	gc.logger.LogInfo(fmt.Sprintf("Checked out to branch %s successfully.", branch))
	return nil
}

// CreateBranch creates a new branch in the repository.
func (gc *gitClient) CreateBranch(repo *Repository, branch string) error {
	gc.logger.LogInfo(fmt.Sprintf("Creating new branch %s in repository %s", branch, repo.Name))

	branchRefName := plumbing.NewBranchReferenceName(branch)
	headRef, err := repo.Instance.Head()
	if err != nil {
		gc.logger.LogError(fmt.Sprintf("Failed to get HEAD reference: %v", err))
		return err
	}

	ref := plumbing.NewHashReference(branchRefName, headRef.Hash())
	err = repo.Instance.Storer.SetReference(ref)
	if err != nil {
		gc.logger.LogError(fmt.Sprintf("Failed to create branch %s: %v", branch, err))
		return err
	}

	gc.logger.LogInfo(fmt.Sprintf("Branch %s created successfully.", branch))
	return nil
}

// MergeBranch merges the source branch into the target branch in the repository.
func (gc *gitClient) MergeBranch(repo *Repository, sourceBranch, targetBranch string) error {
	gc.logger.LogInfo(fmt.Sprintf("Merging branch %s into branch %s in repository %s", sourceBranch, targetBranch, repo.Name))

	wt, err := repo.Instance.Worktree()
	if err != nil {
		gc.logger.LogError(fmt.Sprintf("Failed to get worktree: %v", err))
		return err
	}

	// Checkout to target branch
	err = wt.Checkout(&git.CheckoutOptions{
		Branch: plumbing.NewBranchReferenceName(targetBranch),
		Force:  true,
	})
	if err != nil {
		gc.logger.LogError(fmt.Sprintf("Failed to checkout to branch %s: %v", targetBranch, err))
		return err
	}

	// Merge source branch into target branch
	mergeOptions := &git.MergeOptions{
		Commit: true,
		Author: &object.Signature{
			Name:  "Repocate",
			Email: "repocate@example.com",
			When:  time.Now(),
		},
	}

	err = repo.Instance.Merge(&plumbing.Reference{Name: plumbing.NewBranchReferenceName(sourceBranch)}, mergeOptions)
	if err != nil {
		gc.logger.LogError(fmt.Sprintf("Failed to merge branch %s into %s: %v", sourceBranch, targetBranch, err))
		return err
	}

	gc.logger.LogInfo(fmt.Sprintf("Branch %s merged into %s successfully.", sourceBranch, targetBranch))
	return nil
}

// ListBranches lists all branches in the repository.
func (gc *gitClient) ListBranches(repo *Repository) ([]string, error) {
	gc.logger.LogInfo(fmt.Sprintf("Listing all branches in repository %s", repo.Name))

	branches, err := repo.Instance.Branches()
	if err != nil {
		gc.logger.LogError(fmt.Sprintf("Failed to list branches: %v", err))
		return nil, err
	}

	var branchNames []string
	err = branches.ForEach(func(b *plumbing.Reference) error {
		branchName := b.Name().Short()
		branchNames = append(branchNames, branchName)
		return nil
	})
	if err != nil {
		gc.logger.LogError(fmt.Sprintf("Failed during branch iteration: %v", err))
		return nil, err
	}

	gc.logger.LogInfo("Branches listed successfully.")
	return branchNames, nil
}

// AssociateContainer associates a Docker container ID with the repository.
func (gc *gitClient) AssociateContainer(repo *Repository, containerID string) {
	repo.ContainerID = containerID
	gc.logger.LogInfo(fmt.Sprintf("Repository %s associated with container %s", repo.Name, containerID))
}

// GetAssociatedContainer retrieves the Docker container ID associated with the repository.
func (gc *gitClient) GetAssociatedContainer(repo *Repository) string {
	return repo.ContainerID
}