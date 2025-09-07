package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

// cdactl unlink <tool>
// Example:
//
//	cdactl unlink codex-switch
var unlinkCmd = &cobra.Command{
	Use:   "unlink <tool>",
	Short: "Remove /usr/local/bin symlink for tool",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		link := filepath.Join("/usr/local/bin", args[0])
		if err := os.Remove(link); err != nil {
			return fmt.Errorf("remove %s: %w", link, err)
		}
		fmt.Println("removed", link)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(unlinkCmd)
}
