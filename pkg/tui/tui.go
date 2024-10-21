// pkg/tui/tui.go

package tui

import (
    tea "github.com/charmbracelet/bubbletea"
)

func Start() error {
    p := tea.NewProgram(NewMainMenuModel())
    return p.Start()
}