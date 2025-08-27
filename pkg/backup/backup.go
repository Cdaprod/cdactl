package backup

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"time"
)

const backupDir = "/home/user/backup" // Adjust path as needed

func HandleCommand(args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: cdactl backup <create|restore> [backup_file]")
		return
	}

	switch args[0] {
	case "create":
		createBackup()
	case "restore":
		if len(args) < 2 {
			fmt.Println("Usage: cdactl backup restore <backup_file>")
			return
		}
		restoreBackup(args[1])
	default:
		fmt.Println("Invalid backup command. Use: create or restore")
	}
}

func createBackup() {
	timestamp := time.Now().Format("20060102")
	backupFile := filepath.Join(backupDir, fmt.Sprintf("backup_%s.tar.gz", timestamp))

	cmd := exec.Command("tar", "-czvf", backupFile, "--exclude='minio-persistent-data'", "/home/user") // Adjust path as needed
	err := cmd.Run()
	if err != nil {
		fmt.Println("Error creating backup:", err)
	} else {
		fmt.Printf("Backup created successfully at %s\n", backupFile)
	}
}

func restoreBackup(backupFile string) {
	fullPath := filepath.Join(backupDir, backupFile)
	cmd := exec.Command("tar", "-xzvf", fullPath, "-C", "/home/user") // Adjust path as needed
	err := cmd.Run()
	if err != nil {
		fmt.Println("Error restoring backup:", err)
	} else {
		fmt.Println("Backup restored successfully.")
	}
}

// Handler returns a placeholder message for backup operations.
//
// Example:
//
//	msg, err := backup.Handler()
//	if err != nil {
//	        fmt.Println(err)
//	}
//	fmt.Println(msg)
func Handler() (string, error) {
	return "backup command not supported in TUI", nil
}
