package dotfile

import (
    "bufio"
    "fmt"
    "net"
    "os"
    "os/exec"
    "path/filepath"
    "strings"
    "time"

    tea "github.com/charmbracelet/bubbletea"
    git "github.com/go-git/go-git/v5"
    "github.com/go-git/go-git/v5/config"
    "github.com/go-git/go-git/v5/plumbing"
    "github.com/go-git/go-git/v5/plumbing/object"
    "github.com/go-git/go-git/v5/plumbing/transport/ssh"
    "github.com/spf13/cobra"
    "github.com/spf13/viper"
    "golang.org/x/crypto/ssh/agent"
)

// DotfileConfig holds the configuration for the dotfile package
type DotfileConfig struct {
    RepoPath       string   `mapstructure:"repo_path"`
    RemoteURL      string   `mapstructure:"remote_url"`
    BranchTemplate string   `mapstructure:"branch_template"`
    Excludes       []string `mapstructure:"excludes"`
}

// LoadConfig loads the dotfile configuration using Viper
func LoadConfig() (*DotfileConfig, error) {
    viper.SetConfigName("config")
    viper.SetConfigType("yaml")
    viper.AddConfigPath(filepath.Join(os.Getenv("HOME"), ".config", "cdactl"))

    if err := viper.ReadInConfig(); err != nil {
        return nil, err
    }

    var config DotfileConfig
    if err := viper.UnmarshalKey("dotfile", &config); err != nil {
        return nil, err
    }

    return &config, nil
}

// NewDotfileCmd creates the cobra command for dotfile management
func NewDotfileCmd() *cobra.Command {
    return &cobra.Command{
        Use:   "dotfile",
        Short: "Manage dotfiles",
        Long:  `Initialize, add, pull, pull specific files, and sync dotfiles across devices.`,
        Run: func(cmd *cobra.Command, args []string) {
            model, err := NewDotfileModel()
            if err != nil {
                fmt.Println("✖ Failed to initialize Dotfile module:", err)
                os.Exit(1)
            }

            if err := tea.NewProgram(model).Start(); err != nil {
                fmt.Println("✖ Bubble Tea program failed:", err)
                os.Exit(1)
            }
        },
    }
}

// DotfileModel defines the Bubble Tea model for Dotfile management
type DotfileModel struct {
    cursor        int
    choices       []string
    quitting      bool
    errorMsg      string
    config        *DotfileConfig
    gitRepo       *git.Repository
    currentBranch string
}

// NewDotfileModel initializes the DotfileModel
func NewDotfileModel() (*DotfileModel, error) {
    config, err := LoadConfig()
    if err != nil {
        return nil, err
    }

    var repo *git.Repository
    if _, err := os.Stat(config.RepoPath); os.IsNotExist(err) {
        // Repository does not exist
        repo = nil
    } else {
        // Open existing repository
        repo, err = git.PlainOpen(config.RepoPath)
        if err != nil {
            return nil, fmt.Errorf("failed to open Git repository: %v", err)
        }
    }

    // Determine current branch name based on template
    hostname, _ := os.Hostname()
    archBytes, err := exec.Command("uname", "-m").Output()
    arch := strings.TrimSpace(string(archBytes))
    branchName := config.BranchTemplate
    branchName = strings.ReplaceAll(branchName, "{hostname}", hostname)
    branchName = strings.ReplaceAll(branchName, "{arch}", arch)

    return &DotfileModel{
        choices:       []string{"Initialize", "Add Dotfile", "Pull Dotfiles", "Pull Specific Dotfile", "Sync Dotfiles", "Exit"},
        config:        config,
        gitRepo:       repo,
        currentBranch: branchName,
    }, nil
}

// Init is the Bubble Tea Init function
func (m DotfileModel) Init() tea.Cmd {
    return nil
}

