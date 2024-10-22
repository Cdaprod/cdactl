// pkg/dotfile/dotfile_test.go

package dotfile

import (
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
)

// Helper function to create a temporary config file
func createTempConfig(t *testing.T, content string) string {
	t.Helper()
	tempDir, err := ioutil.TempDir("", "dotfile_test_config")
	if err != nil {
		t.Fatalf("Failed to create temp config dir: %v", err)
	}

	configPath := filepath.Join(tempDir, "config.yaml")
	if err := ioutil.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	// Set HOME to tempDir for testing
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)

	return configPath
}

// Helper function to reset HOME after tests
func resetHome(originalHome string) {
	os.Setenv("HOME", originalHome)
}

// TestLoadConfig verifies that the configuration is loaded correctly
func TestLoadConfig(t *testing.T) {
	originalHome := os.Getenv("HOME")
	defer resetHome(originalHome)

	configContent := `
dotfile:
  repo_path: "/tmp/dotfile_repo"
  remote_url: "file:///tmp/dotfile_remote.git"
  branch_template: "{hostname}/{arch}"
  excludes:
    - ".git"
`
	configPath := createTempConfig(t, configContent)
	defer os.RemoveAll(filepath.Dir(configPath))

	config, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	expectedRepoPath := "/tmp/dotfile_repo"
	if config.RepoPath != expectedRepoPath {
		t.Errorf("Expected RepoPath %s, got %s", expectedRepoPath, config.RepoPath)
	}

	expectedRemoteURL := "file:///tmp/dotfile_remote.git"
	if config.RemoteURL != expectedRemoteURL {
		t.Errorf("Expected RemoteURL %s, got %s", expectedRemoteURL, config.RemoteURL)
	}

	expectedBranchTemplate := "{hostname}/{arch}"
	if config.BranchTemplate != expectedBranchTemplate {
		t.Errorf("Expected BranchTemplate %s, got %s", expectedBranchTemplate, config.BranchTemplate)
	}

	if len(config.Excludes) != 1 || config.Excludes[0] != ".git" {
		t.Errorf("Expected Excludes to contain only .git, got %v", config.Excludes)
	}
}

// TestInitDotfile verifies that the dotfiles repository is initialized correctly
func TestInitDotfile(t *testing.T) {
	originalHome := os.Getenv("HOME")
	defer resetHome(originalHome)

	// Setup temporary directories for repo and remote
	repoDir, err := ioutil.TempDir("", "dotfile_repo")
	if err != nil {
		t.Fatalf("Failed to create temp repo dir: %v", err)
	}
	defer os.RemoveAll(repoDir)

	remoteDir, err := ioutil.TempDir("", "dotfile_remote")
	if err != nil {
		t.Fatalf("Failed to create temp remote dir: %v", err)
	}
	defer os.RemoveAll(remoteDir)

	// Initialize remote bare repository
	remoteRepo, err := git.PlainInit(remoteDir+".git", true)
	if err != nil {
		t.Fatalf("Failed to initialize remote bare repo: %v", err)
	}

	// Initialize DotfileConfig
	config := &DotfileConfig{
		RepoPath:       repoDir,
		RemoteURL:      "file://" + remoteDir + ".git",
		BranchTemplate: "testhost/amd64",
		Excludes:       []string{".git"},
	}

	// Initialize DotfileModel
	model := &DotfileModel{
		config:        config,
		gitRepo:       nil, // Repository not initialized yet
		currentBranch: "testhost/amd64",
	}

	// Execute initDotfile
	msg := model.initDotfile()

	if model.gitRepo == nil {
		t.Fatalf("Repository not initialized")
	}

	// Verify that the branch exists in remote
	remoteRepoRef, err := remoteRepo.Reference(plumbing.NewBranchReferenceName("testhost/amd64"), true)
	if err != nil {
		t.Fatalf("Failed to get reference from remote: %v", err)
	}

	if remoteRepoRef.Name().Short() != "testhost/amd64" {
		t.Errorf("Expected branch name 'testhost/amd64', got '%s'", remoteRepoRef.Name().Short())
	}

	// Check the initial commit message
	commit, err := model.gitRepo.CommitObject(remoteRepoRef.Hash())
	if err != nil {
		t.Fatalf("Failed to get commit object: %v", err)
	}

	if commit.Message != "Initial commit" {
		t.Errorf("Expected commit message 'Initial commit', got '%s'", commit.Message)
	}

	// Check if the remote is set correctly
	remotes, err := model.gitRepo.Remotes()
	if err != nil {
		t.Fatalf("Failed to get remotes: %v", err)
	}

	if len(remotes) != 1 {
		t.Fatalf("Expected 1 remote, got %d", len(remotes))
	}

	if remotes[0].Config().Name != "origin" || remotes[0].Config().URLs[0] != config.RemoteURL {
		t.Errorf("Remote origin not set correctly. Got name: %s, URL: %s", remotes[0].Config().Name, remotes[0].Config().URLs[0])
	}

	// Verify the commit exists in the local repo
	localCommit, err := model.gitRepo.CommitObject(remoteRepoRef.Hash())
	if err != nil {
		t.Fatalf("Failed to get local commit object: %v", err)
	}

	if localCommit.Message != "Initial commit" {
		t.Errorf("Expected local commit message 'Initial commit', got '%s'", localCommit.Message)
	}
}

