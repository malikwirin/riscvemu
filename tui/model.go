package tui

import (
	"fmt"
	"os"
	"strings"

	"codeberg.org/malik/riscvemu/internal/core"
	"github.com/charmbracelet/bubbles/filepicker"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/viewport"
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
// between the register viewport and the help footer. It is
// cleared on the next user action or by pressing Esc.
type feedback struct {
	kind messageKind
	text string
}

// keyMap defines every keyboard binding the TUI responds to.
// The fields are used by both Update (for matching) and help
// (for rendering the help line). Centralising them here
// keeps the key list and the help text in sync.
type keyMap struct {
	Step1  key.Binding
	Step10 key.Binding
	Reset  key.Binding
	Load   key.Binding
	Config key.Binding
	Clear  key.Binding
	Quit   key.Binding
	Up     key.Binding
	Down   key.Binding
}

// ShortHelp returns the keys shown in the bottom help line
// when the viewport is not focused. It implements the
// help.KeyMap interface.
func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Step1, k.Step10, k.Reset, k.Load, k.Config, k.Clear, k.Quit}
}

// FullHelp returns the keys shown when the user requests the
// long help (e.g. with `?`). Empty here because the TUI does
// not provide a long-help mode.
func (k keyMap) FullHelp() [][]key.Binding {
	return nil
}

var defaultKeyMap = keyMap{
	Step1:  key.NewBinding(key.WithKeys("s"), key.WithHelp("s", "step 1")),
	Step10: key.NewBinding(key.WithKeys("S"), key.WithHelp("S", "step 10")),
	Reset:  key.NewBinding(key.WithKeys("r"), key.WithHelp("r", "reset")),
	Load:   key.NewBinding(key.WithKeys("l"), key.WithHelp("l", "load")),
	Config: key.NewBinding(key.WithKeys("c"), key.WithHelp("c", "config")),
	Clear:  key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "clear")),
	Quit:   key.NewBinding(key.WithKeys("q", "ctrl+c"), key.WithHelp("q", "quit")),
	Up:     key.NewBinding(key.WithKeys("up"), key.WithHelp("↑", "up")),
	Down:   key.NewBinding(key.WithKeys("down"), key.WithHelp("↓", "down")),
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

// renderStatsBar produces a single-line summary of the
// emulator's current state. It surfaces three counters:
// cycles (total CPU cycles, reset by Reset/Load/Config),
// retired (instructions retired), and IPC
// (retired/cycles). The cycles counter doubles as the
// step counter the user sees, because every s/S keypress
// advances the CPU by the same number of cycles. The line
// is intentionally compact so it never wraps in normal
// terminal widths.
func renderStatsBar(snap core.Snapshot) string {
	return fmt.Sprintf(
		"cycles=%d  retired=%d  IPC=%.3f",
		snap.Stats.Cycles, snap.Stats.Retired, snap.Stats.IPC(),
	)
}

// buildRegisterRows builds the full set of register rows
// from a snapshot. The number of rows is the architectural
// register count, not the visible viewport height; the
// viewport scrolls if the terminal cannot show them all.
func buildRegisterRows(snap core.Snapshot) []table.Row {
	rows := make([]table.Row, len(snap.Registers))
	for i, v := range snap.Registers {
		rows[i] = table.Row{
			fmt.Sprintf("x%d", i),
			fmt.Sprintf("%d", v),
		}
	}
	return rows
}

// Model is the Bubble-Tea model that drives the TUI. It owns
// a *core.App reference and renders the live register file
// (inside a scrollable viewport) plus a stats bar, a
// feedback line, and a help footer. When pickerMode is true,
// the filepicker is shown instead of the register viewport,
// so the user can navigate the file system to pick a
// program.
type Model struct {
	app        *core.App
	width      int
	height     int
	feedback   feedback
	viewport   viewport.Model
	picker     filepicker.Model
	pickerMode bool
	keys       keyMap
	help       help.Model
}

// NewModel returns a model that reads its state from the given
// shared core.App. The app is mutated by the model's key
// handlers (s/S/r/l/c/up/down), so the same app is observed by
// every frontend. The viewport is pre-filled so callers that
// invoke View() before Update (i.e. without a Bubble-Tea
// Init) still see the register table. The filepicker starts
// in the current working directory so the user finds local
// assembly files without navigating from $HOME, and
// DirAllowed is true so the user can open sub-directories.
func NewModel(app *core.App) Model {
	vp := viewport.New(80, 20)
	fp := filepicker.New()
	if cwd, err := os.Getwd(); err == nil {
		fp.CurrentDirectory = cwd
	}
	fp.DirAllowed = true
	fp.FileAllowed = true
	fp.ShowHidden = false
	hp := help.New()
	hp.ShortSeparator = "  "
	m := Model{
		app:      app,
		viewport: vp,
		picker:   fp,
		keys:     defaultKeyMap,
		help:     hp,
	}
	return m.refreshViewport()
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

// refreshViewport rebuilds the viewport content from the
// current emulator snapshot. Called after every action that
// can change register values (step, reset, load, config).
func (m Model) refreshViewport() Model {
	snap := m.app.Snapshot()
	regW, valueW := tableColumnWidths(m.width)
	cols := []table.Column{
		{Title: "Reg", Width: regW},
		{Title: "Value", Width: valueW},
	}
	t := table.New(
		table.WithColumns(cols),
		table.WithRows(buildRegisterRows(snap)),
		table.WithHeight(len(snap.Registers)+1),
	)
	m.viewport.SetContent(t.View())
	return m
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.refreshViewport().viewport.Init(),
		m.picker.Init(),
	)
}

