package tui

import (
	"fmt"
	"os"
	"strings"

	"charm.land/bubbles/v2/filepicker"
	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/table"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"codeberg.org/malik/riscvemu/arch/cpu"
	"codeberg.org/malik/riscvemu/internal/core"
	"github.com/charmbracelet/x/ansi"
)

// ansiTruncate truncates a single line to the given visual
// width using ANSI-aware truncation. The width is measured by
// the ANSI string width (ignoring escape codes).
func ansiTruncate(s string, w int) string {
	if w <= 0 {
		return s
	}
	return ansi.Truncate(s, w, "")
}

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

// ShortHelp returns the keys shown in the bottom help line.
func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Step1, k.Step10, k.Reset, k.Load, k.Config, k.Clear, k.Quit}
}

// FullHelp returns the keys shown when the user requests long
// help. Empty here because the TUI does not provide it.
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

// Model is the Bubble-Tea v2 model that drives the TUI. It owns
// a *core.App reference and renders the live register file
// (inside a scrollable viewport) plus a stats bar, a
// feedback line, and a help footer.
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
// every frontend.
func NewModel(app *core.App) Model {
	const defaultWidth, defaultHeight = 120, 30
	vp := viewport.New()
	vp.SetWidth(defaultWidth)
	vp.SetHeight(defaultHeight)
	fp := filepicker.New()
	if cwd, err := os.Getwd(); err == nil {
		fp.CurrentDirectory = cwd
	}
	fp.DirAllowed = true
	fp.FileAllowed = true
	hp := help.New()
	hp.ShortSeparator = "  "
	hp.SetWidth(defaultWidth)
	m := Model{
		app:      app,
		viewport: vp,
		picker:   fp,
		keys:     defaultKeyMap,
		help:     hp,
		width:    defaultWidth,
		height:   defaultHeight,
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
// in the error style.
func (m Model) WithError(err error) Model {
	if err == nil {
		return m
	}
	m.feedback = feedback{kind: errorMessage, text: err.Error()}
	return m
}

// WithInfo stores a positive message that the View will render
// in the info style.
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
// can change register values.
func (m Model) refreshViewport() Model {
	snap := m.app.Snapshot()
	regW, valueW := tableColumnWidths(m.width)
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
		table.WithWidth(m.width),
	)
	// bubbles/v2.1.0 has a quirk: New() calls UpdateViewport
	// before WithColumns/WithRows are applied, so the initial
	// render is empty. Force a refresh.
	t.UpdateViewport()
	m.viewport.SetContent(t.View())
	return m
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.refreshViewport().viewport.Init(),
		m.picker.Init(),
	)
}

