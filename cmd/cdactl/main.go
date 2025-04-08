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

var versionFlag bool

var rootCmd = &cobra.Command{
	Use:   "cdactl",
	Short: "cdactl is a CLI tool for system management.",
	Long:  `cdactl is a CLI tool to manage backups, dotfiles, network, SSH connections, and more.`,
	Run: func(cmd *cobra.Command, args []string) {
		if versionFlag {
			fmt.Println("cdactl version 1.0.0") // TODO: automate from git tag
			return
		}
		// Default behavior: show help
		_ = cmd.Help()
	},
}

func init() {
	rootCmd.AddCommand(network.NewNetworkCmd())
	rootCmd.AddCommand(ssh.NewSSHCmd())
	rootCmd.AddCommand(update.NewUpdateCmd())
	rootCmd.AddCommand(backup.NewBackupCmd())
	rootCmd.AddCommand(monitor.NewMonitorCmd())
	rootCmd.AddCommand(dotfile.NewDotfileCmd())
	rootCmd.AddCommand(cred.NewCredCmd())

	rootCmd.AddCommand(&cobra.Command{
		Use:   "tui",
		Short: "Start the Text User Interface (TUI)",
		RunE: func(cmd *cobra.Command, args []string) error {
			return tui.Start()
		},
	})

	// Attach the version flag to the root command
	rootCmd.PersistentFlags().BoolVarP(&versionFlag, "version", "v", false, "Show version info")

	cobra.OnInitialize(initConfig)
}

func initConfig() {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("$HOME/.config/cdactl/")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	} else {
		fmt.Println("Error reading config file:", err)
	}
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}