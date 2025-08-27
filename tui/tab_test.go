package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestTabToggle(t *testing.T) {
	m := NewModel("msg")
	if m.activeTab != 0 {
		t.Fatalf("expected initial tab 0")
	}
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyTab})
	if updated.(Model).activeTab != 1 {
		t.Fatalf("expected tab to toggle to 1")
	}
	updated, _ = updated.(Model).Update(tea.KeyMsg{Type: tea.KeyTab})
	if updated.(Model).activeTab != 0 {
		t.Fatalf("expected tab to toggle back to 0")
	}
}
