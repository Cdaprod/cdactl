package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

// cdactl which <tool>
// Example:
//
//	cdactl which codex-switch
var whichCmd = &cobra.Command{
	Use:   "which <tool>",
	Short: "Show symlink target for a tool",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		link := filepath.Join("/usr/local/bin", args[0])
		target, err := os.Readlink(link)
		if err != nil {
			return fmt.Errorf("%s: %w", link, err)
		}
		fmt.Printf("%s -> %s\n", link, target)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(whichCmd)
}
