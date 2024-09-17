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