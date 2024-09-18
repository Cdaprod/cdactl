// internal/repocate/repocate.go
package repocate

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
	"github.com/fatih/color"
	"github.com/cheggaaa/pb/v3"
	"github.com/Cdaprod/repocate/pkg/docker"
	"github.com/Cdaprod/repocate/pkg/git"
	"github.com/Cdaprod/repocate/pkg/utils"
)

var logger = utils.NewLogger()
var gitClient = git.NewGitClient(logger)
var dockerClient = docker.NewDockerClient(logger)

// rootCmd is the root command for the CLI
var rootCmd = &cobra.Command{
	Use:   "repocate",
	Short: "Repocate is a tool for managing development environments using Docker containers.",
	Long:  `Repocate allows you to clone repositories, create isolated development environments, and manage them using Docker containers.`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help() // Show help if no subcommand is provided
	},
}

// init initializes the configuration and commands
func init() {
	cobra.OnInitialize(initConfig)

	// Register commands
	rootCmd.AddCommand(startCmd)
	rootCmd.AddCommand(CreateCmd)
	rootCmd.AddCommand(EnterCmd)
	rootCmd.AddCommand(StopCmd)
	rootCmd.AddCommand(RebuildCmd)
	rootCmd.AddCommand(CloneCmd)
	rootCmd.AddCommand(ListCmd)
	rootCmd.AddCommand(VersionCmd)
	rootCmd.AddCommand(HelpCmd)
}

// initConfig loads any necessary configurations
func initConfig() {
	if err := utils.LoadConfig(""); err != nil {
		fmt.Println("Error loading config:", err)
		os.Exit(1)
	}
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() error {
	displayBanner()
	return rootCmd.Execute()
}

func displayBanner() {
	color.Cyan(`  ______                           _       
 | ___ \                         | |      
 | |_/ /___ _ __   ___   ___ __ _| |_ ___ 
 |    // _ \ '_ \ / _ \ / __/ _` + "`" + ` | __/ _ \
 | |\ \  __/ |_) | (_) | (_| (_| | ||  __/
 \_| \_\___| .__/ \___/ \___\__,_|\__\___|
           | |                            
           |_|`)

	fmt.Println(color.HiMagentaString("By: David Cannan aka Cdaprod"))
	color.Green("\nRepocate allows you to clone repositories, create isolated development environments, and manage them using Docker containers.")
	color.Blue("\n\nUsage:")
	fmt.Println("  repocate [command]")

	color.Blue("\n\nAvailable Commands:")
	displayCommands()

	color.Blue("\n\nFlags:")
	fmt.Println("  -h, --help   help for repocate")
	fmt.Println(color.GreenString("\nUse \"repocate [command] --help\" for more information about a command."))
}

func displayCommands() {
	commands := []struct {
		name        string
		description string
	}{
		{"clone", "Clone a repository"},
		{"create", "Clone a repo and create/start a development container"},
		{"enter", "Enter the development container for a specific repo"},
		{"help", "Show help information"},
		{"list", "List all repocate containers and plugins"},
		{"rebuild", "Rebuild the development container for a specific repo"},
		{"stop", "Stop the development container for a specific repo"},
		{"version", "Show version information"},
	}

	for _, cmd := range commands {
		fmt.Printf("  %-12s %s\n", cmd.name, cmd.description)
	}
}

// Commands
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Initialize and start the default Repocate container",
	Run: func(cmd *cobra.Command, args []string) {
		container.HandleDefaultContainer()
	},
}

var CreateCmd = &cobra.Command{
	Use:   "create [repoURL]",
	Short: "Clone a repo and create/start a development container",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		repoURL := args[0]
		repoName := utils.ExtractRepoName(repoURL)
		repoPath := filepath.Join(os.Getenv("HOME"), "Repocate", repoName)

		repo := &git.Repository{
			URL:    repoURL,
			Name:   repoName,
			Path:   repoPath,
			Logger: logger,
		}

		// Clone the repository
		err := gitClient.CloneRepo(repo)
		if err != nil {
			logger.LogError("Failed to clone repository.")
			os.Exit(1)
		}

		// Build Docker image
		imageTag := fmt.Sprintf("repocate/%s:latest", repoName)
		err = dockerClient.BuildImage(filepath.Join(repoPath, "Dockerfile"), imageTag)
		if err != nil {
			logger.LogError("Failed to build Docker image.")
			os.Exit(1)
		}

		// Create Docker container
		containerID, err := dockerClient.CreateContainer(
			docker.CreateContainerConfig(imageTag),
			docker.CreateHostConfig(repoPath),
			docker.CreateNetworkConfig(),
		)
		if err != nil {
			logger.LogError("Failed to create Docker container.")
			os.Exit(1)
		}

		// Start Docker container
		err = dockerClient.StartContainer(containerID)
		if err != nil {
			logger.LogError("Failed to start Docker container.")
			os.Exit(1)
		}

		// Associate container with repository
		gitClient.AssociateContainer(repo, containerID)

		logger.LogInfo(fmt.Sprintf("Repository %s is now associated with container %s", repo.Name, containerID))
	},
}