// Update is the Bubble-Tea v2 message handler.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Always forward non-key messages (readDir, BackgroundColorMsg,
	// etc.) to the filepicker so its internal state stays current.
	if _, ok := msg.(tea.KeyPressMsg); !ok {
		var cmd tea.Cmd
		m.picker, cmd = m.picker.Update(msg)
		if cmd != nil {
			_ = cmd
		}
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		viewportHeight := msg.Height - 5
		if viewportHeight < 3 {
			viewportHeight = 3
		}
		m.viewport.SetWidth(msg.Width)
		m.viewport.SetHeight(viewportHeight)
		m.help.SetWidth(msg.Width)
		m.picker.SetHeight(viewportHeight)
		return m.refreshViewport(), nil
	case tea.KeyPressMsg:
		if m.pickerMode {
			if key.Matches(msg, m.keys.Clear) || key.Matches(msg, m.keys.Quit) {
				m.pickerMode = false
				return m.WithInfo("load cancelled"), nil
			}
			pathBefore := m.picker.Path
			var pickerCmd tea.Cmd
			m.picker, pickerCmd = m.picker.Update(msg)
			if m.picker.Path != "" && m.picker.Path != pathBefore {
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

// View renders the register viewport (or the filepicker when
// in picker mode), a stats bar, a feedback line (if any), and
// a help footer. Bubble-Tea v2 requires View() to return a
// tea.View struct, not a string.
func (m Model) View() tea.View {
	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205"))
	infoStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("114"))
	errorStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("196"))
	statsStyle := lipgloss.NewStyle().Faint(true)

	middle := m.layoutBody()
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
	parts = append(parts, renderHelp(m))
	return tea.NewView(strings.Join(parts, "\n"))
}

// renderHelp produces a help line that wraps to the terminal
// width. The bubbles/v2 help model truncates the line silently
// when its width is set; we set the width here from the
// current terminal width so the line fits the screen even on
// tests that never send a WindowSizeMsg.
func renderHelp(m Model) string {
	if m.width > 0 {
		m.help.SetWidth(m.width)
	}
	view := m.help.View(m.keys)
	if m.width > 0 {
		if wrapped := lipgloss.Wrap(view, m.width, " "); wrapped != "" {
			return wrapped
		}
	}
	return view
}

// clipToWidth truncates each line of s to the given visual
// width using ANSI-aware truncation. Multi-line strings get
// each line clipped independently; ANSI escape codes are
// preserved.
func clipToWidth(s string, w int) string {
	if w <= 0 {
		return s
	}
	lines := strings.Split(s, "\n")
	for i, line := range lines {
		lines[i] = ansiTruncate(line, w)
	}
	return strings.Join(lines, "\n")
}

// layoutBody returns the central region of the TUI. When the
// layoutBody returns the central region of the TUI. When the
// filepicker is open it takes over the whole region; otherwise
// the register and pipeline panes are arranged across the
// available width. The arrangement is chosen by splitPanes:
//
//   - 3-column (reg | alu | lsu): wide terminals only.
//   - 2-column (reg | pipeline): medium terminals.
//   - stacked (pipeline below reg): narrow terminals.
//
// Each layer is line-clipped to its pane width so the
// compositor's bounds exactly match the terminal width.
func (m Model) layoutBody() string {
	if m.pickerMode {
		return m.picker.View()
	}
	snap := m.app.Snapshot()
	regW, aluW, lsuW, three := splitPanes(m.width)

	// Narrow: stack vertically.
	if !three && lsuW == 0 && aluW == m.width {
		leftContent := clipToWidth(m.viewport.View(), m.width)
		rightContent := clipToWidth(renderPipelineView(snap, m.width), m.width)
		return leftContent + "\n" + rightContent
	}

	// Wide: 3-column layout.
	if three {
		regContent := clipToWidth(m.viewport.View(), regW)
		aluContent := clipToWidth(renderALUView(snap, aluW), aluW)
		lsuContent := clipToWidth(renderLSUAndStatusView(snap, lsuW), lsuW)
		regLayer := lipgloss.NewLayer(regContent).X(0).Y(0)
		aluLayer := lipgloss.NewLayer(aluContent).X(regW).Y(0)
		lsuLayer := lipgloss.NewLayer(lsuContent).X(regW + aluW).Y(0)
		return lipgloss.NewCompositor(regLayer, aluLayer, lsuLayer).Render()
	}

	// Medium: 2-column layout.
	leftContent := clipToWidth(m.viewport.View(), regW)
	rightContent := clipToWidth(renderPipelineView(snap, aluW), aluW)
	leftLayer := lipgloss.NewLayer(leftContent).X(0).Y(0)
	rightLayer := lipgloss.NewLayer(rightContent).X(regW).Y(0)
	return lipgloss.NewCompositor(leftLayer, rightLayer).Render()
}

// renderStatsBar produces a single-line summary of the
// emulator's current state. It surfaces three counters:
// cycles, retired, and IPC.
func renderStatsBar(snap core.Snapshot) string {
	return fmt.Sprintf(
		"cycles=%d  retired=%d  IPC=%.3f",
		snap.Stats.Cycles, snap.Stats.Retired, snap.Stats.IPC(),
	)
}

// renderPipelineView builds the right-hand side of the TUI:
// three bubbles/tables that show the state of the ALU and
// LSU reservation stations and the active rename tags on
// the architectural register file. It is a vertical stack
// of the three sub-tables, intended for terminals that
// cannot afford a three-column split.
func renderPipelineView(snap core.Snapshot, paneWidth int) string {
	sections := []string{
		renderALURSTable(snap, paneWidth),
		renderLSURSTable(snap, paneWidth),
		renderRegisterStatusTable(snap, paneWidth),
	}
	return strings.Join(sections, "\n\n")
}

// renderALUView renders the ALU reservation-station table
// alone, for the three-column layout. Its natural width
// (Slot + Busy + Op + Rd + Vj + Vk + Qj + Qk + gaps) is the
// widest of the three sub-tables; we hand it as much
// horizontal space as the layout allocates.
func renderALUView(snap core.Snapshot, paneWidth int) string {
	return renderALURSTable(snap, paneWidth)
}

// renderLSUAndStatusView renders the LSU reservation-station
// table plus the rename-tag register status. The two are
// stacked vertically inside a single column because both
// fit comfortably in the same width and the user is
// typically looking at one or the other.
func renderLSUAndStatusView(snap core.Snapshot, paneWidth int) string {
	return renderLSURSTable(snap, paneWidth) + "\n\n" + renderRegisterStatusTable(snap, paneWidth)
}

// splitPanes distributes the available terminal width across
// the register and pipeline panes. The register pane gets
// just enough room for "x31" + 32-bit value + padding; the
// remainder goes to the pipeline pane. If the terminal is
// too narrow for that minimum, both panes are clamped to
// the same width and the layout falls back to vertical
// stacking.
//
// For wide terminals the pipeline pane is split into two
// sub-panes (ALU RS, LSU RS + Reg Status) so that all three
// elements can sit side by side: register, ALU, LSU.
func splitPanes(termWidth int) (regW, aluW, lsuW int, threeColumn bool) {
	const regMin = 20 // "x31" + 11-digit value + padding
	if termWidth < 100 {
		return termWidth, termWidth, 0, false
	}
	// Try three-column: reg | alu | lsu.
	// ALU natural width: Slot(4) + Busy(4) + Op(5) + Rd(4) +
	//   Vj(6) + Vk(6) + Qj(7) + Qk(7) + 7 gaps = 50.
	// LSU natural width: Slot(4) + Busy(4) + Op(5) + Rd(4) +
	//   A(8) + Qj(7) + 5 gaps = 37.
	const aluMin = 50
	const lsuMin = 37
	if termWidth >= regMin+aluMin+lsuMin+4 { // +4 for separators
		regW = regMin
		aluW = aluMin
		lsuW = termWidth - regW - aluW
		return regW, aluW, lsuW, true
	}
	// Fall back to two-column: reg | pipeline.
	regW = regMin
	pipelineW := termWidth - regW
	if pipelineW < 60 {
		// Too narrow for both; stack vertically.
		return termWidth, termWidth, 0, false
	}
	return regW, pipelineW, 0, false
}

func tableColumnWidths(paneWidth int) (reg, value int) {
	if paneWidth < 20 {
		paneWidth = 20
	}
	reg = 4
	value = paneWidth - reg - 4
	if value < 12 {
		value = 12
	}
	return reg, value
}

func boolMark(b bool) string {
	if b {
		return "✓"
	}
	return "·"
}

func renderALURSTable(snap core.Snapshot, paneWidth int) string {
	cols := []table.Column{
		{Title: "Slot", Width: 4},
		{Title: "Busy", Width: 4},
		{Title: "Op", Width: 7},
		{Title: "Rd", Width: 4},
		{Title: "Vj", Width: 6},
		{Title: "Vk", Width: 6},
		{Title: "Qj", Width: 7},
		{Title: "Qk", Width: 7},
	}
	_ = paneWidth
	rows := make([]table.Row, len(snap.RS.ALU))
	for i, e := range snap.RS.ALU {
		rows[i] = table.Row{
			fmt.Sprintf("%d", i),
			boolMark(e.Busy),
			e.Kind.String(),
			fmt.Sprintf("x%d", e.Rd),
			fmt.Sprintf("%d", e.Vj),
			fmt.Sprintf("%d", e.Vk),
			e.Qj.String(),
			e.Qk.String(),
		}
	}
	t := table.New(
		table.WithColumns(cols),
		table.WithRows(rows),
		table.WithHeight(len(rows)+1),
		table.WithWidth(paneWidth),
	)
	t.UpdateViewport()
	return pipelineHeaderStyle().Render("ALU RS") + "\n" + t.View()
}

func renderLSURSTable(snap core.Snapshot, paneWidth int) string {
	cols := []table.Column{
		{Title: "Slot", Width: 4},
		{Title: "Busy", Width: 4},
		{Title: "Op", Width: 7},
		{Title: "Rd", Width: 4},
		{Title: "A", Width: 8},
		{Title: "Qj", Width: 7},
	}
	_ = paneWidth
	rows := make([]table.Row, len(snap.RS.LSU))
	for i, e := range snap.RS.LSU {
		rows[i] = table.Row{
			fmt.Sprintf("%d", i),
			boolMark(e.Busy),
			e.Kind.String(),
			fmt.Sprintf("x%d", e.Rd),
			fmt.Sprintf("%d", e.Imm),
			e.Qj.String(),
		}
	}
	t := table.New(
		table.WithColumns(cols),
		table.WithRows(rows),
		table.WithHeight(len(rows)+1),
		table.WithWidth(paneWidth),
	)
	t.UpdateViewport()
	return pipelineHeaderStyle().Render("LSU RS") + "\n" + t.View()
}

// renderRegisterStatusTable lists only the architectural
// registers that currently carry an active rename tag.
func renderRegisterStatusTable(snap core.Snapshot, paneWidth int) string {
	cols := []table.Column{
		// The Reg column must fit the empty-state placeholder
		// "(none)" (6 chars); without that headroom the column
		// truncates the placeholder to "(no…".
		{Title: "Reg", Width: 6},
		{Title: "Value", Width: 11},
		{Title: "Qi", Width: 9},
	}
	_ = paneWidth
	rows := []table.Row{}
	for i, q := range snap.Qi {
		if q == cpu.NoTag {
			continue
		}
		rows = append(rows, table.Row{
			fmt.Sprintf("x%d", i),
			fmt.Sprintf("%d", snap.Registers[i]),
			q.String(),
		})
	}
	if len(rows) == 0 {
		rows = []table.Row{{"(none)", "-", "-"}}
	}
	t := table.New(
		table.WithColumns(cols),
		table.WithRows(rows),
		table.WithHeight(len(rows)+1),
		table.WithWidth(paneWidth),
	)
	t.UpdateViewport()
	return pipelineHeaderStyle().Render("Reg Status") + "\n" + t.View()
}

func pipelineHeaderStyle() lipgloss.Style {
	return lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("141"))
}