// TestAddDotfile verifies that a dotfile can be added, committed, and pushed correctly
func TestAddDotfile(t *testing.T) {
	originalHome := os.Getenv("HOME")
	defer resetHome(originalHome)

	// Setup temporary directories for repo and remote
	repoDir, err := ioutil.TempDir("", "dotfile_repo")
	if err != nil {
		t.Fatalf("Failed to create temp repo dir: %v", err)
	}
	defer os.RemoveAll(repoDir)

	remoteDir, err := ioutil.TempDir("", "dotfile_remote")
	if err != nil {
		t.Fatalf("Failed to create temp remote dir: %v", err)
	}
	defer os.RemoveAll(remoteDir)

	// Initialize remote bare repository
	remoteRepo, err := git.PlainInit(remoteDir+".git", true)
	if err != nil {
		t.Fatalf("Failed to initialize remote bare repo: %v", err)
	}

	// Initialize local bare repository
	localRepo, err := git.PlainInit(repoDir, true)
	if err != nil {
		t.Fatalf("Failed to initialize local bare repo: %v", err)
	}

	// Add remote origin
	_, err = localRepo.CreateRemote(&config.RemoteConfig{
		Name: "origin",
		URLs: []string{"file://" + remoteDir + ".git"},
	})
	if err != nil && err != git.ErrRemoteExists {
		t.Fatalf("Failed to add remote origin: %v", err)
	}

	// Initialize DotfileConfig
	config := &DotfileConfig{
		RepoPath:       repoDir,
		RemoteURL:      "file://" + remoteDir + ".git",
		BranchTemplate: "testhost/amd64",
		Excludes:       []string{".git"},
	}

	// Initialize DotfileModel
	model := &DotfileModel{
		config:        config,
		gitRepo:       localRepo,
		currentBranch: "testhost/amd64",
	}

	// Initialize the repository
	model.initDotfile()

	// Create a temporary dotfile to add
	tempFile, err := ioutil.TempFile("", "testdotfile.*.txt")
	if err != nil {
		t.Fatalf("Failed to create temp dotfile: %v", err)
	}
	defer os.Remove(tempFile.Name())

	dotfileContent := "This is a test dotfile."
	if _, err := tempFile.WriteString(dotfileContent); err != nil {
		t.Fatalf("Failed to write to temp dotfile: %v", err)
	}
	tempFile.Close()

	// Simulate adding the dotfile
	// Since addDotfile reads from stdin, we need to refactor the code to accept a file path for testing
	// For this test, we'll assume a refactored version of addDotfile that accepts a file path

	// Refactored addDotfile function (for testing purposes)
	addDotfileTest := func(model *DotfileModel, filePath string) error {
		fullPath := filepath.Join(os.Getenv("HOME"), filePath)
		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			return fmt.Errorf("file does not exist: %s", fullPath)
		}

		// Stage the file
		wt, err := model.gitRepo.Worktree()
		if err != nil {
			return fmt.Errorf("failed to get worktree: %v", err)
		}

		_, err = wt.Add(filePath)
		if err != nil {
			return fmt.Errorf("failed to add file: %v", err)
		}

		// Commit the change
		commitMsg := fmt.Sprintf("Add %s", filePath)
		_, err = wt.Commit(commitMsg, &git.CommitOptions{
			Author: &object.Signature{
				Name:  "cdactl",
				Email: "cdactl@example.com",
				When:  time.Now(),
			},
		})
		if err != nil {
			return fmt.Errorf("failed to commit: %v", err)
		}

		// Push the change
		auth, err := sshAuth()
		if err != nil {
			return fmt.Errorf("failed to get SSH auth: %v", err)
		}

		err = model.gitRepo.Push(&git.PushOptions{
			Auth:       auth,
			RemoteName: "origin",
		})
		if err != nil && err != git.NoErrAlreadyUpToDate {
			return fmt.Errorf("failed to push: %v", err)
		}

		return nil
	}

	// Copy the temp dotfile to the HOME directory as expected by the addDotfile function
	dotfileRelPath, err := filepath.Rel(os.Getenv("HOME"), tempFile.Name())
	if err != nil {
		t.Fatalf("Failed to get relative path: %v", err)
	}
	if err := os.Rename(tempFile.Name(), filepath.Join(os.Getenv("HOME"), dotfileRelPath)); err != nil {
		t.Fatalf("Failed to move dotfile to HOME: %v", err)
	}

	// Add the dotfile using the refactored function
	err = addDotfileTest(model, dotfileRelPath)
	if err != nil {
		t.Fatalf("addDotfile failed: %v", err)
	}

	// Verify that the file exists in the remote repository
	remoteRepo, err = git.PlainOpen(remoteDir + ".git")
	if err != nil {
		t.Fatalf("Failed to open remote repository: %v", err)
	}

	// Check if the file exists in the remote repository's branch
	remoteRef, err := remoteRepo.Reference(plumbing.NewBranchReferenceName("testhost/amd64"), true)
	if err != nil {
		t.Fatalf("Failed to get remote branch reference: %v", err)
	}

	remoteCommit, err := remoteRepo.CommitObject(remoteRef.Hash())
	if err != nil {
		t.Fatalf("Failed to get remote commit object: %v", err)
	}

	remoteTree, err := remoteCommit.Tree()
	if err != nil {
		t.Fatalf("Failed to get remote commit tree: %v", err)
	}

	_, err = remoteTree.File(dotfileRelPath)
	if err != nil {
		t.Fatalf("Dotfile not found in remote repository: %v", err)
	}
}

// TestSSHAuth verifies that SSH authentication is set up correctly
func TestSSHAuth(t *testing.T) {
	// This test assumes that ssh-agent is running and SSH keys are added
	auth, err := sshAuth()
	if err != nil {
		t.Fatalf("sshAuth failed: %v", err)
	}

	if auth == nil {
		t.Fatal("sshAuth returned nil AuthMethod")
	}
}

// Additional tests like TestPullDotfiles, TestPullSpecificDotfile, and TestSyncDotfiles
// would follow a similar pattern, initializing repositories, performing operations, and verifying outcomes.
// These tests would require further refactoring of the DotfileModel to allow injecting parameters
// instead of relying on user input via stdin.