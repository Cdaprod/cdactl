package ui

import "testing"

func TestInitialViewDeterministic(t *testing.T) {
	m1 := NewModel()
	m2 := NewModel()
	if m1.View() != m2.View() {
		t.Fatalf("views differ\n%q\n%q", m1.View(), m2.View())
	}
}
