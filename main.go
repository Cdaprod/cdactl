package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/term"

	"github.com/spf13/cobra"
)

// Color definitions using ANSI escape codes
const (
	RED    = "\033[0;31m"
	GREEN  = "\033[0;32m"
	YELLOW = "\033[1;33m"
	BLUE   = "\033[0;34m"
	NC     = "\033[0m" // No Color
)

// Helper functions for colored output
func printHeader(text string) {
	fmt.Printf("%s=== %s ===%s\n", BLUE, text, NC)
}

func printSuccess(text string) {
	fmt.Printf("%s✔ %s%s\n", GREEN, text, NC)
}

func printError(text string) {
	fmt.Printf("%s✖ %s%s\n", RED, text, NC)
}

func printWarning(text string) {
	fmt.Printf("%s⚠ %s%s\n", YELLOW, text, NC)
}

// Execute a shell command and return its output and error status
func executeCommand(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	cmd.Stderr = os.Stderr
	output, err := cmd.Output()
	return string(output), err
}

// Read password securely without echoing
func readPassword(prompt string) (string, error) {
	fmt.Printf("%s: ", prompt)
	bytePassword, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Println()
	if err != nil {
		return "", err
	}
	return string(bytePassword), nil
}

// Get the device's hostname
func getHostname() (string, error) {
	hostname, err := os.Hostname()
	if err != nil {
		return "", err
	}
	return hostname, nil
}

// Get the device's architecture
func getArchitecture() (string, error) {
	out, err := exec.Command("uname", "-m").Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

// Get the branch name based on hostname and architecture
func getBranchName() (string, error) {
	hostname, err := getHostname()
	if err != nil {
		return "", err
	}
	arch, err := getArchitecture()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/%s", hostname, arch), nil
}

// resolveBackupDir returns the directory to use for backups
func resolveBackupDir(dir string) string {
	if dir != "" {
		return dir
	}
	return filepath.Join(os.Getenv("HOME"), "backup")
}

// cfgGit returns a git command scoped to the ~/.cfg repository
func cfgGit(args ...string) *exec.Cmd {
	cfgDir := filepath.Join(os.Getenv("HOME"), ".cfg")
	base := []string{"--git-dir=" + cfgDir, "--work-tree=" + os.Getenv("HOME")}
	return exec.Command("git", append(base, args...)...)
}

// ensureCfgRepo clones the dotfile repository if it does not exist
func ensureCfgRepo() error {
	cfgDir := filepath.Join(os.Getenv("HOME"), ".cfg")
	if _, err := os.Stat(cfgDir); os.IsNotExist(err) {
		cmd := exec.Command("git", "clone", "--bare", "https://github.com/Cdaprod/cda.cfg.git", cfgDir)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	}
	return nil
}

// Root command
var rootCmd = &cobra.Command{
	Use:   "cdactl",
	Short: "cdactl is a CLI tool to manage system configurations and dotfiles.",
	Long: `cdactl is a comprehensive command-line tool designed to manage
system configurations, network connections, SSH into devices, update packages,
manage backups, monitor system resources, handle dotfiles, and manage credentials.`,
}

// Network command
var networkCmd = &cobra.Command{
	Use:   "network",
	Short: "Manage network connections",
	Long:  `Manage network connections, including viewing status and restarting the network manager.`,
}

var networkStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "View network status",
	Run: func(cmd *cobra.Command, args []string) {
		printHeader("Network Status")
		output, err := executeCommand("ip", "-c", "addr", "show")
		if err != nil {
			printError("Failed to retrieve network status.")
			os.Exit(1)
		}
		fmt.Println(output)
	},
}

var networkRestartCmd = &cobra.Command{
	Use:   "restart",
	Short: "Restart NetworkManager",
	Run: func(cmd *cobra.Command, args []string) {
		printHeader("Restarting Network")
		_, err := executeCommand("sudo", "systemctl", "restart", "NetworkManager")
		if err != nil {
			printError("Network restart failed.")
			os.Exit(1)
		}
		printSuccess("Network restart completed successfully.")
	},
}

