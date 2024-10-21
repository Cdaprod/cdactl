// pkg/tui/main_menu.go

package tui

import (
    "fmt"

    "github.com/Cdaprod/cdactl/pkg/backup"
    "github.com/Cdaprod/cdactl/pkg/cred"
    "github.com/Cdaprod/cdactl/pkg/dotfiles"
    "github.com/Cdaprod/cdactl/pkg/identifiers"
    "github.com/Cdaprod/cdactl/pkg/monitor"
    "github.com/Cdaprod/cdactl/pkg/network"
    "github.com/Cdaprod/cdactl/pkg/ssh"
    "github.com/Cdaprod/cdactl/pkg/update"
    tea "github.com/charmbracelet/bubbletea"
    "github.com/spf13/cobra"
)

type MainMenuModel struct {
    cursor   int
    menus    []string
    quitting bool
}

func NewMainMenuModel() MainMenuModel {
    return MainMenuModel{
        menus: []string{
            "Network Management",
            "SSH Management",
            "Update System Packages",
            "Backup Management",
            "System Monitoring",
            "Dotfiles Management",
            "Credential Management",
            "System Identifiers",
            "Exit",
        },
        cursor: 0,
    }
}