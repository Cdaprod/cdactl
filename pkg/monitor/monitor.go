package monitor

import (
	"fmt"
	"os/exec"
)

func HandleCommand(args []string) {
	if len(args) > 0 {
		fmt.Println("Usage: cdactl monitor")
		return
	}
	monitorSystem()
}

func monitorSystem() {
	cmd := exec.Command("top", "-bn1")
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println("Error running monitor:", err)
	} else {
		fmt.Println(string(output))
	}
}

// Handler returns a placeholder message for monitor operations.
//
// Example:
//
//	msg, err := monitor.Handler()
//	if err != nil {
//	        fmt.Println(err)
//	}
//	fmt.Println(msg)
func Handler() (string, error) {
	return "monitor command not supported in TUI", nil
}
