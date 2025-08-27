package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestHandlerInvocation(t *testing.T) {
	m := NewModel()
	m.items[0].handler = func() (string, error) { return "hello", nil }
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if !strings.Contains(updated.(Model).View(), "hello") {
		t.Fatalf("expected handler output in view")
	}
}

func TestUpdateQuit(t *testing.T) {
	m := NewModel()
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}}
	updated, cmd := m.Update(msg)
	if cmd == nil {
		t.Fatalf("expected quit command")
	}
	if !updated.(Model).quitting {
		t.Fatalf("expected model to be quitting")
	}
}