// SSH command
var sshCmd = &cobra.Command{
	Use:   "ssh [hostname]",
	Short: "SSH into devices",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		hostname := args[0]
		printHeader(fmt.Sprintf("Connecting to device: %s", hostname))
		execCmd := exec.Command("ssh", hostname)
		execCmd.Stdin = os.Stdin
		execCmd.Stdout = os.Stdout
		execCmd.Stderr = os.Stderr
		if err := execCmd.Run(); err != nil {
			printError(fmt.Sprintf("SSH connection to %s failed.", hostname))
			os.Exit(1)
		}
	},
}

// Update command
var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update system packages",
	Run: func(cmd *cobra.Command, args []string) {
		printHeader("Updating System Packages")
		// Execute 'sudo apt update && sudo apt upgrade -y'
		updateCmd := exec.Command("sudo", "apt", "update")
		updateCmd.Stdout = os.Stdout
		updateCmd.Stderr = os.Stderr
		if err := updateCmd.Run(); err != nil {
			printError("Failed to update package lists.")
			os.Exit(1)
		}

		upgradeCmd := exec.Command("sudo", "apt", "upgrade", "-y")
		upgradeCmd.Stdout = os.Stdout
		upgradeCmd.Stderr = os.Stderr
		if err := upgradeCmd.Run(); err != nil {
			printError("Failed to upgrade packages.")
			os.Exit(1)
		}

		printSuccess("System packages updated successfully.")
	},
}

// Backup command
var backupCmd = &cobra.Command{
	Use:   "backup",
	Short: "Manage backups",
	Long:  `Create and restore system backups.`,
}
var backupDir string

func init() {
	backupCmd.PersistentFlags().StringVar(&backupDir, "dir", "", "backup directory (default $HOME/backup)")
}

var backupCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a backup",
	Run: func(cmd *cobra.Command, args []string) {
		dir := resolveBackupDir(backupDir)
		printHeader("Creating Backup")
		// Create backup directory
		if err := os.MkdirAll(dir, 0755); err != nil {
			printError("Failed to create backup directory.")
			os.Exit(1)
		}

		// Create tar.gz archive excluding 'minio-persistent-data'
		backupFile := fmt.Sprintf("backup_%s.tar.gz", time.Now().Format("20060102"))
		backupPath := filepath.Join(dir, backupFile)
		tarCmd := exec.Command("tar", "--exclude='minio-persistent-data'", "-czvf", backupPath, os.Getenv("HOME"))
		tarCmd.Stdout = os.Stdout
		tarCmd.Stderr = os.Stderr
		if err := tarCmd.Run(); err != nil {
			printError("Backup creation failed.")
			os.Exit(1)
		}

		printSuccess(fmt.Sprintf("Backup created at %s", backupPath))
	},
}

var backupRestoreCmd = &cobra.Command{
	Use:   "restore [backup_file]",
	Short: "Restore from a backup",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		dir := resolveBackupDir(backupDir)
		backupFile := args[0]
		backupPath := filepath.Join(dir, backupFile)

		// Check if backup file exists
		if _, err := os.Stat(backupPath); os.IsNotExist(err) {
			printError(fmt.Sprintf("Backup file not found: %s", backupFile))
			os.Exit(1)
		}

		printHeader("Restoring from Backup")
		tarCmd := exec.Command("tar", "-xzvf", backupPath, "-C", os.Getenv("HOME"))
		tarCmd.Stdout = os.Stdout
		tarCmd.Stderr = os.Stderr
		if err := tarCmd.Run(); err != nil {
			printError("Backup restoration failed.")
			os.Exit(1)
		}

		printSuccess("Backup restored successfully.")
	},
}

