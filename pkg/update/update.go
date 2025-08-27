package update

import (
	"fmt"
	"os/exec"
)

func HandleCommand(args []string) {
	if len(args) > 0 {
		fmt.Println("Usage: cdactl update")
		return
	}
	updateSystem()
}

func updateSystem() {
	fmt.Println("Updating system packages...")
	cmd := exec.Command("sudo", "apt", "update")
	cmd.Stdout = exec.Command("sudo", "apt", "upgrade", "-y").Stdout
	err := cmd.Run()
	if err != nil {
		fmt.Println("Error updating system:", err)
	} else {
		fmt.Println("System updated successfully.")
	}
}

// Handler returns a placeholder message for system updates.
//
// Example:
//
//	msg, err := update.Handler()
//	if err != nil {
//	        fmt.Println(err)
//	}
//	fmt.Println(msg)
func Handler() (string, error) {
	return "update command not supported in TUI", nil
}