// Update handles messages and updates the model state
func (m DotfileModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {

    case tea.KeyMsg:
        switch msg.String() {

        case "ctrl+c", "q":
            m.quitting = true
            return m, tea.Quit

        case "up", "k":
            if m.cursor > 0 {
                m.cursor--
            }

        case "down", "j":
            if m.cursor < len(m.choices)-1 {
                m.cursor++
            }

        case "enter":
            selected := m.choices[m.cursor]
            switch selected {
            case "Initialize":
                return m, m.initDotfile()
            case "Add Dotfile":
                return m, m.addDotfile()
            case "Pull Dotfiles":
                return m, m.pullDotfiles()
            case "Pull Specific Dotfile":
                return m, m.pullSpecificDotfile()
            case "Sync Dotfiles":
                return m, m.syncDotfiles()
            case "Exit":
                m.quitting = true
                return m, tea.Quit
            }
        }
    }

    return m, nil
}

// View renders the Bubble Tea UI
func (m DotfileModel) View() string {
    if m.quitting {
        return ""
    }

    s := "=== Dotfile Management ===\n\n"

    for i, choice := range m.choices {
        cursor := " " // no cursor
        if m.cursor == i {
            cursor = ">" // cursor
        }
        line := fmt.Sprintf("%s %s", cursor, choice)
        if m.cursor == i {
            line = fmt.Sprintf("\x1b[7m%s\x1b[0m", line) // Invert colors for selection
        }
        s += line + "\n"
    }

    s += "\nPress q to quit.\n"

    if m.errorMsg != "" {
        s += fmt.Sprintf("\n✖ %s\n", m.errorMsg)
    }

    return s
}

// sshAuth returns the SSH authentication method
func sshAuth() (ssh.AuthMethod, error) {
    sshAgentSock := os.Getenv("SSH_AUTH_SOCK")
    if sshAgentSock == "" {
        return nil, fmt.Errorf("SSH_AUTH_SOCK not set. Ensure ssh-agent is running.")
    }

    conn, err := net.Dial("unix", sshAgentSock)
    if err != nil {
        return nil, fmt.Errorf("Failed to connect to SSH agent: %v", err)
    }

    agentClient := agent.NewClient(conn)
    return ssh.PublicKeysCallback(agentClient.Signers), nil
}

// initDotfile initializes the dotfiles repository
func (m *DotfileModel) initDotfile() tea.Cmd {
    return func() tea.Msg {
        if m.gitRepo != nil {
            m.errorMsg = "Repository already initialized."
            return nil
        }

        // Initialize bare repository
        repo, err := git.PlainInit(m.config.RepoPath, true) // Bare repository
        if err != nil {
            m.errorMsg = fmt.Sprintf("Failed to initialize bare repository: %v", err)
            return nil
        }
        m.gitRepo = repo

        // Set up remote origin
        _, err = m.gitRepo.CreateRemote(&config.RemoteConfig{
            Name: "origin",
            URLs: []string{m.config.RemoteURL},
        })
        if err != nil && err != git.ErrRemoteExists {
            m.errorMsg = fmt.Sprintf("Failed to add remote origin: %v", err)
            return nil
        }

        // Set up the device-specific branch
        branchRef := plumbing.NewBranchReferenceName(m.currentBranch)
        // Create an initial empty commit
        wt, err := m.gitRepo.Worktree()
        if err != nil {
            m.errorMsg = fmt.Sprintf("Failed to get worktree: %v", err)
            return nil
        }

        // Since it's a bare repo, we need to use plumbing to create an empty commit
        // Create initial commit object
        commitObj := &object.Commit{
            Author: object.Signature{
                Name:  os.Getenv("USER"),
                Email: fmt.Sprintf("%s@localhost", os.Getenv("USER")),
                When:  time.Now(),
            },
            Committer: object.Signature{
                Name:  os.Getenv("USER"),
                Email: fmt.Sprintf("%s@localhost", os.Getenv("USER")),
                When:  time.Now(),
            },
            Message: "Initial commit",
            TreeHash: plumbing.NewHash("4b825dc642cb6eb9a060e54bf8d69288fbee4904"), // Empty tree hash
        }

        // Write the commit to the repository
        obj := m.gitRepo.Storer.NewEncodedObject()
        err = commitObj.Encode(obj)
        if err != nil {
            m.errorMsg = fmt.Sprintf("Failed to encode commit: %v", err)
            return nil
        }
        commitHash, err := m.gitRepo.Storer.SetEncodedObject(obj)
        if err != nil {
            m.errorMsg = fmt.Sprintf("Failed to store commit object: %v", err)
            return nil
        }

        // Update reference to point to the new commit
        ref := plumbing.NewHashReference(branchRef, commitHash)
        err = m.gitRepo.Storer.SetReference(ref)
        if err != nil {
            m.errorMsg = fmt.Sprintf("Failed to set branch reference: %v", err)
            return nil
        }

        // Push the branch to remote
        auth, err := sshAuth()
        if err != nil {
            m.errorMsg = fmt.Sprintf("Failed to get SSH auth: %v", err)
            return nil
        }

        err = m.gitRepo.Push(&git.PushOptions{
            Auth:       auth,
            RemoteName: "origin",
            RefSpecs:   []config.RefSpec{config.RefSpec(fmt.Sprintf("+%s:%s", branchRef, branchRef))},
        })
        if err != nil && err != git.NoErrAlreadyUpToDate {
            m.errorMsg = fmt.Sprintf("Failed to push: %v", err)
            return nil
        }

        m.errorMsg = "✔ Dotfiles repository initialized successfully."
        return nil
    }
}

