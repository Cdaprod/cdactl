package tui

import "testing"

func TestUpdateSetMessage(t *testing.T) {
	m := NewModel("old")
	updated, _ := m.Update(setMessage("new"))
	if updated.(Model).message != "new" {
		t.Fatalf("expected message to update")
	}
}
