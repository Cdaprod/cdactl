package main

import (
	"fmt"
	"path/filepath"

	"github.com/Cdaprod/cdactl/internal/installer"
	"github.com/spf13/cobra"
)

// cdactl list
// Example:
//
//	cdactl list
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List available tool manifests",
	RunE: func(cmd *cobra.Command, args []string) error {
		ms, err := installer.LoadAllManifests(defaultToolsDir())
		if err != nil {
			return err
		}
		if len(ms) == 0 {
			fmt.Println("no manifests found")
			return nil
		}
		fmt.Println("Available tools:")
		for _, m := range ms {
			fmt.Printf("  %s (%s/%s) asset:%s prefix:%s\n", m.Name, m.Owner, m.Repo, m.AssetTemplate, filepath.Clean(m.Prefix))
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
