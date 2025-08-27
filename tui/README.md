# TUI

This package provides a minimal [Bubble Tea](https://github.com/charmbracelet/bubbletea) interface for invoking `cdactl` command handlers.

## Usage

```go
m := tui.NewModel()
p := tea.NewProgram(m)
_ = p.Start()
```

## Tests

```sh
go test ./...
```
