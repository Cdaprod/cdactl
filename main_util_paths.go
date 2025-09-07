package main

import (
	"os"
	"path/filepath"
)

// defaultToolsDir returns location of tool manifests.
// Usage:
//
//	dir := defaultToolsDir()
func defaultToolsDir() string {
	if v := os.Getenv("CDACTL_TOOLS_DIR"); v != "" {
		return v
	}
	if wd, err := os.Getwd(); err == nil {
		if st, err := os.Stat(filepath.Join(wd, "tools.d")); err == nil && st.IsDir() {
			return filepath.Join(wd, "tools.d")
		}
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "cdactl", "tools.d")
}
