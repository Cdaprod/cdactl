// Package ui provides a Bubble Tea-based tabbed interface.
package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

// Run launches the interactive TUI.
// Example:
//
//	cdactl tui
func Run() error {
	m := NewModel()
	p := tea.NewProgram(m)
	return p.Start()
}

type tabItem struct{ title string }

func (i tabItem) Title() string       { return i.title }
func (i tabItem) Description() string { return "" }
func (i tabItem) FilterValue() string { return i.title }

type section struct {
	title     string
	items     []string
	collapsed bool
}

type model struct {
	focusLeft   bool
	tabs        list.Model
	tabSections map[string][]section
	sections    []section
	cursor      int
}

// NewModel creates the initial model with deterministic layout.
func NewModel() model {
	items := []list.Item{
		tabItem{title: "Network"},
		tabItem{title: "Backup"},
		tabItem{title: "Monitor"},
	}
	l := list.New(items, list.NewDefaultDelegate(), 20, 14)
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.SetShowHelp(false)

	tabSections := map[string][]section{
		"Network": {
			{title: "Status", items: []string{"Show IPs"}},
			{title: "Restart", items: []string{"Restart Service"}},
		},
		"Backup": {
			{title: "Create", items: []string{"Full", "Incremental"}},
			{title: "Restore", items: []string{"From File"}},
		},
		"Monitor": {
			{title: "CPU", items: []string{"Usage"}},
			{title: "Memory", items: []string{"Usage"}},
		},
	}

	return model{
		focusLeft:   true,
		tabs:        l,
		tabSections: tabSections,
		sections:    tabSections["Network"],
	}
}

func (m model) Init() tea.Cmd { return nil }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.focusLeft {
			switch msg.String() {
			case "right", "enter":
				if it, ok := m.tabs.SelectedItem().(tabItem); ok {
					m.sections = m.tabSections[it.title]
				}
				m.focusLeft = false
				m.cursor = 0
				return m, nil
			}
			var cmd tea.Cmd
			m.tabs, cmd = m.tabs.Update(msg)
			return m, cmd
		}

		switch msg.String() {
		case "left":
			m.focusLeft = true
		case "up":
			m.cursor--
			if m.cursor < 0 {
				m.cursor = len(m.sections) - 1
			}
		case "down":
			m.cursor++
			if m.cursor >= len(m.sections) {
				m.cursor = 0
			}
		case "enter", "right":
			m.sections[m.cursor].collapsed = !m.sections[m.cursor].collapsed
		}
	}
	var cmd tea.Cmd
	m.tabs, cmd = m.tabs.Update(msg)
	return m, cmd
}

func (m model) View() string {
	left := m.tabs.View()
	right := m.renderSections()
	leftLines := strings.Split(strings.TrimRight(left, "\n"), "\n")
	rightLines := strings.Split(strings.TrimRight(right, "\n"), "\n")
	h := len(leftLines)
	if len(rightLines) > h {
		h = len(rightLines)
	}
	var b strings.Builder
	for i := 0; i < h; i++ {
		l := ""
		if i < len(leftLines) {
			l = fmt.Sprintf("%-20s", leftLines[i])
		} else {
			l = fmt.Sprintf("%-20s", "")
		}
		r := ""
		if i < len(rightLines) {
			r = rightLines[i]
		}
		b.WriteString(l + r)
		if i < h-1 {
			b.WriteString("\n")
		}
	}
	return b.String()
}

func (m model) renderSections() string {
	var b strings.Builder
	for i, s := range m.sections {
		cursor := " "
		if i == m.cursor && !m.focusLeft {
			cursor = ">"
		}
		indicator := "▼"
		if s.collapsed {
			indicator = "▶"
		}
		b.WriteString(fmt.Sprintf("%s %s %s\n", cursor, indicator, s.title))
		if !s.collapsed {
			for _, item := range s.items {
				b.WriteString(fmt.Sprintf("    %s\n", item))
			}
		}
	}
	return strings.TrimRight(b.String(), "\n")
}
