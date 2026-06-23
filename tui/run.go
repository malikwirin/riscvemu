package tui

import (
	"fmt"

	tea "charm.land/bubbletea/v2"
	"codeberg.org/malik/riscvemu/internal/core"
)

// Run starts the TUI and blocks until the user quits. The
// returned error is non-nil only when the underlying
// Bubble-Tea program fails; quitting via 'q' / 'esc' /
// 'ctrl+c' returns nil.
func Run(app *core.App) error {
	if app == nil {
		return fmt.Errorf("tui.Run: app is nil")
	}
	p := tea.NewProgram(NewModel(app))
	if _, err := p.Run(); err != nil {
		return err
	}
	return nil
}
