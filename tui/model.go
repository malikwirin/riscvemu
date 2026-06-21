package tui

import (
	"fmt"
	"strings"

	"codeberg.org/malik/riscvemu/internal/core"
	"github.com/charmbracelet/bubbles/table"
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

// tableColumnWidths returns the (Reg, Value) column widths
// for the register table given a terminal width. The total
// table width is at most the terminal width minus a small
// margin; if the terminal is too narrow we fall back to
// the minimum widths that still let the headers fit.
func tableColumnWidths(termWidth int) (reg, value int) {
	const margin = 4
	const minReg = 4
	const minValue = 12
	if termWidth <= 0 {
		termWidth = 80
	}
	available := termWidth - margin
	if available < minReg+minValue {
		available = minReg + minValue
	}
	reg = minReg
	value = available - reg
	if value < minValue {
		value = minValue
	}
	return reg, value
}

// newRegisterTable builds a bubbles/table.Model with one row
// per architectural register. The columns are "Reg" and
// "Value"; widths scale with the terminal width so the table
// fits the user's screen.
func newRegisterTable(snap core.Snapshot, termWidth int) table.Model {
	regW, valueW := tableColumnWidths(termWidth)
	cols := []table.Column{
		{Title: "Reg", Width: regW},
		{Title: "Value", Width: valueW},
	}
	rows := make([]table.Row, len(snap.Registers))
	for i, v := range snap.Registers {
		rows[i] = table.Row{
			fmt.Sprintf("x%d", i),
			fmt.Sprintf("%d", v),
		}
	}
	t := table.New(
		table.WithColumns(cols),
		table.WithRows(rows),
		table.WithHeight(len(rows)+1),
	)
	return t
}

// refreshTable rebuilds the table rows from the current
// emulator snapshot. Called after every action that can
// change register values (step, reset, load, config).
func (m Model) refreshTable() Model {
	snap := m.app.Snapshot()
	rows := make([]table.Row, len(snap.Registers))
	for i, v := range snap.Registers {
		rows[i] = table.Row{
			fmt.Sprintf("x%d", i),
			fmt.Sprintf("%d", v),
		}
	}
	m.regTable.SetRows(rows)
	return m
}

// Model is the Bubble-Tea model that drives the TUI. It owns a
// *core.App reference and renders the live register file plus a
// footer with the current cycle count, retired count, and IPC.
type Model struct {
	app      *core.App
	width    int
	height   int
	feedback feedback
	regTable table.Model
}

// NewModel returns a model that reads its state from the given
// shared core.App. The app is mutated by the model's key
// handlers (s/S/r/l/c), so the same app is observed by every
// frontend.
func NewModel(app *core.App) Model {
	return Model{app: app, regTable: newRegisterTable(app.Snapshot(), 0)}
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
		_, valueW := tableColumnWidths(msg.Width)
		m.regTable.SetWidth(msg.Width)
		_ = valueW
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
			return m.refreshTable().WithInfo("reset"), nil
		case "S":
			if err := m.app.Step(10); err != nil {
				return m.WithError(err), nil
			}
			return m.refreshTable().WithInfo("stepped 10 cycles"), nil
		case "s":
			if err := m.app.Step(1); err != nil {
				return m.WithError(err), nil
			}
			return m.refreshTable().WithInfo("stepped 1 cycle"), nil
		case "l":
			if err := LoadProgramForm(m.app); err != nil {
				if IsCancelled(err) {
					return m.WithInfo("load cancelled"), nil
				}
				return m.WithError(err), nil
			}
			return m.refreshTable().WithInfo("program loaded"), nil
		case "c":
			if err := ConfigForm(m.app); err != nil {
				if IsCancelled(err) {
					return m.WithInfo("config cancelled"), nil
				}
				return m.WithError(err), nil
			}
			return m.refreshTable().WithInfo("config applied"), nil
		}
	}
	return m, nil
}

// View renders the registers as a bubbles/table, plus a
// feedback line (if any) and a stats footer. The footer
// wraps onto multiple lines when the terminal is narrow
// so the user never has to scroll horizontally.
func (m Model) View() string {
	snap := m.app.Snapshot()
	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205"))
	infoStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("114"))
	errorStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("196"))
	footerStyle := lipgloss.NewStyle().Faint(true)

	parts := []string{
		headerStyle.Render("riscvemu-tui"),
		m.regTable.View(),
	}
	switch m.feedback.kind {
	case infoMessage:
		parts = append(parts, infoStyle.Render("ok: "+m.feedback.text))
	case errorMessage:
		parts = append(parts, errorStyle.Render("error: "+m.feedback.text))
	}
	parts = append(parts, renderStats(snap, footerStyle, m.width))
	return strings.Join(parts, "\n")
}

// renderStats produces a one- or two-line status footer. The
// line is wrapped on word boundaries if it would otherwise
// exceed the terminal width. Stats and key hints share the
// same logical line so the wrap reads naturally.
func renderStats(snap core.Snapshot, style lipgloss.Style, termWidth int) string {
	if termWidth <= 0 {
		termWidth = 80
	}
	line := fmt.Sprintf(
		"cycles=%d  retired=%d  IPC=%.3f  (s)tep  (S)tep10  (r)eset  (l)oad  (c)onfig  (esc) clear  (q)uit",
		snap.Stats.Cycles, snap.Stats.Retired, snap.Stats.IPC(),
	)
	if len(line) <= termWidth {
		return style.Render(line)
	}
	return style.Render(strings.Join(wrapLine(line, termWidth), "\n"))
}

// wrapLine splits s on word boundaries so each chunk fits
// within w characters. The cut respects whitespace; a word
// longer than w is hard-broken to avoid infinite loops.
func wrapLine(s string, w int) []string {
	if w <= 0 || len(s) <= w {
		return []string{s}
	}
	var out []string
	for len(s) > w {
		cut := strings.LastIndex(s[:w], " ")
		if cut <= 0 {
			cut = w
		}
		out = append(out, strings.TrimRight(s[:cut], " "))
		s = strings.TrimLeft(s[cut:], " ")
	}
	out = append(out, s)
	return out
}
