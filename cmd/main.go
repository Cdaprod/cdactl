package main

import (
	"fmt"
	"os"

	"github.com/Cdaprod/cdactl/pkg/backup"
	"github.com/Cdaprod/cdactl/pkg/cred"
	"github.com/Cdaprod/cdactl/pkg/dotfile"
	"github.com/Cdaprod/cdactl/pkg/docker"
	"github.com/Cdaprod/cdactl/pkg/git"
	"github.com/Cdaprod/cdactl/pkg/monitor"
	"github.com/Cdaprod/cdactl/pkg/network"
	"github.com/Cdaprod/cdactl/pkg/record"
	"github.com/Cdaprod/cdactl/pkg/ssh"
	"github.com/Cdaprod/cdactl/pkg/update"
	"github.com/Cdaprod/cdactl/internal/repocate"
)

var gitClient = git.NewGitClient()    // Initialize Git client
var dockerClient = docker.NewDockerClient() // Initialize Docker client

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
	case "dotfile":
		dotfile.HandleCommand(os.Args[2:])
	case "cred":
		cred.HandleCommand(os.Args[2:])
	case "record":
		record.RunCommands()
	case "git":
		handleGitCommand(os.Args[2:])
	case "docker":
		handleDockerCommand(os.Args[2:])
	case "repocate":
		repocate.Execute() // Execute repocate CLI with its subcommands
	default:
		showUsage()
	}
}

func handleGitCommand(args []string) {
	if len(args) < 1 {
		fmt.Println("git command requires an argument: [clone, commit, push, pull, etc.]")
		return
	}

	subCommand := args[0]
	switch subCommand {
	case "clone":
		if len(args) < 2 {
			fmt.Println("Usage: git clone <repoURL>")
			return
		}
		repoURL := args[1]
		gitClient.CloneRepo(repoURL, "./") // Default to current directory
	case "commit":
		if len(args) < 3 {
			fmt.Println("Usage: git commit <repoPath> <message>")
			return
		}
		repoPath := args[1]
		message := args[2]
		gitClient.CommitChanges(repoPath, message)
	case "push":
		if len(args) < 3 {
			fmt.Println("Usage: git push <repoPath> <branch>")
			return
		}
		repoPath := args[1]
		branch := args[2]
		gitClient.PushChanges(repoPath, branch)
	case "pull":
		if len(args) < 3 {
			fmt.Println("Usage: git pull <repoPath> <branch>")
			return
		}
		repoPath := args[1]
		branch := args[2]
		gitClient.PullChanges(repoPath, branch)
	default:
		fmt.Println("Unknown git subcommand:", subCommand)
	}
}

func handleDockerCommand(args []string) {
	if len(args) < 1 {
		fmt.Println("docker command requires an argument: [build, run, exec, etc.]")
		return
	}

	subCommand := args[0]
	switch subCommand {
	case "build":
		if len(args) < 3 {
			fmt.Println("Usage: docker build <pathToDockerfile> <imageTag>")
			return
		}
		pathToDockerfile := args[1]
		imageTag := args[2]
		dockerClient.BuildImage(pathToDockerfile, imageTag)
	case "run":
		if len(args) < 3 {
			fmt.Println("Usage: docker run <imageTag> <containerName>")
			return
		}
		imageTag := args[1]
		containerName := args[2]
		dockerClient.RunContainer(imageTag, containerName)
	case "exec":
		if len(args) < 3 {
			fmt.Println("Usage: docker exec <containerName> <command>")
			return
		}
		containerName := args[1]
		command := args[2]
		dockerClient.ExecCommand(containerName, command)
	case "stop":
		if len(args) < 2 {
			fmt.Println("Usage: docker stop <containerName>")
			return
		}
		containerName := args[1]
		dockerClient.StopContainer(containerName)
	default:
		fmt.Println("Unknown docker subcommand:", subCommand)
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
	fmt.Println("  dotfile    - Manage dotfiles (get <target>)")
	fmt.Println("  cred       - Manage credentials (store, retrieve)")
	fmt.Println("  record     - Record terminal session using Asciinema")
	fmt.Println("  git        - Manage Git repositories (clone, commit, push, pull)")
	fmt.Println("  docker     - Manage Docker containers (build, run, exec, stop)")
	fmt.Println("  repocate   - Clone repositories and create isolated development environments using Docker containers.")
}