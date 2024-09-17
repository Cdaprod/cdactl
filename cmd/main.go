package main

import (
	"fmt"
	"os"

	"github.com/Cdaprod/cdactl/pkg/backup"
	"github.com/Cdaprod/cdactl/pkg/cred"
	"github.com/Cdaprod/cdactl/pkg/dotfile"
	"github.com/Cdaprod/cdactl/pkg/monitor"
	"github.com/Cdaprod/cdactl/pkg/network"
	"github.com/Cdaprod/cdactl/pkg/record"
	"github.com/Cdaprod/cdactl/pkg/ssh"
	"github.com/Cdaprod/cdactl/pkg/update"
)

func main() {
	if len(os.Args) < 2 {
		showUsage()
		return
	}

	command := os.Args[1]
	switch command {
	case "network":
		network.HandleCommand(os.Args[2:])
	case "ssh":
		ssh.HandleCommand(os.Args[2:])
	case "update":
		update.HandleCommand(os.Args[2:])
	case "backup":
		backup.HandleCommand(os.Args[2:])
	case "monitor":
		monitor.HandleCommand(os.Args[2:])
	case "dotfile": // updated to singular form to match the package name
		dotfile.HandleCommand(os.Args[2:])
	case "cred":
		cred.HandleCommand(os.Args[2:])
	case "record": // added new case for recording functionality
		record.RunCommands()
	default:
		showUsage()
	}
}

func showUsage() {
	fmt.Println("Usage: cdactl [command] [options]")
	fmt.Println("Commands:")
	fmt.Println("  network    - Manage network connections")
	fmt.Println("  ssh        - SSH into devices")
	fmt.Println("  update     - Update system packages")
	fmt.Println("  backup     - Manage backups")
	fmt.Println("  monitor    - Monitor system resources")
	fmt.Println("  dotfile    - Manage dotfiles (get <target>)") // Updated to match singular form
	fmt.Println("  cred       - Manage credentials (store, retrieve)")
	fmt.Println("  record     - Record terminal session using Asciinema")
}