// addDotfile initiates the Add Dotfile process
func (m *DotfileModel) addDotfile() tea.Cmd {
    return func() tea.Msg {
        if m.gitRepo == nil {
            m.errorMsg = "Repository not initialized. Please initialize first."
            return nil
        }

        reader := bufio.NewReader(os.Stdin)
        fmt.Print("Enter the path of the dotfile to add (relative to home): ")
        filePath, err := reader.ReadString('\n')
        if err != nil {
            m.errorMsg = fmt.Sprintf("Error reading input: %v", err)
            return nil
        }
        filePath = strings.TrimSpace(filePath)
        if filePath == "" {
            m.errorMsg = "File path cannot be empty."
            return nil
        }

        fullPath := filepath.Join(os.Getenv("HOME"), filePath)
        if _, err := os.Stat(fullPath); os.IsNotExist(err) {
            m.errorMsg = "File does not exist."
            return nil
        }

        // Stage the file
        wt, err := m.gitRepo.Worktree()
        if err != nil {
            m.errorMsg = fmt.Sprintf("Failed to get worktree: %v", err)
            return nil
        }

        // Checkout the device-specific branch
        branchRef := plumbing.NewBranchReferenceName(m.currentBranch)
        err = wt.Checkout(&git.CheckoutOptions{
            Branch: branchRef,
            Create: false,
            Force:  true,
        })
        if err != nil {
            m.errorMsg = fmt.Sprintf("Failed to checkout branch %s: %v", m.currentBranch, err)
            return nil
        }

        // Add the file
        _, err = wt.Add(filePath)
        if err != nil {
            m.errorMsg = fmt.Sprintf("Failed to add file: %v", err)
            return nil
        }

        // Commit the change
        commitMsg := fmt.Sprintf("Add %s", filePath)
        _, err = wt.Commit(commitMsg, &git.CommitOptions{
            Author: &object.Signature{
                Name:  os.Getenv("USER"),
                Email: fmt.Sprintf("%s@localhost", os.Getenv("USER")),
                When:  time.Now(),
            },
        })
        if err != nil {
            m.errorMsg = fmt.Sprintf("Failed to commit: %v", err)
            return nil
        }

        // Push the change
        auth, err := sshAuth()
        if err != nil {
            m.errorMsg = fmt.Sprintf("Failed to get SSH auth: %v", err)
            return nil
        }

        err = m.gitRepo.Push(&git.PushOptions{
            Auth:       auth,
            RemoteName: "origin",
        })
        if err != nil && err != git.NoErrAlreadyUpToDate {
            m.errorMsg = fmt.Sprintf("Failed to push: %v", err)
            return nil
        }

        m.errorMsg = fmt.Sprintf("✔ Dotfile %s added and pushed successfully.", filePath)
        return nil
    }
}