// Monitor command
var monitorCmd = &cobra.Command{
	Use:   "monitor",
	Short: "Monitor system resources",
	Run: func(cmd *cobra.Command, args []string) {
		printHeader("System Resource Monitor")
		topCmd := exec.Command("top", "-bn1")
		pipe, err := topCmd.StdoutPipe()
		if err != nil {
			printError("Failed to execute top command.")
			os.Exit(1)
		}

		if err := topCmd.Start(); err != nil {
			printError("Failed to start top command.")
			os.Exit(1)
		}

		scanner := bufio.NewScanner(pipe)
		lineCount := 0
		for scanner.Scan() {
			fmt.Println(scanner.Text())
			lineCount++
			if lineCount >= 20 {
				break
			}
		}

		if err := topCmd.Wait(); err != nil {
			printError("Top command failed.")
			os.Exit(1)
		}
	},
}

// Dotfiles command and its subcommands
var dotfilesCmd = &cobra.Command{
	Use:   "dotfiles",
	Short: "Manage dotfiles",
	Long:  `Initialize, add, pull, pull specific files, and sync dotfiles across devices.`,
}

var dotfilesInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize dotfiles repository",
	Run: func(cmd *cobra.Command, args []string) {
		printHeader("Initializing Dotfiles Repository")
		if err := ensureCfgRepo(); err != nil {
			printError("Failed to ensure dotfile repository.")
			os.Exit(1)
		}

		bashrcPath := filepath.Join(os.Getenv("HOME"), ".bashrc")
		aliasLine := "alias cfg='/usr/bin/git --git-dir=$HOME/.cfg/ --work-tree=$HOME'\n"
		data, _ := os.ReadFile(bashrcPath)
		if !strings.Contains(string(data), aliasLine) {
			file, err := os.OpenFile(bashrcPath, os.O_APPEND|os.O_WRONLY, 0644)
			if err == nil {
				defer file.Close()
				file.WriteString(aliasLine)
				printSuccess("Alias 'cfg' added to ~/.bashrc. Please run 'source ~/.bashrc' to apply changes.")
			}
		}

		printSuccess("Dotfiles repository ready.")
	},
}

var dotfilesAddCmd = &cobra.Command{
	Use:   "add [file]",
	Short: "Add a dotfile",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		filePath := args[0]
		if err := ensureCfgRepo(); err != nil {
			printError("Failed to ensure dotfile repository.")
			os.Exit(1)
		}
		branchName, err := getBranchName()
		if err != nil {
			printError("Failed to get branch name.")
			os.Exit(1)
		}

		// Switch to device-specific branch
		gitCheckout := cfgGit("checkout", branchName)
		gitCheckout.Stdout = os.Stdout
		gitCheckout.Stderr = os.Stderr
		if err := gitCheckout.Run(); err != nil {
			// If branch doesn't exist remotely, create it
			gitCheckout = cfgGit("checkout", "-b", branchName)
			gitCheckout.Stdout = os.Stdout
			gitCheckout.Stderr = os.Stderr
			if err := gitCheckout.Run(); err != nil {
				printError(fmt.Sprintf("Failed to checkout or create branch %s.", branchName))
				os.Exit(1)
			}

			// Make an initial empty commit
			gitCommit := cfgGit("commit", "--allow-empty", "-m", fmt.Sprintf("Initial commit on %s", branchName))
			gitCommit.Stdout = os.Stdout
			gitCommit.Stderr = os.Stderr
			if err := gitCommit.Run(); err != nil {
				printError(fmt.Sprintf("Failed to make initial commit on %s branch.", branchName))
				os.Exit(1)
			}

			// Push branch to remote
			gitPush := cfgGit("push", "--set-upstream", "origin", branchName)
			gitPush.Stdout = os.Stdout
			gitPush.Stderr = os.Stderr
			if err := gitPush.Run(); err != nil {
				printError(fmt.Sprintf("Failed to push branch %s to remote.", branchName))
				os.Exit(1)
			}
		}

		// Add the specified file
		gitAdd := cfgGit("add", filePath)
		gitAdd.Stdout = os.Stdout
		gitAdd.Stderr = os.Stderr
		if err := gitAdd.Run(); err != nil {
			printError(fmt.Sprintf("Failed to add file %s.", filePath))
			os.Exit(1)
		}

		// Commit the changes
		gitCommit := cfgGit("commit", "-m", fmt.Sprintf("Add %s", filePath))
		gitCommit.Stdout = os.Stdout
		gitCommit.Stderr = os.Stderr
		if err := gitCommit.Run(); err != nil {
			printError(fmt.Sprintf("Failed to commit file %s.", filePath))
			os.Exit(1)
		}

		// Push the changes
		gitPush := cfgGit("push", "origin", branchName)
		gitPush.Stdout = os.Stdout
		gitPush.Stderr = os.Stderr
		if err := gitPush.Run(); err != nil {
			printError(fmt.Sprintf("Failed to push changes to branch %s.", branchName))
			os.Exit(1)
		}

		printSuccess(fmt.Sprintf("Dotfile %s added and pushed to %s branch.", filePath, branchName))
	},
}

