// Package tui implements terminal UI components.
package tui

// View renders the Model using predefined styles.
//
// Example:
//
//	m := NewModel("Hello")
//	fmt.Println(m.View())
func (m Model) View() string {
	if m.quitting {
		return ""
	}
	return TitleStyle.Render(m.message)
}
