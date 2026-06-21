package tui

import (
	"fmt"
	"strings"

	"codeberg.org/malik/riscvemu/internal/core"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Model is the Bubble-Tea model that drives the TUI. It owns a
// *core.App reference and renders the live register file plus a
// footer with the current cycle count, retired count, and IPC.
type Model struct {
	app    *core.App
	width  int
	height int
}

// NewModel returns a model that reads its state from the given
// shared core.App. The app is mutated by the model's key
// handlers (s/S/r), so the same app is observed by every
// frontend.
func NewModel(app *core.App) Model {
	return Model{app: app}
}

// Snapshot returns the current view of the emulator. It is the
// same value the model's View() reads, but exposed as a method
// so tests and helpers can inspect state without re-rendering.
func (m Model) Snapshot() core.Snapshot {
	return m.app.Snapshot()
}

func (m Model) Init() tea.Cmd { return nil }

// Update is the Bubble-Tea message handler. WindowSizeMsgs
// keep the model informed of the terminal dimensions; KeyMsgs
// drive the emulator via the shared core.App; everything else
// is ignored.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc", "ctrl+c":
			return m, tea.Quit
		case "r":
			if err := m.app.Reset(); err != nil {
				return m, errorCmd(err)
			}
			return m, nil
		case "S":
			if err := m.app.Step(10); err != nil {
				return m, errorCmd(err)
			}
			return m, nil
		case "s":
			if err := m.app.Step(1); err != nil {
				return m, errorCmd(err)
			}
			return m, nil
		}
	}
	return m, nil
}

// View renders the registers and a footer with the current
// statistics. The styling is intentionally minimal; a richer
// layout (panels, tabs) is added in a follow-up.
func (m Model) View() string {
	snap := m.app.Snapshot()
	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205"))
	bodyStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("250"))
	footerStyle := lipgloss.NewStyle().Faint(true)

	header := headerStyle.Render("riscvemu-tui")
	regs := renderRegisters(snap, bodyStyle)
	stats := renderStats(snap, footerStyle)
	return strings.Join([]string{header, regs, stats}, "\n")
}

// renderRegisters produces an aligned column of register
// values. The width is fixed at three registers per row to
// keep the layout readable in the 8-register Spec layout as
// well as the 32-register default layout.
func renderRegisters(snap core.Snapshot, style lipgloss.Style) string {
	rows := []string{}
	for i := uint32(0); i < uint32(len(snap.Registers)); i++ {
		row := fmt.Sprintf("x%d=%d", i, snap.Registers[i])
		rows = append(rows, style.Render(row))
	}
	return strings.Join(rows, "  ")
}

// renderStats produces a one-line status footer.
func renderStats(snap core.Snapshot, style lipgloss.Style) string {
	line := fmt.Sprintf(
		"cycles=%d  retired=%d  IPC=%.3f  (s)tep  (S)tep10  (r)eset  (q)uit",
		snap.Stats.Cycles, snap.Stats.Retired, snap.Stats.IPC(),
	)
	return style.Render(line)
}

// errorCmd returns a tea.Cmd that surfaces the given error in
// the next Update call via tea.ErrMsg. It exists so the key
// handlers can keep returning a tea.Cmd without inventing a
// second return value.
func errorCmd(err error) tea.Cmd {
	return func() tea.Msg { return errMsg{err} }
}

// errMsg is the tea.Msg type that the key handlers use to
// report errors. It is a private type so the package's
// external surface stays small.
type errMsg struct{ err error }

// Ensure errMsg satisfies tea.Msg at compile time.
var _ tea.Msg = errMsg{}

// Ensure Model satisfies tea.Model at compile time.
var _ tea.Model = Model{}

// table.Model is referenced only via NewModel's table.New
// below; this is the only spot in the package that imports
// bubbles/table. We keep the import minimal for the
// skeleton: the table itself is rendered with a simple
// join. Follow-up commits add the table.Model-based
// register grid.
var _ = table.New