var EnterCmd = &cobra.Command{
	Use:   "enter [repoName]",
	Short: "Enter the development container for a specific repo",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		repoName := args[0]
		repoPath := filepath.Join(os.Getenv("HOME"), "Repocate", repoName)
		repo := &git.Repository{
			Name:   repoName,
			Path:   repoPath,
			Logger: logger,
		}

		// Retrieve the associated container
		containerID := gitClient.GetAssociatedContainer(repo)
		if containerID == "" {
			logger.LogError("No associated container found for this repository.")
			os.Exit(1)
		}

		// Exec into the container
		err := dockerClient.ExecCommand(containerID, []string{"zsh"})
		if err != nil {
			logger.LogError("Failed to exec into the container.")
			os.Exit(1)
		}
	},
}

var StopCmd = &cobra.Command{
	Use:   "stop [repoName]",
	Short: "Stop the development container for a specific repo",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		repoName := args[0]
		repoPath := filepath.Join(os.Getenv("HOME"), "Repocate", repoName)
		repo := &git.Repository{
			Name:   repoName,
			Path:   repoPath,
			Logger: logger,
		}

		containerID := gitClient.GetAssociatedContainer(repo)
		if containerID == "" {
			logger.LogError("No associated container found for this repository.")
			os.Exit(1)
		}

		err := dockerClient.StopContainer(containerID)
		if err != nil {
			logger.LogError("Failed to stop the container.")
			os.Exit(1)
		}

		logger.LogInfo(fmt.Sprintf("Container %s for repository %s stopped successfully.", containerID, repoName))
	},
}

var RebuildCmd = &cobra.Command{
	Use:   "rebuild [repoName]",
	Short: "Rebuild the development container for a specific repo",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		repoName := args[0]
		repoPath := filepath.Join(os.Getenv("HOME"), "Repocate", repoName)
		repo := &git.Repository{
			Name:   repoName,
			Path:   repoPath,
			Logger: logger,
		}

		// Build the Docker image again
		imageTag := fmt.Sprintf("repocate/%s:latest", repoName)
		err := dockerClient.BuildImage(filepath.Join(repoPath, "Dockerfile"), imageTag)
		if err != nil {
			logger.LogError("Failed to rebuild Docker image.")
			os.Exit(1)
		}

		// Stop existing container if running
		containerID := gitClient.GetAssociatedContainer(repo)
		if containerID != "" {
			err = dockerClient.StopContainer(containerID)
			if err != nil {
				logger.LogError("Failed to stop the existing container.")
			}
		}

	// Recreate the container
	containerID, err = dockerClient.CreateContainer(
		docker.CreateContainerConfig(imageTag),
		docker.CreateHostConfig(repoPath),
		docker.CreateNetworkConfig(),
	)
	if err != nil {
		logger.LogError("Failed to create Docker container.")
		os.Exit(1)
	}

	// Start the new container
	err = dockerClient.StartContainer(containerID)
	if err != nil {
		logger.LogError("Failed to start the container.")
		os.Exit(1)
	}

	// Associate the new container with the repository
	gitClient.AssociateContainer(repo, containerID)

	logger.LogInfo(fmt.Sprintf("Successfully rebuilt and started container for repository %s", repoName))
	},
}

var CloneCmd = &cobra.Command{
	Use:   "clone [repoURL]",
	Short: "Clone a Git repository to the local machine",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		repoURL := args[0]
		repoName := utils.ExtractRepoName(repoURL)
		repoPath := filepath.Join(os.Getenv("HOME"), "Repocate", repoName)

		repo := &git.Repository{
			URL:    repoURL,
			Name:   repoName,
			Path:   repoPath,
			Logger: logger,
		}

		// Clone the repository
		err := gitClient.CloneRepo(repo)
		if err != nil {
			logger.LogError("Failed to clone repository.")
			os.Exit(1)
		}

		logger.LogInfo(fmt.Sprintf("Repository %s successfully cloned to %s", repoName, repoPath))
	},
}

var ListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all Repocate containers and plugins",
	Run: func(cmd *cobra.Command, args []string) {
		containers, err := dockerClient.ListContainers()
		if err != nil {
			logger.LogError("Failed to list containers.")
			os.Exit(1)
		}

		if len(containers) == 0 {
			fmt.Println("No active Repocate containers found.")
		} else {
			fmt.Println("Active Repocate containers:")
			for _, container := range containers {
				fmt.Printf("  %s (ID: %s)\n", container.Names[0], container.ID)
			}
		}
	},
}

var VersionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show Repocate version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Repocate v1.0.0")
	},
}

var HelpCmd = &cobra.Command{
	Use:   "help",
	Short: "Show help information for Repocate",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

// showProgress displays a simple progress bar for ongoing tasks
func showProgress(message string, milliseconds int) {
	fmt.Print(message)
	bar := pb.New(100)
	bar.SetMaxWidth(80)
	bar.Start()
	for i := 0; i < 100; i++ {
		bar.Increment()
		time.Sleep(time.Duration(milliseconds/100) * time.Millisecond)
	}
	bar.Finish()
	fmt.Println()
}