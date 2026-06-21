package tui

import (
	"fmt"

	"codeberg.org/malik/riscvemu/internal/core"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Run starts the TUI and blocks until the user quits. The
// returned error is non-nil only when the underlying
// Bubble-Tea program fails; quitting via 'q' / 'esc' /
// 'ctrl+c' returns nil.
func Run(app *core.App) error {
	if app == nil {
		return fmt.Errorf("tui.Run: app is nil")
	}
	p := tea.NewProgram(InitialModel(app), tea.WithAltScreen())
	_, err := p.Run()
	return err
}

// InitialModel returns the Bubble-Tea model that drives the
// TUI. The model owns a reference to the shared core.App so
// updates and views can read live emulator state.
func InitialModel(app *core.App) initialModel {
	return initialModel{app: app}
}

// initialModel is a minimal Bubble-Tea model. It renders a
// header line and reacts to 'q', 'esc', and 'ctrl+c' to quit.
type initialModel struct {
	app    *core.App
	width  int
	height int
}

func (m initialModel) Init() tea.Cmd { return nil }

func (m initialModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc", "ctrl+c":
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m initialModel) View() string {
	style := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("205"))
	return style.Render("riscvemu-tui  (q to quit)") + "\n"
}
