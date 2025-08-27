// Package tui implements terminal UI components.
package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"

	"github.com/cdaprod/cdactl/pkg/backup"
	"github.com/cdaprod/cdactl/pkg/cred"
	"github.com/cdaprod/cdactl/pkg/dotfile"
	"github.com/cdaprod/cdactl/pkg/monitor"
	"github.com/cdaprod/cdactl/pkg/network"
	"github.com/cdaprod/cdactl/pkg/ssh"
	"github.com/cdaprod/cdactl/pkg/update"
)

// menuItem links a title to a command handler.
type menuItem struct {
	title   string
	handler func() (string, error)
}

// Model holds the state of the terminal UI.
//
// Example:
//
//	m := NewModel()
//	p := tea.NewProgram(m)
//	_ = p.Start()
type Model struct {
	items    []menuItem
	cursor   int
	output   string
	quitting bool
}

// NewModel returns a Model configured with command handlers.
func NewModel() Model {
	items := []menuItem{
		{"backup", backup.Handler},
		{"cred", cred.Handler},
		{"dotfile", dotfile.Handler},
		{"monitor", monitor.Handler},
		{"network", network.Handler},
		{"ssh", ssh.Handler},
		{"update", update.Handler},
	}
	return Model{items: items}
}

// Init satisfies the tea.Model interface.
func (m Model) Init() tea.Cmd { return nil }

// Update handles incoming Bubble Tea messages.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.quitting = true
			return m, tea.Quit
		case "up":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down":
			if m.cursor < len(m.items)-1 {
				m.cursor++
			}
		case "enter":
			out, err := m.items[m.cursor].handler()
			if err != nil {
				m.output = ErrorStyle.Render(err.Error())
				break
			}
			rendered, rerr := glamour.Render(out, "dark")
			if rerr != nil {
				m.output = out
			} else {
				m.output = rendered
			}
		}
	}
	return m, nil
}

// View renders the menu and any command output.
//
// Example:
//
//	m := NewModel()
//	fmt.Println(m.View())
func (m Model) View() string {
	if m.quitting {
		return ""
	}
	var b strings.Builder
	for i, item := range m.items {
		cursor := " "
		if i == m.cursor {
			cursor = ">"
		}
		b.WriteString(fmt.Sprintf("%s %s\n", cursor, item.title))
	}
	if m.output != "" {
		b.WriteString("\n" + m.output)
	}
	return b.String()
}
