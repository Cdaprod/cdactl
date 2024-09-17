package ssh

import (
	"fmt"
	"os"
	"os/exec"
)

func HandleCommand(args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: cdactl ssh <hostname>")
		return
	}
	hostname := args[0]
	connectToHost(hostname)
}

func connectToHost(hostname string) {
	cmd := exec.Command("ssh", hostname)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	err := cmd.Run()
	if err != nil {
		fmt.Printf("Error connecting to host %s: %v\n", hostname, err)
	}
}