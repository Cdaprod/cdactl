package backup

import (
	"fmt"
	"os"
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

// cmd/backup/backup.go (continued)

var backupCronCmd = &cobra.Command{
	Use:   "cron",
	Short: "Backup scheduled tasks (cron jobs)",
	Run: func(cmd *cobra.Command, args []string) {
		backupDir := filepath.Join(os.Getenv("HOME"), "backup", "cron")
		fmt.Println("=== Backing Up Cron Jobs ===")
		if err := os.MkdirAll(backupDir, 0755); err != nil {
			fmt.Println("✖ Failed to create backup directory for cron jobs.")
			os.Exit(1)
		}

		// Backup user cron jobs
		userCronFile := filepath.Join(backupDir, "user_cron_jobs.txt")
		cmdUserCron := exec.Command("crontab", "-l")
		userCronOutput, err := cmdUserCron.Output()
		if err == nil {
			if err := ioutil.WriteFile(userCronFile, userCronOutput, 0644); err != nil {
				fmt.Println("✖ Failed to write user cron jobs.")
				os.Exit(1)
			}
			fmt.Println("✔ User cron jobs backed up successfully.")
		} else {
			fmt.Println("⚠ No user cron jobs found or failed to retrieve.")
		}

		// Backup root cron jobs
		rootCronFile := filepath.Join(backupDir, "root_cron_jobs.txt")
		cmdRootCron := exec.Command("crontab", "-u", "root", "-l")
		rootCronOutput, err := cmdRootCron.Output()
		if err == nil {
			if err := ioutil.WriteFile(rootCronFile, rootCronOutput, 0644); err != nil {
				fmt.Println("✖ Failed to write root cron jobs.")
				os.Exit(1)
			}
			fmt.Println("✔ Root cron jobs backed up successfully.")
		} else {
			fmt.Println("⚠ No root cron jobs found or failed to retrieve.")
		}
	},
}

func init() {
	BackupCmd.AddCommand(backupCreateCmd)
	BackupCmd.AddCommand(backupRestoreCmd)
	BackupCmd.AddCommand(backupLogsCmd)
	BackupCmd.AddCommand(backupCronCmd)
}