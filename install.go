package main

import (
	"context"
	"os"

	"github.com/Cdaprod/cdactl/internal/installer"
	"github.com/spf13/cobra"
)

// cdactl install <tool>
// Example:
//
//	cdactl install codex-switch
var installCmd = &cobra.Command{
	Use:   "install <tool>",
	Short: "Install a tool from manifest",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		tool := args[0]
		cfg := installer.Config{
			ToolsDir:    defaultToolsDir(),
			PrefixBase:  "/opt/cdaprod",
			BinDirLink:  "/usr/local/bin",
			Channel:     channel,
			Force:       force,
			GHToken:     os.Getenv("GITHUB_TOKEN"),
			HTTPTimeout: installer.DefaultHTTPTimeout,
		}
		return installer.Install(context.Background(), cfg, tool, version)
	},
}

var (
	version string
	channel string
	force   bool
)

func init() {
	installCmd.Flags().StringVar(&version, "version", "", "specific release tag")
	installCmd.Flags().StringVar(&channel, "channel", "stable", "release channel")
	installCmd.Flags().BoolVar(&force, "force", false, "reinstall even if present")
	rootCmd.AddCommand(installCmd)
}
