package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

// cdactl remove <tool>
// Example:
//
//	cdactl remove codex-switch --all
var (
	removeAll bool
)

var removeCmd = &cobra.Command{
	Use:   "remove <tool>",
	Short: "Remove installed versions for a tool",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		prefix := filepath.Join("/opt/cdaprod", args[0])
		if _, err := os.Stat(prefix); err != nil {
			return fmt.Errorf("no install tree at %s", prefix)
		}
		if removeAll {
			return os.RemoveAll(prefix)
		}
		entries, err := os.ReadDir(prefix)
		if err != nil {
			return err
		}
		fmt.Println("existing versions in", prefix)
		for _, e := range entries {
			if e.IsDir() {
				fmt.Println("  ", e.Name())
			}
		}
		fmt.Println("use --all to remove all versions")
		return nil
	},
}

func init() {
	removeCmd.Flags().BoolVar(&removeAll, "all", false, "delete all versions")
	rootCmd.AddCommand(removeCmd)
}
