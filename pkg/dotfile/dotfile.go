package dotfile

import (
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strings"
)

// HandleCommand processes the command-line arguments for dotfile operations
func HandleCommand(args []string) {
	if len(args) < 2 {
		fmt.Println("Usage: cdactl dotfile get <target>")
		return
	}

	switch args[0] {
	case "get":
		getDotfile(args[1])
	default:
		fmt.Println("Invalid dotfile command. Use: get <target>")
	}
}

// getDotfile retrieves the specified dotfile from the correct branch and saves it locally
func getDotfile(target string) {
	usr, _ := user.Current()
	repoDir := filepath.Join(usr.HomeDir, ".cfg")

	// Normalize target to ensure it has a dot prefix
	if !strings.HasPrefix(target, ".") {
		target = "." + target
	}

	// Ensure we are on the correct branch
	setBranchBasedOnDevice(repoDir)

	// Pull the latest changes from the correct branch
	pullLatestChanges(repoDir)

	// Retrieve the specified dotfile
	copyDotfile(repoDir, target)
}

// setBranchBasedOnDevice checks out the branch corresponding to the current device's hostname and architecture
func setBranchBasedOnDevice(repoDir string) {
	hostname, _ := os.Hostname()
	arch, err := exec.Command("uname", "-m").Output()
	if err != nil {
		fmt.Println("Error retrieving device architecture:", err)
		return
	}

	branchName := fmt.Sprintf("%s/%s", hostname, strings.TrimSpace(string(arch)))
	existingBranches, err := exec.Command("git", "--git-dir="+repoDir, "branch", "--list", branchName).Output()
	if err != nil {
		fmt.Println("Error listing branches:", err)
		return
	}

	if len(existingBranches) == 0 {
		fmt.Printf("Branch for device %s does not exist.\n", branchName)
		return
	} else {
		fmt.Printf("Switching to existing branch for device: %s\n", branchName)
		exec.Command("git", "--git-dir="+repoDir, "checkout", branchName).Run()
	}
}

// pullLatestChanges pulls the latest changes from the remote repository
func pullLatestChanges(repoDir string) {
	cmd := exec.Command("git", "--git-dir="+repoDir, "pull", "origin", getBranchName())
	err := cmd.Run()
	if err != nil {
		fmt.Println("Error pulling latest changes:", err)
	} else {
		fmt.Println("Pulled latest changes successfully.")
	}
}

// copyDotfile copies the specified dotfile from the repository to the user's home directory
func copyDotfile(repoDir, target string) {
	usr, _ := user.Current()
	sourcePath := filepath.Join(repoDir, target)
	targetPath := filepath.Join(usr.HomeDir, target)

	cmd := exec.Command("cp", sourcePath, targetPath)
	err := cmd.Run()
	if err != nil {
		fmt.Printf("Error retrieving dotfile %s: %v\n", target, err)
	} else {
		fmt.Printf("Dotfile %s retrieved and saved successfully.\n", target)
	}
}

// getBranchName returns the branch name based on the device's hostname and architecture
func getBranchName() string {
	hostname, _ := os.Hostname()
	arch, _ := exec.Command("uname", "-m").Output()
	return fmt.Sprintf("%s/%s", hostname, strings.TrimSpace(string(arch)))
}

// Handler returns a placeholder message for dotfile operations.
//
// Example:
//
//	msg, err := dotfile.Handler()
//	if err != nil {
//	        fmt.Println(err)
//	}
//	fmt.Println(msg)
func Handler() (string, error) {
	return "dotfile command not supported in TUI", nil
}