// Update is the Bubble-Tea message handler. WindowSizeMsgs
// keep the model informed of the terminal dimensions; KeyMsgs
// drive the emulator via the shared core.App; everything else
// is ignored. When pickerMode is on, all keys are forwarded
// to the filepicker until it reports a file selection
// (directories are navigated into, not loaded).
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Always forward the filepicker's own readDir messages
	// (produced by its Init) so the picker's file list
	// stays current even when the user has not opened the
	// picker.
	if _, ok := msg.(tea.KeyMsg); !ok {
		var cmd tea.Cmd
		m.picker, cmd = m.picker.Update(msg)
		if cmd != nil {
			// Defer the cmd; the picker is alive but idle.
			// (tea.Batch would be cleaner but the
			// short-circuits below are easier to read.)
		}
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		// Reserve space: header (1) + feedback (1) + help (1)
		// + margin (2). The viewport gets whatever is left.
		viewportHeight := msg.Height - 5
		if viewportHeight < 3 {
			viewportHeight = 3
		}
		m.viewport.Width = msg.Width
		m.viewport.Height = viewportHeight
		m.help.Width = msg.Width
		m.picker.SetHeight(viewportHeight)
		return m.refreshViewport(), nil
	case tea.KeyMsg:
		// Picker mode: forward every key to the filepicker
		// until it reports a file selection.
		if m.pickerMode {
			if key.Matches(msg, m.keys.Clear) || key.Matches(msg, m.keys.Quit) {
				m.pickerMode = false
				return m.WithInfo("load cancelled"), nil
			}
			// Snapshot the path BEFORE Update; the filepicker
			// only sets Path on Select, so a change means the
			// user picked a file (not a directory).
			pathBefore := m.picker.Path
			var pickerCmd tea.Cmd
			m.picker, pickerCmd = m.picker.Update(msg)
			if m.picker.Path != "" && m.picker.Path != pathBefore {
				// Selection completed; refuse directories.
				if info, err := os.Stat(m.picker.Path); err == nil && info.IsDir() {
					return m, pickerCmd
				}
				path := m.picker.Path
				m.pickerMode = false
				if err := m.app.LoadProgramFromFile(path); err != nil {
					return m.WithError(err), nil
				}
				return m.refreshViewport().WithInfo("program loaded"), nil
			}
			return m, pickerCmd
		}
		// Arrow keys scroll the register viewport; we forward
		// the key before any other handler so the viewport
		// sees it even when the user is mid-action.
		switch msg.String() {
		case "up", "down", "pgup", "pgdown", "home", "end":
			var cmd tea.Cmd
			m.viewport, cmd = m.viewport.Update(msg)
			return m, cmd
		}
		switch {
		case key.Matches(msg, m.keys.Quit):
			return m, tea.Quit
		case key.Matches(msg, m.keys.Clear):
			return m.ClearFeedback(), nil
		case key.Matches(msg, m.keys.Reset):
			if err := m.app.Reset(); err != nil {
				return m.WithError(err), nil
			}
			return m.refreshViewport().WithInfo("reset"), nil
		case key.Matches(msg, m.keys.Step10):
			if err := m.app.Step(10); err != nil {
				return m.WithError(err), nil
			}
			return m.refreshViewport().WithInfo("stepped 10 cycles"), nil
		case key.Matches(msg, m.keys.Step1):
			if err := m.app.Step(1); err != nil {
				return m.WithError(err), nil
			}
			return m.refreshViewport().WithInfo("stepped 1 cycle"), nil
		case key.Matches(msg, m.keys.Load):
			m.pickerMode = true
			return m.WithInfo("pick a file (enter to select)"), nil
		case key.Matches(msg, m.keys.Config):
			if err := ConfigForm(m.app); err != nil {
				if IsCancelled(err) {
					return m.WithInfo("config cancelled"), nil
				}
				return m.WithError(err), nil
			}
			return m.refreshViewport().WithInfo("config applied"), nil
		}
	}
	return m, nil
}

// View renders the register viewport (or the filepicker
// when in picker mode), a feedback line (if any), and a help
// footer. The bubbles/viewport clips the register table to
// the available height and the bubbles/help component
// renders the key hints as a single wrap-aware line, so the
// View never grows beyond the terminal and the help line
// never duplicates.
func (m Model) View() string {
	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205"))
	infoStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("114"))
	errorStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("196"))
	statsStyle := lipgloss.NewStyle().Faint(true)

	middle := m.viewport.View()
	if m.pickerMode {
		middle = m.picker.View()
	}
	parts := []string{
		headerStyle.Render("riscvemu-tui"),
		middle,
		statsStyle.Render(renderStatsBar(m.app.Snapshot())),
	}
	switch m.feedback.kind {
	case infoMessage:
		parts = append(parts, infoStyle.Render("ok: "+m.feedback.text))
	case errorMessage:
		parts = append(parts, errorStyle.Render("error: "+m.feedback.text))
	}
	parts = append(parts, m.help.View(m.keys))
	return strings.Join(parts, "\n")
}
