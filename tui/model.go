// Package tui implements terminal UI components.
package tui

import tea "github.com/charmbracelet/bubbletea"

// Model holds the state of the terminal UI.
//
// Example:
//
//	m := NewModel("Hello")
//	p := tea.NewProgram(m)
//	_ = p.Start()
type Model struct {
	message  string
	quitting bool
}

// NewModel returns a Model with the supplied message.
func NewModel(msg string) Model {
	return Model{message: msg}
}

// Init satisfies the tea.Model interface.
func (m Model) Init() tea.Cmd {
	return nil
}

// Update handles incoming Bubble Tea messages.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.quitting = true
			return m, tea.Quit
		}
	}
	return m, nil
}
