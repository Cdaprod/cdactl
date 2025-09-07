package main

import (
	"context"
	"os"

	"github.com/Cdaprod/cdactl/internal/installer"
	"github.com/spf13/cobra"
)

// cdactl update <tool>
// Example:
//
//	cdactl update codex-switch --channel canary
var updateCmd = &cobra.Command{
	Use:   "update <tool>",
	Short: "Update a tool to latest or specific version",
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

func init() {
	updateCmd.Flags().StringVar(&version, "version", "", "explicit release tag")
	updateCmd.Flags().StringVar(&channel, "channel", "stable", "release channel")
	updateCmd.Flags().BoolVar(&force, "force", false, "reinstall even if present")
	rootCmd.AddCommand(updateCmd)
}
