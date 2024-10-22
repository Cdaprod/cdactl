package tui

import (
    tea "github.com/charmbracelet/bubbletea"
    "fmt"
    "os"
)

// MainMenuModel holds the current choice and state for the main menu
type MainMenuModel struct {
    choices  []string // Menu options
    cursor   int      // Which item the cursor is pointing at
    selected int      // The selected item
}

// NewMainMenuModel initializes the main menu options
func NewMainMenuModel() MainMenuModel {
    return MainMenuModel{
        choices: []string{"Network", "SSH", "Update", "Backup", "Monitor", "Dotfiles", "Credentials", "Exit"},
        cursor:  0,
        selected: -1,
    }
}

// Init is called when the program starts up
func (m MainMenuModel) Init() tea.Cmd {
    return nil
}

// Update handles incoming messages and updates the model
func (m MainMenuModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch msg.String() {
        case "ctrl+c", "q":
            return m, tea.Quit

        case "up":
            if m.cursor > 0 {
                m.cursor--
            }

        case "down":
            if m.cursor < len(m.choices)-1 {
                m.cursor++
            }

        case "enter":
            m.selected = m.cursor
            switch m.choices[m.selected] {
            case "Exit":
                return m, tea.Quit
            default:
                return m, handleMenuSelection(m.choices[m.selected])
            }
        }
    }

    return m, nil
}

// View renders the UI of the main menu
func (m MainMenuModel) View() string {
    s := "Welcome to the cdactl CLI Tool. Please select an option:\n\n"

    for i, choice := range m.choices {
        cursor := " " // No cursor by default
        if m.cursor == i {
            cursor = ">" // This is the cursor
        }

        s += fmt.Sprintf("%s %s\n", cursor, choice)
    }

    s += "\nPress 'q' to quit."
    return s
}

// handleMenuSelection processes the selected menu item
func handleMenuSelection(selection string) tea.Cmd {
    return func() tea.Msg {
        fmt.Printf("\nYou selected: %s\n", selection)
        switch selection {
        case "Network":
            // Call the network package handler
            fmt.Println("Calling network command...")
        case "SSH":
            // Call the SSH package handler
            fmt.Println("Calling ssh command...")
        case "Update":
            // Call the update package handler
            fmt.Println("Calling update command...")
        case "Backup":
            // Call the backup package handler
            fmt.Println("Calling backup command...")
        case "Monitor":
            // Call the monitor package handler
            fmt.Println("Calling monitor command...")
        case "Dotfiles":
            // Call the dotfiles package handler
            fmt.Println("Calling dotfile command...")
        case "Credentials":
            // Call the credentials package handler
            fmt.Println("Calling cred command...")
        default:
            fmt.Println("Invalid option.")
        }
        return nil
    }
}

// Start launches the TUI program
func Start() error {
    p := tea.NewProgram(NewMainMenuModel())
    if err := p.Start(); err != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        os.Exit(1)
    }
    return nil
}