// pullDotfiles initiates the Pull Dotfiles process
func (m *DotfileModel) pullDotfiles() tea.Cmd {
    return func() tea.Msg {
        if m.gitRepo == nil {
            m.errorMsg = "Repository not initialized. Please initialize first."
            return nil
        }

        wt, err := m.gitRepo.Worktree()
        if err != nil {
            m.errorMsg = fmt.Sprintf("Failed to get worktree: %v", err)
            return nil
        }

        // Checkout the device-specific branch
        branchRef := plumbing.NewBranchReferenceName(m.currentBranch)
        err = wt.Checkout(&git.CheckoutOptions{
            Branch: branchRef,
            Create: false,
            Force:  true,
        })
        if err != nil {
            m.errorMsg = fmt.Sprintf("Failed to checkout branch %s: %v", m.currentBranch, err)
            return nil
        }

        // Pull the latest changes
        auth, err := sshAuth()
        if err != nil {
            m.errorMsg = fmt.Sprintf("Failed to get SSH auth: %v", err)
            return nil
        }

        err = wt.Pull(&git.PullOptions{
            RemoteName:    "origin",
            ReferenceName: branchRef,
            SingleBranch:  true,
            Force:         true,
            Auth:          auth,
        })
        if err != nil && err != git.NoErrAlreadyUpToDate {
            m.errorMsg = fmt.Sprintf("Failed to pull dotfiles: %v", err)
            return nil
        }

        m.errorMsg = "✔ Dotfiles pulled successfully."
        return nil
    }
}

// pullSpecificDotfile initiates the Pull Specific Dotfile process
func (m *DotfileModel) pullSpecificDotfile() tea.Cmd {
    return func() tea.Msg {
        if m.gitRepo == nil {
            m.errorMsg = "Repository not initialized. Please initialize first."
            return nil
        }

        reader := bufio.NewReader(os.Stdin)
        fmt.Print("Enter the branch name to pull from: ")
        targetBranch, err := reader.ReadString('\n')
        if err != nil {
            m.errorMsg = fmt.Sprintf("Error reading branch name: %v", err)
            return nil
        }
        targetBranch = strings.TrimSpace(targetBranch)

        fmt.Print("Enter the name of the dotfile to pull (relative to home): ")
        filePath, err := reader.ReadString('\n')
        if err != nil {
            m.errorMsg = fmt.Sprintf("Error reading file name: %v", err)
            return nil
        }
        filePath = strings.TrimSpace(filePath)
        if filePath == "" {
            m.errorMsg = "File name cannot be empty."
            return nil
        }

        // Fetch the target branch
        auth, err := sshAuth()
        if err != nil {
            m.errorMsg = fmt.Sprintf("Failed to get SSH auth: %v", err)
            return nil
        }

        err = m.gitRepo.Fetch(&git.FetchOptions{
            RemoteName: "origin",
            RefSpecs:   []config.RefSpec{config.RefSpec("+refs/heads/*:refs/remotes/origin/*")},
            Auth:       auth,
        })
        if err != nil && err != git.NoErrAlreadyUpToDate {
            m.errorMsg = fmt.Sprintf("Failed to fetch from remote: %v", err)
            return nil
        }

        // Get the commit hash of the target branch
        ref, err := m.gitRepo.Reference(plumbing.NewRemoteReferenceName("origin", targetBranch), true)
        if err != nil {
            m.errorMsg = fmt.Sprintf("Branch '%s' does not exist on remote.", targetBranch)
            return nil
        }

        // Get the commit object
        commit, err := m.gitRepo.CommitObject(ref.Hash())
        if err != nil {
            m.errorMsg = fmt.Sprintf("Failed to get commit object: %v", err)
            return nil
        }

        // Get the tree
        tree, err := commit.Tree()
        if err != nil {
            m.errorMsg = fmt.Sprintf("Failed to get tree: %v", err)
            return nil
        }

        // Find the file
        entry, err := tree.FindEntry(filePath)
        if err != nil {
            m.errorMsg = fmt.Sprintf("File '%s' not found in branch '%s'.", filePath, targetBranch)
            return nil
        }

        // Get the file blob
        blob, err := m.gitRepo.BlobObject(entry.Hash)
        if err != nil {
            m.errorMsg = fmt.Sprintf("Failed to get blob object: %v", err)
            return nil
        }

        // Read the blob content
        readerBlob, err := blob.Reader()
        if err != nil {
            m.errorMsg = fmt.Sprintf("Failed to read blob: %v", err)
            return nil
        }
        defer readerBlob.Close()

        content, err := io.ReadAll(readerBlob)
        if err != nil {
            m.errorMsg = fmt.Sprintf("Failed to read file content: %v", err)
            return nil
        }

        // Backup existing file if it exists
        destPath := filepath.Join(os.Getenv("HOME"), filePath)
        if _, err := os.Stat(destPath); err == nil {
            backupPath := fmt.Sprintf("%s.backup_%s", destPath, time.Now().Format("20060102_150405"))
            err := os.Rename(destPath, backupPath)
            if err != nil {
                m.errorMsg = fmt.Sprintf("Failed to backup existing file to %s.", backupPath)
                return nil
            }
            fmt.Printf("Existing file backed up to %s.\n", backupPath)
        }

        // Ensure destination directory exists
        destDir := filepath.Dir(destPath)
        if err := os.MkdirAll(destDir, 0755); err != nil {
            m.errorMsg = fmt.Sprintf("Failed to create directory %s.", destDir)
            return nil
        }

        // Write the file content
        err = os.WriteFile(destPath, content, 0644)
        if err != nil {
            m.errorMsg = fmt.Sprintf("Failed to write file to %s.", destPath)
            return nil
        }

        m.errorMsg = fmt.Sprintf("✔ Dotfile '%s' pulled from branch '%s'.", filePath, targetBranch)
        return nil
    }
}