var dotfilesPullCmd = &cobra.Command{
	Use:   "pull",
	Short: "Pull dotfiles from remote",
	Run: func(cmd *cobra.Command, args []string) {
		if err := ensureCfgRepo(); err != nil {
			printError("Failed to ensure dotfile repository.")
			os.Exit(1)
		}
		branchName, err := getBranchName()
		if err != nil {
			printError("Failed to get branch name.")
			os.Exit(1)
		}

		// Switch to device-specific branch
		gitCheckout := cfgGit("checkout", branchName)
		gitCheckout.Stdout = os.Stdout
		gitCheckout.Stderr = os.Stderr
		if err := gitCheckout.Run(); err != nil {
			printError(fmt.Sprintf("Failed to checkout branch %s.", branchName))
			os.Exit(1)
		}

		// Pull from remote
		gitPull := cfgGit("pull", "origin", branchName)
		gitPull.Stdout = os.Stdout
		gitPull.Stderr = os.Stderr
		if err := gitPull.Run(); err != nil {
			printError(fmt.Sprintf("Failed to pull from branch %s.", branchName))
			os.Exit(1)
		}

		printSuccess(fmt.Sprintf("Dotfiles pulled from branch %s.", branchName))
	},
}

var dotfilesPullFileCmd = &cobra.Command{
	Use:   "pull-file [branch] [filename.ext]",
	Short: "Pull a specific dotfile from a branch",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		targetBranch := args[0]
		filename := args[1]

		if err := ensureCfgRepo(); err != nil {
			printError("Failed to ensure dotfile repository.")
			os.Exit(1)
		}

		printHeader(fmt.Sprintf("Pulling '%s' from branch '%s'", filename, targetBranch))

		gitFetch := cfgGit("fetch", "origin")
		gitFetch.Stdout = os.Stdout
		gitFetch.Stderr = os.Stderr
		if err := gitFetch.Run(); err != nil {
			printError("Failed to fetch from remote.")
			os.Exit(1)
		}

		gitLsRemote := cfgGit("ls-remote", "--heads", "origin", targetBranch)
		output, err := gitLsRemote.Output()
		if err != nil {
			printError(fmt.Sprintf("Failed to check remote branches: %v", err))
			os.Exit(1)
		}
		if !strings.Contains(string(output), targetBranch) {
			printError(fmt.Sprintf("Branch '%s' does not exist on remote.", targetBranch))
			os.Exit(1)
		}

		gitLsTree := cfgGit("ls-tree", "-r", "--name-only", targetBranch)
		treeOutput, err := gitLsTree.Output()
		if err != nil {
			printError("Failed to list files in the branch.")
			os.Exit(1)
		}

		lines := strings.Split(string(treeOutput), "\n")
		var matches []string
		for _, line := range lines {
			if filepath.Base(line) == filename {
				matches = append(matches, line)
			}
		}

		if len(matches) == 0 {
			printError(fmt.Sprintf("File '%s' not found in branch '%s'.", filename, targetBranch))
			os.Exit(1)
		} else if len(matches) > 1 {
			fmt.Printf("%sMultiple matches found for '%s' in branch '%s':%s\n", YELLOW, filename, targetBranch, NC)
			for i, match := range matches {
				fmt.Printf("  %d. %s\n", i+1, match)
			}
			printError("Please specify the exact path or rename your file to avoid ambiguity.")
			os.Exit(1)
		}

		filePath := matches[0]
		destPath := filepath.Join(os.Getenv("HOME"), filePath)

		destDir := filepath.Dir(destPath)
		if err := os.MkdirAll(destDir, 0755); err != nil {
			printError(fmt.Sprintf("Failed to create directory %s.", destDir))
			os.Exit(1)
		}

		if _, err := os.Stat(destPath); err == nil {
			backupPath := fmt.Sprintf("%s.backup_%s", destPath, time.Now().Format("20060102_150405"))
			if err := os.Rename(destPath, backupPath); err != nil {
				printError(fmt.Sprintf("Failed to backup existing file to %s.", backupPath))
				os.Exit(1)
			}
			printWarning(fmt.Sprintf("Existing file backed up to %s.", backupPath))
		}

		gitShow := cfgGit("show", fmt.Sprintf("%s:%s", targetBranch, filePath))
		fileContent, err := gitShow.Output()
		if err != nil {
			printError(fmt.Sprintf("Failed to pull '%s' from branch '%s'.", filePath, targetBranch))
			os.Exit(1)
		}

		if err := os.WriteFile(destPath, fileContent, 0644); err != nil {
			printError(fmt.Sprintf("Failed to write file to %s.", destPath))
			os.Exit(1)
		}

		printSuccess(fmt.Sprintf("Pulled '%s' from branch '%s' to '%s'.", filePath, targetBranch, destPath))
	},
}

var dotfilesSyncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Sync dotfiles with remote",
	Run: func(cmd *cobra.Command, args []string) {
		if err := ensureCfgRepo(); err != nil {
			printError("Failed to ensure dotfile repository.")
			os.Exit(1)
		}
		branchName, err := getBranchName()
		if err != nil {
			printError("Failed to get branch name.")
			os.Exit(1)
		}

		// Switch to device-specific branch
		gitCheckout := cfgGit("checkout", branchName)
		gitCheckout.Stdout = os.Stdout
		gitCheckout.Stderr = os.Stderr
		if err := gitCheckout.Run(); err != nil {
			printError(fmt.Sprintf("Failed to checkout branch %s.", branchName))
			os.Exit(1)
		}

		// Add all changes
		gitAdd := cfgGit("add", "-A")
		gitAdd.Stdout = os.Stdout
		gitAdd.Stderr = os.Stderr
		if err := gitAdd.Run(); err != nil {
			printError("Failed to add changes.")
			os.Exit(1)
		}

		// Commit changes
		gitCommit := cfgGit("commit", "-m", "Sync dotfiles")
		gitCommit.Stdout = os.Stdout
		gitCommit.Stderr = os.Stderr
		if err := gitCommit.Run(); err != nil {
			printError("Failed to commit changes.")
			os.Exit(1)
		}

		// Push changes
		gitPush := cfgGit("push", "origin", branchName)
		gitPush.Stdout = os.Stdout
		gitPush.Stderr = os.Stderr
		if err := gitPush.Run(); err != nil {
			printError(fmt.Sprintf("Failed to push changes to branch %s.", branchName))
			os.Exit(1)
		}

		printSuccess(fmt.Sprintf("Dotfiles synced with branch %s.", branchName))
	},
}

