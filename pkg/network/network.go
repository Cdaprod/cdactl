package network

import (
	"fmt"
	"os/exec"
)

func HandleCommand(args []string) {
	if len(args) < 1 {
		fmt.Println("Invalid network command. Use: status or restart")
		return
	}

	switch args[0] {
	case "status":
		showNetworkStatus()
	case "restart":
		restartNetwork()
	default:
		fmt.Println("Invalid network command. Use: status or restart")
	}
}

func showNetworkStatus() {
	cmd := exec.Command("ip", "-c", "addr", "show")
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Println(string(output))
}

func restartNetwork() {
	cmd := exec.Command("sudo", "systemctl", "restart", "NetworkManager")
	err := cmd.Run()
	if err != nil {
		fmt.Println("Error restarting network:", err)
	} else {
		fmt.Println("Network restarted successfully.")
	}
}