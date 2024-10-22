// pkg/dotfile/dotfile.go

package dotfile

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// DotfileConfig holds the configuration for the dotfile package
type DotfileConfig struct {
	RepoPath       string   `mapstructure:"repo_path"`
	BranchTemplate string   `mapstructure:"branch_template"`
	Excludes       []string `mapstructure:"excludes"`
}

// LoadConfig loads the dotfile configuration using Viper
func LoadConfig() (*DotfileConfig, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(filepath.Join(os.Getenv("HOME"), ".cdactl"))

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	var config DotfileConfig
	if err := viper.Unmarshal(&config); err != nil {
		return nil, err
	}

	return &config, nil
}

// DotfileCmd represents the dotfile command
var DotfileCmd = &cobra.Command{
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

	repo, err := git.PlainOpen(config.RepoPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open Git repository: %v", err)
	}

	// Determine current branch
	headRef, err := repo.Head()
	if err != nil {
		return nil, fmt.Errorf("failed to get HEAD: %v", err)
	}
	currentBranch := headRef.Name().Short()

	return &DotfileModel{
		choices:       []string{"Add Dotfile", "Pull Dotfiles", "Pull Specific Dotfile", "Sync Dotfiles", "Exit"},
		config:        config,
		gitRepo:       repo,
		currentBranch: currentBranch,
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

// addDotfile initiates the Add Dotfile process
func (m *DotfileModel) addDotfile() tea.Cmd {
	return func() tea.Msg {
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

		_, err = wt.Add(filePath)
		if err != nil {
			m.errorMsg = fmt.Sprintf("Failed to add file: %v", err)
			return nil
		}

		// Commit the change
		commitMsg := fmt.Sprintf("Add %s", filePath)
		_, err = wt.Commit(commitMsg, &git.CommitOptions{
			Author: &git.Signature{
				Name:  "cdactl",
				Email: "cdactl@example.com",
				When:  time.Now(),
			},
		})
		if err != nil {
			m.errorMsg = fmt.Sprintf("Failed to commit: %v", err)
			return nil
		}

		// Push the change
		err = m.gitRepo.Push(&git.PushOptions{})
		if err != nil {
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
		// Pull the latest changes
		err := m.gitRepo.Pull(&git.PullOptions{
			RemoteName:    "origin",
			ReferenceName: plumbing.NewBranchReferenceName(m.currentBranch),
			SingleBranch:  true,
			Force:         true,
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
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("Enter the name of the dotfile to pull (relative to home): ")
		filePath, err := reader.ReadString('\n')
		if err != nil {
			m.errorMsg = fmt.Sprintf("Error reading input: %v", err)
			return nil
		}
		filePath = strings.TrimSpace(filePath)
		if filePath == "" {
			m.errorMsg = "File name cannot be empty."
			return nil
		}

		// Checkout the specific file from the repository
		wt, err := m.gitRepo.Worktree()
		if err != nil {
			m.errorMsg = fmt.Sprintf("Failed to get worktree: %v", err)
			return nil
		}

		err = wt.Checkout(&git.CheckoutOptions{
			Branch: plumbing.NewBranchReferenceName(m.currentBranch),
			Keep:   true,
			Path:   filePath,
		})
		if err != nil {
			m.errorMsg = fmt.Sprintf("Failed to checkout file: %v", err)
			return nil
		}

		m.errorMsg = fmt.Sprintf("✔ Dotfile %s pulled successfully.", filePath)
		return nil
	}
}

// syncDotfiles initiates the Sync Dotfiles process
func (m *DotfileModel) syncDotfiles() tea.Cmd {
	return func() tea.Msg {
		wt, err := m.gitRepo.Worktree()
		if err != nil {
			m.errorMsg = fmt.Sprintf("Failed to get worktree: %v", err)
			return nil
		}

		// Add all changes
		_, err = wt.AddWithOptions(&git.AddOptions{
			All: true,
		})
		if err != nil {
			m.errorMsg = fmt.Sprintf("Failed to add changes: %v", err)
			return nil
		}

		// Commit changes
		commitMsg := "Sync dotfiles"
		_, err = wt.Commit(commitMsg, &git.CommitOptions{
			Author: &git.Signature{
				Name:  "cdactl",
				Email: "cdactl@example.com",
				When:  time.Now(),
			},
		})
		if err != nil {
			m.errorMsg = fmt.Sprintf("Failed to commit: %v", err)
			return nil
		}

		// Push changes
		err = m.gitRepo.Push(&git.PushOptions{})
		if err != nil {
			m.errorMsg = fmt.Sprintf("Failed to push: %v", err)
			return nil
		}

		m.errorMsg = "✔ Dotfiles synced successfully."
		return nil
	}
}