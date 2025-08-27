package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestModelViewContainsMessage(t *testing.T) {
	m := NewModel("test")
	if !strings.Contains(m.View(), "test") {
		t.Fatalf("expected view to contain message")
	}
}

func TestUpdateQuit(t *testing.T) {
	m := NewModel("bye")
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}}
	updated, cmd := m.Update(msg)
	if cmd == nil {
		t.Fatalf("expected quit command")
	}
	if !updated.(Model).quitting {
		t.Fatalf("expected model to be quitting")
	}
}
