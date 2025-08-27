package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestResolveBackupDir(t *testing.T) {
	custom := "/tmp/custom"
	if got := resolveBackupDir(custom); got != custom {
		t.Fatalf("expected %s, got %s", custom, got)
	}
	def := resolveBackupDir("")
	if !strings.HasSuffix(def, "/backup") {
		t.Fatalf("expected default path to end with /backup, got %s", def)
	}
}

func TestGetArchitecture(t *testing.T) {
	arch, err := getArchitecture()
	if err != nil {
		t.Fatalf("error getting architecture: %v", err)
	}
	if arch == "" {
		t.Fatalf("architecture should not be empty")
	}
}
func TestCfgGit(t *testing.T) {
	cmd := cfgGit("status")
	home := os.Getenv("HOME")
	expectedGitDir := "--git-dir=" + filepath.Join(home, ".cfg")
	expectedWorkTree := "--work-tree=" + home
	args := cmd.Args
	if len(args) < 4 || args[1] != expectedGitDir || args[2] != expectedWorkTree || args[3] != "status" {
		t.Fatalf("unexpected args: %v", args)
	}
}
