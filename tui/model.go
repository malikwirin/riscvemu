package tui

import (
	"fmt"
	"strings"

	"codeberg.org/malik/riscvemu/internal/core"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// messageKind tags the origin of a feedback line so the View
// can render it in the right style.
type messageKind int

const (
	noMessage messageKind = iota
	infoMessage
	errorMessage
)

// feedback is a single status line that the View renders
// between the register table and the stats footer. It is
// cleared on the next user action or by pressing Esc.
type feedback struct {
	kind messageKind
	text string
}

// Model is the Bubble-Tea model that drives the TUI. It owns a
// *core.App reference and renders the live register file plus a
// footer with the current cycle count, retired count, and IPC.
type Model struct {
	app      *core.App
	width    int
	height   int
	feedback feedback
}

// NewModel returns a model that reads its state from the given
// shared core.App. The app is mutated by the model's key
// handlers (s/S/r/l/c), so the same app is observed by every
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

// WithError stores an error message that the View will render
// in the error style. Used by the l/c key handlers when a form
// reports a real failure. Returns the updated model so callers
// can chain it.
func (m Model) WithError(err error) Model {
	if err == nil {
		return m
	}
	m.feedback = feedback{kind: errorMessage, text: err.Error()}
	return m
}

// WithInfo stores a positive message that the View will render
// in the info style. Used by the s/S/r/l/c key handlers on
// success. Returns the updated model.
func (m Model) WithInfo(text string) Model {
	if text == "" {
		return m
	}
	m.feedback = feedback{kind: infoMessage, text: text}
	return m
}

// ClearFeedback drops the stored feedback line. Tied to the
// Esc key and to any successful action that wants to replace
// the previous message.
func (m Model) ClearFeedback() Model {
	m.feedback = feedback{}
	return m
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
		case "q", "ctrl+c":
			return m, tea.Quit
		case "esc":
			return m.ClearFeedback(), nil
		case "r":
			if err := m.app.Reset(); err != nil {
				return m.WithError(err), nil
			}
			return m.WithInfo("reset"), nil
		case "S":
			if err := m.app.Step(10); err != nil {
				return m.WithError(err), nil
			}
			return m.WithInfo("stepped 10 cycles"), nil
		case "s":
			if err := m.app.Step(1); err != nil {
				return m.WithError(err), nil
			}
			return m.WithInfo("stepped 1 cycle"), nil
		case "l":
			if err := LoadProgramForm(m.app); err != nil {
				if IsCancelled(err) {
					return m.WithInfo("load cancelled"), nil
				}
				return m.WithError(err), nil
			}
			return m.WithInfo("program loaded"), nil
		case "c":
			if err := ConfigForm(m.app); err != nil {
				if IsCancelled(err) {
					return m.WithInfo("config cancelled"), nil
				}
				return m.WithError(err), nil
			}
			return m.WithInfo("config applied"), nil
		}
	}
	return m, nil
}

// View renders the registers and a footer with the current
// statistics. If a feedback line is set (info or error), it is
// rendered between the register table and the footer in the
// matching style. The styling is intentionally minimal; a
// richer layout (panels, tabs) is added in a follow-up.
func (m Model) View() string {
	snap := m.app.Snapshot()
	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205"))
	bodyStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("250"))
	footerStyle := lipgloss.NewStyle().Faint(true)
	infoStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("114"))
	errorStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("196"))

	header := headerStyle.Render("riscvemu-tui")
	regs := renderRegisters(snap, bodyStyle)
	parts := []string{header, regs}
	switch m.feedback.kind {
	case infoMessage:
		parts = append(parts, infoStyle.Render("ok: "+m.feedback.text))
	case errorMessage:
		parts = append(parts, errorStyle.Render("error: "+m.feedback.text))
	}
	parts = append(parts, renderStats(snap, footerStyle))
	return strings.Join(parts, "\n")
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
		"cycles=%d  retired=%d  IPC=%.3f  (s)tep  (S)tep10  (r)eset  (l)oad  (c)onfig  (esc) clear  (q)uit",
		snap.Stats.Cycles, snap.Stats.Retired, snap.Stats.IPC(),
	)
	return style.Render(line)
}