// syncDotfiles initiates the Sync Dotfiles process
func (m *DotfileModel) syncDotfiles() tea.Cmd {
    return func() tea.Msg {
        if m.gitRepo == nil {
            m.errorMsg = "Repository not initialized. Please initialize first."
            return nil
        }

        wt, err := m.gitRepo.Worktree()
        if err != nil {
            m.errorMsg = fmt.Sprintf("Failed to get worktree: %v", err)
            return nil
        }

        // Checkout the device-specific branch
        branchRef := plumbing.NewBranchReferenceName(m.currentBranch)
        err = wt.Checkout(&git.CheckoutOptions{
            Branch: branchRef,
            Create: false,
            Force:  true,
        })
        if err != nil {
            m.errorMsg = fmt.Sprintf("Failed to checkout branch %s: %v", m.currentBranch, err)
            return nil
        }

        // Add all changes
        _, err = wt.Add(".")
        if err != nil {
            m.errorMsg = fmt.Sprintf("Failed to add changes: %v", err)
            return nil
        }

        // Commit changes
        commitMsg := "Sync dotfiles"
        _, err = wt.Commit(commitMsg, &git.CommitOptions{
            Author: &object.Signature{
                Name:  os.Getenv("USER"),
                Email: fmt.Sprintf("%s@localhost", os.Getenv("USER")),
                When:  time.Now(),
            },
        })
        if err != nil {
            m.errorMsg = fmt.Sprintf("Failed to commit: %v", err)
            return nil
        }

        // Push changes
        auth, err := sshAuth()
        if err != nil {
            m.errorMsg = fmt.Sprintf("Failed to get SSH auth: %v", err)
            return nil
        }

        err = m.gitRepo.Push(&git.PushOptions{
            Auth:       auth,
            RemoteName: "origin",
        })
        if err != nil && err != git.NoErrAlreadyUpToDate {
            m.errorMsg = fmt.Sprintf("Failed to push: %v", err)
            return nil
        }

        m.errorMsg = "✔ Dotfiles synced successfully."
        return nil
    }
}