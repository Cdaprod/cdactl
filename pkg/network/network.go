// pkg/network/network.go

package network

import (
	"fmt"
	"os/exec"

	"github.com/spf13/cobra"
)

func NewNetworkCmd() *cobra.Command {
	networkCmd := &cobra.Command{
		Use:   "network",
		Short: "Manage network settings",
		Long:  `Commands to manage network settings and configurations.`,
		Run: func(cmd *cobra.Command, args []string) {
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
		},
	}

	return networkCmd
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