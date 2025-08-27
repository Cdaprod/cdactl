// Package tui implements terminal UI components.
package tui

import "github.com/charmbracelet/lipgloss"

// TitleStyle defines the style for headline text.
//
// Example:
//
//	s := TitleStyle.Render("cdactl")
//	fmt.Println(s)
var TitleStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205"))

// ErrorStyle highlights error messages.
//
// Example:
//
//	s := ErrorStyle.Render("oops")
//	fmt.Println(s)
var ErrorStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