// Cred command and its subcommands
var credCmd = &cobra.Command{
	Use:   "cred",
	Short: "Manage credentials",
	Long:  `Store and retrieve credentials securely.`,
}

var credStoreCmd = &cobra.Command{
	Use:   "store [service] [username]",
	Short: "Store credentials",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		service := args[0]
		username := args[1]

		password, err := readPassword("Input Secure Key")
		if err != nil {
			printError("Failed to read password.")
			os.Exit(1)
		}

		passphrase := os.Getenv("GPG_PASSPHRASE")
		if passphrase == "" {
			passphrase, err = readPassword("GPG Passphrase")
			if err != nil {
				printError("Failed to read passphrase.")
				os.Exit(1)
			}
		}

		credPath := filepath.Join(os.Getenv("HOME"), ".credentials.gpg")
		credContent := fmt.Sprintf("service=%s\nusername=%s\npassword=%s\n", service, username, password)

		gpgCmd := exec.Command("gpg", "--batch", "--yes", "--passphrase", passphrase, "-c", "-o", credPath, "-")
		gpgCmd.Stdin = strings.NewReader(credContent)
		gpgCmd.Stdout = os.Stdout
		gpgCmd.Stderr = os.Stderr
		if err := gpgCmd.Run(); err != nil {
			printError("Failed to encrypt credentials.")
			os.Exit(1)
		}

		printSuccess(fmt.Sprintf("Credentials for %s stored securely.", service))
	},
}

var credRetrieveCmd = &cobra.Command{
	Use:   "retrieve",
	Short: "Retrieve stored credentials",
	Run: func(cmd *cobra.Command, args []string) {
		passphrase := os.Getenv("GPG_PASSPHRASE")
		var err error
		if passphrase == "" {
			passphrase, err = readPassword("GPG Passphrase")
			if err != nil {
				printError("Failed to read passphrase.")
				os.Exit(1)
			}
		}

		credPath := filepath.Join(os.Getenv("HOME"), ".credentials.gpg")
		printHeader("Retrieving Stored Credentials")

		gpgCmd := exec.Command("gpg", "--batch", "--yes", "--passphrase", passphrase, "-d", credPath)
		output, err := gpgCmd.Output()
		if err != nil {
			if _, statErr := os.Stat(credPath); os.IsNotExist(statErr) {
				printWarning("No credentials stored.")
				return
			}
			printError("Failed to decrypt credentials.")
			os.Exit(1)
		}

		fmt.Println(string(output))
	},
}

func init() {
	// Add subcommands to network
	networkCmd.AddCommand(networkStatusCmd)
	networkCmd.AddCommand(networkRestartCmd)

	// Add subcommands to backup
	backupCmd.AddCommand(backupCreateCmd)
	backupCmd.AddCommand(backupRestoreCmd)

	// Add subcommands to dotfiles
	dotfilesCmd.AddCommand(dotfilesInitCmd)
	dotfilesCmd.AddCommand(dotfilesAddCmd)
	dotfilesCmd.AddCommand(dotfilesPullCmd)
	dotfilesCmd.AddCommand(dotfilesPullFileCmd)
	dotfilesCmd.AddCommand(dotfilesSyncCmd)

	// Add subcommands to cred
	credCmd.AddCommand(credStoreCmd)
	credCmd.AddCommand(credRetrieveCmd)

	// Add all main commands to root
	rootCmd.AddCommand(networkCmd)
	rootCmd.AddCommand(sshCmd)
	rootCmd.AddCommand(updateCmd)
	rootCmd.AddCommand(backupCmd)
	rootCmd.AddCommand(monitorCmd)
	rootCmd.AddCommand(dotfilesCmd)
	rootCmd.AddCommand(credCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		printError(err.Error())
		os.Exit(1)
	}
}
