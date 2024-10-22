package main

import (
	"fmt"
	"os"

	"github.com/Cdaprod/cdactl/pkg/backup"
	"github.com/Cdaprod/cdactl/pkg/cred"
	"github.com/Cdaprod/cdactl/pkg/dotfile"
	"github.com/Cdaprod/cdactl/pkg/monitor"
	"github.com/Cdaprod/cdactl/pkg/network"
	"github.com/Cdaprod/cdactl/pkg/ssh"
	"github.com/Cdaprod/cdactl/pkg/tui"
	"github.com/Cdaprod/cdactl/pkg/update"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// Root command
var rootCmd = &cobra.Command{
	Use:   "cdactl",
	Short: "cdactl is a CLI tool for system management.",
	Long:  "cdactl is a CLI tool to manage backups, dotfiles, network, SSH connections, and more.",
}

func init() {
	// Initialize Cobra commands
	rootCmd.AddCommand(networkCmd)
	rootCmd.AddCommand(sshCmd)
	rootCmd.AddCommand(updateCmd)
	rootCmd.AddCommand(backupCmd)
	rootCmd.AddCommand(monitorCmd)
	rootCmd.AddCommand(dotfileCmd)
	rootCmd.AddCommand(credCmd)

	// Add the TUI start command
	rootCmd.AddCommand(&cobra.Command{
		Use:   "tui",
		Short: "Start the Text User Interface (TUI)",
		RunE: func(cmd *cobra.Command, args []string) error {
			return tui.Start()
		},
	})

	// Load the configuration
	cobra.OnInitialize(initConfig)
}

// initConfig reads in config file and ENV variables if set
func initConfig() {
	// Define where the config file is located
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("$HOME/.config/cdactl/") // Path to look for the config file

	// Environment variables (optional)
	viper.AutomaticEnv()

	// Load the config file
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	} else {
		fmt.Println("Error reading config file:", err)
	}
}

// Main entry point
func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}