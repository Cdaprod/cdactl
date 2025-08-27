//go:build tools

// Package tools pins tool dependencies for go module.
//
// Example:
//
//	go get github.com/charmbracelet/bubbletea@latest
//	go get github.com/charmbracelet/lipgloss@latest
//	go get github.com/charmbracelet/glamour@latest
package tools

import (
	_ "github.com/charmbracelet/bubbletea"
	_ "github.com/charmbracelet/glamour"
	_ "github.com/charmbracelet/lipgloss"
)
