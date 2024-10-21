// pkg/identifiers/identifiers.go

package identifiers

import (
    "fmt"
    "io/ioutil"
    "os"
    "os/exec"
    "path/filepath"

    "github.com/spf13/cobra"
)

// IdentifiersCmd represents the identifiers command
var IdentifiersCmd = &cobra.Command{
    Use:   "identifiers",
    Short: "Manage system-specific identifiers",
    Long:  `Handle system-specific identifiers like machine-id, SSH host keys, hostnames, and analyze disk usage.`,
}

// Subcommand: regenerate
var identifiersRegenerateCmd = &cobra.Command{
    Use:   "regenerate",
    Short: "Regenerate /etc/machine-id",
    Run: func(cmd *cobra.Command, args []string) {
        fmt.Println("=== Regenerating /etc/machine-id ===")
        cmd1 := exec.Command("sudo", "truncate", "-s", "0", "/etc/machine-id")
        cmd1.Stdout = os.Stdout
        cmd1.Stderr = os.Stderr
        if err := cmd1.Run(); err != nil {
            fmt.Println("✖ Failed to truncate /etc/machine-id:", err)
            os.Exit(1)
        }

        cmd2 := exec.Command("sudo", "systemd-machine-id-setup")
        cmd2.Stdout = os.Stdout
        cmd2.Stderr = os.Stderr
        if err := cmd2.Run(); err != nil {
            fmt.Println("✖ Failed to regenerate /etc/machine-id:", err)
            os.Exit(1)
        }

        fmt.Println("✔ /etc/machine-id regenerated successfully.")
    },
}

// Subcommand: regen-ssh
var identifiersRegenSSHCmd = &cobra.Command{
    Use:   "regen-ssh",
    Short: "Regenerate SSH host keys",
    Run: func(cmd *cobra.Command, args []string) {
        fmt.Println("=== Regenerating SSH Host Keys ===")
        cmd1 := exec.Command("sudo", "rm", "-f", "/etc/ssh/ssh_host_*")
        cmd1.Stdout = os.Stdout
        cmd1.Stderr = os.Stderr
        if err := cmd1.Run(); err != nil {
            fmt.Println("✖ Failed to remove existing SSH host keys:", err)
            os.Exit(1)
        }

        cmd2 := exec.Command("sudo", "dpkg-reconfigure", "openssh-server")
        cmd2.Stdout = os.Stdout
        cmd2.Stderr = os.Stderr
        if err := cmd2.Run(); err != nil {
            fmt.Println("✖ Failed to regenerate SSH host keys:", err)
            os.Exit(1)
        }

        fmt.Println("✔ SSH host keys regenerated successfully.")
    },
}

// Subcommand: update-hostname
var identifiersUpdateHostnameCmd = &cobra.Command{
    Use:   "update-hostname [new-hostname]",
    Short: "Update system hostname",
    Args:  cobra.ExactArgs(1),
    Run: func(cmd *cobra.Command, args []string) {
        newHostname := args[0]
        fmt.Printf("=== Updating Hostname to %s ===\n", newHostname)

        cmd1 := exec.Command("sudo", "hostnamectl", "set-hostname", newHostname)
        cmd1.Stdout = os.Stdout
        cmd1.Stderr = os.Stderr
        if err := cmd1.Run(); err != nil {
            fmt.Println("✖ Failed to set hostname:", err)
            os.Exit(1)
        }

        // Update /etc/hosts
        hostsFile := "/etc/hosts"
        contentCmd := exec.Command("grep", "-v", "127.0.1.1", hostsFile)
        content, err := contentCmd.Output()
        if err != nil {
            fmt.Println("✖ Failed to read /etc/hosts:", err)
            os.Exit(1)
        }

        newHostsContent := fmt.Sprintf("127.0.1.1\t%s\n", newHostname)
        fullContent := fmt.Sprintf("%s%s", string(content), newHostsContent)

        if err := ioutil.WriteFile(hostsFile, []byte(fullContent), 0644); err != nil {
            fmt.Println("✖ Failed to write to /etc/hosts:", err)
            os.Exit(1)
        }

        fmt.Println("✔ Hostname updated successfully in /etc/hosts.")
    },
}

// Subcommand: disk-usage
var identifiersDiskUsageCmd = &cobra.Command{
    Use:   "disk-usage",
    Short: "Analyze disk usage",
    Run: func(cmd *cobra.Command, args []string) {
        fmt.Println("=== Analyzing Disk Usage ===")
        diskUsageCmd := exec.Command("du", "-h", "--max-depth=1", "/")
        diskUsageCmd.Stdout = os.Stdout
        diskUsageCmd.Stderr = os.Stderr
        if err := diskUsageCmd.Run(); err != nil {
            fmt.Println("✖ Failed to analyze disk usage:", err)
            os.Exit(1)
        }
    },
}

func init() {
    // Add subcommands to IdentifiersCmd
    IdentifiersCmd.AddCommand(identifiersRegenerateCmd)
    IdentifiersCmd.AddCommand(identifiersRegenSSHCmd)
    IdentifiersCmd.AddCommand(identifiersUpdateHostnameCmd)
    IdentifiersCmd.AddCommand(identifiersDiskUsageCmd)
}