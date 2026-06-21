package tui

import (
	"fmt"

	"codeberg.org/malik/riscvemu/internal/core"
	tea "github.com/charmbracelet/bubbletea"
)

// Run starts the TUI and blocks until the user quits. The
// returned error is non-nil only when the underlying
// Bubble-Tea program fails; quitting via 'q' / 'esc' /
// 'ctrl+c' returns nil.
func Run(app *core.App) error {
	if app == nil {
		return fmt.Errorf("tui.Run: app is nil")
	}
	p := tea.NewProgram(NewModel(app), tea.WithAltScreen())
	_, err := p.Run()
	return err
}
