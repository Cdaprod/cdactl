package main

import (
	"bytes"
	"strings"
	"testing"
)

// TestTUIFlag verifies the --tui flag defaults to true and produces idempotent CLI output
// when disabled.
func TestTUIFlag(t *testing.T) {
	flag := rootCmd.PersistentFlags().Lookup("tui")
	if flag == nil {
		t.Fatalf("tui flag not found")
	}
	if flag.Value.String() != "true" {
		t.Fatalf("expected default true, got %s", flag.Value.String())
	}

	defer func() {
		tuiEnabled = true
		_ = flag.Value.Set("true")
		rootCmd.SetArgs(nil)
	}()

	buf1 := new(bytes.Buffer)
	rootCmd.SetOut(buf1)
	rootCmd.SetErr(buf1)
	rootCmd.SetArgs([]string{"--tui=false"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("first execute: %v", err)
	}
	out1 := buf1.String()
	if !strings.Contains(out1, "--tui") {
		t.Fatalf("expected help output, got %q", out1)
	}

	buf2 := new(bytes.Buffer)
	rootCmd.SetOut(buf2)
	rootCmd.SetErr(buf2)
	rootCmd.SetArgs([]string{"--tui=false"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("second execute: %v", err)
	}
	out2 := buf2.String()

	if out1 != out2 {
		t.Errorf("expected idempotent output")
	}
}
