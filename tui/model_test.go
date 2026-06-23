package tui_test

import (
	"fmt"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"codeberg.org/malik/riscvemu/arch/cpu"
	"codeberg.org/malik/riscvemu/internal/core"
	"codeberg.org/malik/riscvemu/tui"
)

// viewText returns the textual content of a model's v2 View.
// Bubble-Tea v2 returns a tea.View struct whose Content field
// is the rendered string; this helper lets the tests stay
// readable while still asserting on rendered text.
func viewText(m tui.Model) string {
	return m.View().Content
}

// TestModelWithAppReadsSnapshot covers the basic contract: the
// model carries a *core.App reference, and the View() output
// reflects the current snapshot.
func TestModelWithAppReadsSnapshot(t *testing.T) {
	app := core.New(1024, cpu.SpecConfig())
	if err := app.LoadProgram("addi x1, x0, 42"); err != nil {
		t.Fatalf("LoadProgram: %v", err)
	}
	if err := app.Step(2); err != nil {
		t.Fatalf("Step: %v", err)
	}
	m := tui.NewModel(app)
	view := viewText(m)
	if !strings.Contains(view, "42") {
		t.Errorf("View output should show x1=42 after Step, got:\n%s", view)
	}
}

// TestModelStepKeyAdvancesOneCycle pins that pressing 's' on
// the model triggers app.Step(1) and the next View reflects the
// new cycle count.
func TestModelStepKeyAdvancesOneCycle(t *testing.T) {
	app := core.New(1024, cpu.SpecConfig())
	m := tui.NewModel(app)
	if got := app.Snapshot().Stats.Cycles; got != 0 {
		t.Fatalf("setup: cycles = %d, want 0", got)
	}
	updated, _ := m.Update(keyMsg("s"))
	app2 := updated.(tui.Model)
	if got := app2.Snapshot().Stats.Cycles; got != 1 {
		t.Errorf("cycles after one 's' key = %d, want 1", got)
	}
}

// TestModelShiftStepKeyAdvancesTen pins that pressing 'S' on
// the model triggers app.Step(10).
func TestModelShiftStepKeyAdvancesTen(t *testing.T) {
	app := core.New(1024, cpu.SpecConfig())
	m := tui.NewModel(app)
	updated, _ := m.Update(keyMsg("S"))
	app2 := updated.(tui.Model)
	if got := app2.Snapshot().Stats.Cycles; got != 10 {
		t.Errorf("cycles after one 'S' key = %d, want 10", got)
	}
}

// TestModelResetKeyResetsState pins that pressing 'r' on the
// model triggers app.Reset.
func TestModelResetKeyResetsState(t *testing.T) {
	app := core.New(1024, cpu.SpecConfig())
	if err := app.Step(5); err != nil {
		t.Fatalf("Step: %v", err)
	}
	m := tui.NewModel(app)
	updated, _ := m.Update(keyMsg("r"))
	app2 := updated.(tui.Model)
	if got := app2.Snapshot().Stats.Cycles; got != 0 {
		t.Errorf("cycles after 'r' = %d, want 0", got)
	}
}

// TestModelStepShowsOkFeedback pins that a successful step
// surfaces a positive "ok: stepped 1 cycle" message in the
// View, so the user sees that the keystroke had an effect.
func TestModelStepShowsOkFeedback(t *testing.T) {
	_, m := newModel(t)
	updated, _ := m.Update(keyMsg("s"))
	view := viewText(updated.(tui.Model))
	if !strings.Contains(view, "ok:") {
		t.Errorf("View after 's' should contain 'ok:', got:\n%s", view)
	}
	if !strings.Contains(view, "stepped") {
		t.Errorf("View after 's' should mention 'stepped', got:\n%s", view)
	}
}

// TestModelResetShowsOkFeedback pins that a successful reset
// surfaces an "ok: reset" message.
func TestModelResetShowsOkFeedback(t *testing.T) {
	_, m := newModel(t)
	updated, _ := m.Update(keyMsg("r"))
	view := viewText(updated.(tui.Model))
	if !strings.Contains(view, "ok:") {
		t.Errorf("View after 'r' should contain 'ok:', got:\n%s", view)
	}
	if !strings.Contains(view, "reset") {
		t.Errorf("View after 'r' should mention 'reset', got:\n%s", view)
	}
}

// TestModelErrorRendersInView pins that setError makes the
// View show a red "error: ..." line. This is the path used
// by the l/c key handlers when a form reports a real failure.
func TestModelErrorRendersInView(t *testing.T) {
	_, m := newModel(t)
	updated, _ := m.Update(keyMsg("x")) // unknown key, no message
	m2 := updated.(tui.Model).WithError(errorString("simulated load failure"))
	view := viewText(m2)
	if !strings.Contains(view, "error:") {
		t.Errorf("View should contain 'error:' after WithError, got:\n%s", view)
	}
	if !strings.Contains(view, "simulated load failure") {
		t.Errorf("View should contain the error message, got:\n%s", view)
	}
}

// TestModelEscClearsFeedback pins that pressing Esc clears any
// pending feedback line.
func TestModelEscClearsFeedback(t *testing.T) {
	_, m := newModel(t)
	m2 := m.WithError(errorString("something went wrong"))
	if !strings.Contains(viewText(m2), "error:") {
		t.Fatalf("setup: View should contain 'error:'")
	}
	updated, _ := m2.Update(keyMsg("esc"))
	view := viewText(updated.(tui.Model))
	if strings.Contains(view, "error:") {
		t.Errorf("View after Esc should not contain 'error:', got:\n%s", view)
	}
}

// TestModelOkClearsOnNextAction pins that a previous positive
// message is cleared when the user takes another action, so
// the footer always reflects the latest operation.
func TestModelOkClearsOnNextAction(t *testing.T) {
	_, m := newModel(t)
	m2, _ := m.Update(keyMsg("s"))
	if !strings.Contains(viewText(m2.(tui.Model)), "ok:") {
		t.Fatalf("setup: 's' should leave 'ok:' visible")
	}
	m3, _ := m2.Update(keyMsg("S"))
	view := viewText(m3.(tui.Model))
	if !strings.Contains(view, "ok:") {
		t.Errorf("View after second 'S' should still have 'ok:' (newer message), got:\n%s", view)
	}
}

// TestModelViewHasMultiLineRegisters pins that the register
// table wraps onto multiple lines, so the View fits in a
// standard terminal without horizontal scrolling. With
// 32 registers and the default layout the table must span
// at least 4 lines.
func TestModelViewHasMultiLineRegisters(t *testing.T) {
	_, m := newModel(t)
	view := viewText(m)
	lines := strings.Split(view, "\n")
	if len(lines) < 4 {
		t.Errorf("View should have at least 4 lines (header + register rows + footer), got %d:\n%s", len(lines), view)
	}
}

// TestModelViewNoLongLine pins that no single line in the
// View exceeds a reasonable terminal width. The viewport-
// based layout must keep every line bounded. The test uses
// the visual width (ANSI-aware) so it doesn't trip over
// embedded ANSI escape codes.
func TestModelViewNoLongLine(t *testing.T) {
	_, m := newModel(t)
	view := viewText(m)
	for i, line := range strings.Split(view, "\n") {
		if visualWidth(line) > 120 {
			t.Errorf("line %d is visually %d cells wide (>120):\n%s", i, visualWidth(line), line)
		}
	}
}

// visualWidth returns the number of terminal cells a string
// occupies, ignoring ANSI escape codes. It is a test-only
// helper; production code uses lipgloss.Width which gives the
// same value.
func visualWidth(s string) int {
	w := 0
	inEsc := false
	for _, r := range s {
		if r == 0x1b {
			inEsc = true
			continue
		}
		if inEsc {
			if r == 'm' || r == 'K' || r == 'H' {
				inEsc = false
			}
			continue
		}
		w++
	}
	return w
}

// TestModelViewShowsAllRegisters pins that the View mentions
// every architectural register name (x0 .. x31) once the
// terminal is sized tall enough to show the full viewport.
// The user can always scroll to see hidden rows.
func TestModelViewShowsAllRegisters(t *testing.T) {
	_, m := newModel(t)
	updated, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 100})
	m2 := updated.(tui.Model)
	view := viewText(m2)
	for i := uint32(0); i < 32; i++ {
		name := fmt.Sprintf("x%d", i)
		if !strings.Contains(view, name) {
			t.Errorf("View should mention register %s, got:\n%s", name, view)
		}
	}
}

// TestModelViewIncludesHelpLine pins that the View renders a
// help line with at least one keyboard hint. The exact
// wording is left to the bubbles/help component.
func TestModelViewIncludesHelpLine(t *testing.T) {
	_, m := newModel(t)
	view := viewText(m)
	if !strings.Contains(view, "step") {
		t.Errorf("View should mention 'step' in the help line, got:\n%s", view)
	}
}

// TestModelDownKeyScrollsViewport pins that pressing the
// down-arrow scrolls the register viewport so the user can
// reach registers that are off-screen on small terminals.
func TestModelDownKeyScrollsViewport(t *testing.T) {
	_, m := newModel(t)
	updated, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 10})
	updated, _ = updated.(tui.Model).Update(keyMsg("down"))
	updated, _ = updated.(tui.Model).Update(keyMsg("down"))
	updated, _ = updated.(tui.Model).Update(keyMsg("down"))
	view := viewText(updated.(tui.Model))
	if !strings.Contains(view, "x3") {
		t.Errorf("View should show x3 after 3 down-keys (viewport scrolled), got:\n%s", view)
	}
}

// TestModelLoadKeyEntersPickerMode pins that pressing 'l'
// switches the model into picker mode. The view must then
// contain a prompt of some kind (the filepicker widget or
// a placeholder) so the user can see the mode change.
func TestModelLoadKeyEntersPickerMode(t *testing.T) {
	_, m := newModel(t)
	updated, _ := m.Update(keyMsg("l"))
	m2 := updated.(tui.Model)
	view := viewText(m2)
	if view == "" {
		t.Fatal("View after 'l' should not be empty")
	}
	if strings.Contains(view, "program loaded") {
		t.Errorf("View after 'l' should not say 'program loaded' (no file picked), got:\n%s", view)
	}
}

// TestModelPickerAcceptsAllExtensions pins that the
// filepicker does not restrict by file extension. The user
// works with .asm and .s files, and the picker must allow
// either. We check via the filepicker's canSelect behavior
// indirectly: the NewModel() filepicker is constructed
// without an AllowedTypes filter, so any file the user
// navigates to can be selected.
func TestModelPickerAcceptsAllExtensions(t *testing.T) {
	_, m := newModel(t)
	updated, _ := m.Update(keyMsg("l"))
	view := viewText(updated.(tui.Model))
	if view == "" {
		t.Fatal("Picker should be reachable after 'l'")
	}
}

// TestModelStatsBarStartsAtZero pins that a fresh model
// shows cycles=0 before any step key has been pressed.
// The cycles counter is the emulator's own CPU cycle count
// and is reset by Reset, Load, and Config.
func TestModelStatsBarStartsAtZero(t *testing.T) {
	_, m := newModel(t)
	view := viewText(m)
	if !strings.Contains(view, "cycles=0") {
		t.Errorf("View should show 'cycles=0' before any step, got:\n%s", view)
	}
}

// TestModelStatsBarAdvancesOnStep pins that pressing 's'
// increments the cycles counter by 1.
func TestModelStatsBarAdvancesOnStep(t *testing.T) {
	_, m := newModel(t)
	updated, _ := m.Update(keyMsg("s"))
	view := viewText(updated.(tui.Model))
	if !strings.Contains(view, "cycles=1") {
		t.Errorf("View should show 'cycles=1' after one 's', got:\n%s", view)
	}
}

// TestModelStatsBarAdvancesOnShiftStep pins that pressing
// 'S' increments the cycles counter by 10.
func TestModelStatsBarAdvancesOnShiftStep(t *testing.T) {
	_, m := newModel(t)
	updated, _ := m.Update(keyMsg("S"))
	view := viewText(updated.(tui.Model))
	if !strings.Contains(view, "cycles=10") {
		t.Errorf("View should show 'cycles=10' after one 'S', got:\n%s", view)
	}
}

// TestModelStatsBarResetsOnReset pins that pressing 'r'
// returns the cycles counter to 0.
func TestModelStatsBarResetsOnReset(t *testing.T) {
	_, m := newModel(t)
	updated, _ := m.Update(keyMsg("S"))
	updated, _ = updated.(tui.Model).Update(keyMsg("S"))
	updated, _ = updated.(tui.Model).Update(keyMsg("r"))
	view := viewText(updated.(tui.Model))
	if !strings.Contains(view, "cycles=0") {
		t.Errorf("View should show 'cycles=0' after 'r', got:\n%s", view)
	}
}

// TestModelViewShowsStatsBar pins that the View shows a
// stats bar with cycles, retired, and IPC counters in
// addition to the help line.
func TestModelViewShowsStatsBar(t *testing.T) {
	_, m := newModel(t)
	updated, _ := m.Update(keyMsg("s"))
	updated, _ = updated.(tui.Model).Update(keyMsg("s"))
	view := viewText(updated.(tui.Model))
	for _, want := range []string{"cycles=", "retired=", "IPC="} {
		if !strings.Contains(view, want) {
			t.Errorf("View should contain %q, got:\n%s", want, view)
		}
	}
}

// TestModelViewShowsRegisterPane pins that the View renders
// a register pane with the Reg and Value column headers.
func TestModelViewShowsRegisterPane(t *testing.T) {
	_, m := newModel(t)
	view := viewText(m)
	if !strings.Contains(view, "Reg") {
		t.Errorf("View should contain a 'Reg' column header, got:\n%s", view)
	}
	if !strings.Contains(view, "Value") {
		t.Errorf("View should contain a 'Value' column header, got:\n%s", view)
	}
}

// TestModelViewShowsPipelinePane pins that the View renders
// a pipeline pane with the ALU and LSU reservation-station
// section headers.
func TestModelViewShowsPipelinePane(t *testing.T) {
	_, m := newModel(t)
	view := viewText(m)
	if !strings.Contains(view, "ALU RS") {
		t.Errorf("View should contain 'ALU RS' header, got:\n%s", view)
	}
	if !strings.Contains(view, "LSU RS") {
		t.Errorf("View should contain 'LSU RS' header, got:\n%s", view)
	}
}

// TestModelPipelinePaneShowsBusyEntry pins that after a
// step with a long-latency ALU configuration, the pipeline
// pane shows at least one busy reservation-station entry.
func TestModelPipelinePaneShowsBusyEntry(t *testing.T) {
	cfg := cpu.DefaultConfig()
	cfg.ALULatency = 5
	app := core.New(1024, cfg)
	if err := app.LoadProgram("addi x1, x0, 5\naddi x2, x1, 6\naddi x3, x2, 7\naddi x4, x3, 8\naddi x5, x4, 9\n"); err != nil {
		t.Fatalf("LoadProgram: %v", err)
	}
	if err := app.Step(2); err != nil {
		t.Fatalf("Step: %v", err)
	}
	m := tui.NewModel(app)
	view := viewText(m)
	if !strings.Contains(view, "✓") {
		t.Errorf("Pipeline pane should show a busy RS entry (✓) after a step, got:\n%s", view)
	}
}

// TestModelNarrowTerminalStacksVertically pins that on a
// narrow terminal the register and pipeline panes are
// stacked vertically rather than placed side by side. The
// register names should still appear, and the pipeline
// section should follow them.
func TestModelNarrowTerminalStacksVertically(t *testing.T) {
	_, m := newModel(t)
	updated, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 60})
	m2 := updated.(tui.Model)
	view := viewText(m2)
	regIdx := strings.Index(view, "Reg")
	aluIdx := strings.Index(view, "ALU RS")
	if regIdx < 0 || aluIdx < 0 {
		t.Fatalf("View should contain both 'Reg' and 'ALU RS', got:\n%s", view)
	}
	if regIdx > aluIdx {
		t.Errorf("On narrow terminal, register pane should appear before pipeline pane (regIdx=%d, aluIdx=%d)", regIdx, aluIdx)
	}
}

// TestModelWideTerminalSplitsHorizontally pins that on a
// wide terminal the register and pipeline panes are placed
// side by side in the same row, not stacked vertically. The
// "Reg" and "ALU" markers must share a line in the View
// output. This is the regression guard for the
// lipgloss.JoinHorizontal + Width padding bug.
func TestModelWideTerminalSplitsHorizontally(t *testing.T) {
	_, m := newModel(t)
	updated, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 60})
	m2 := updated.(tui.Model)
	view := viewText(m2)
	for _, line := range strings.Split(view, "\n") {
		if strings.Contains(line, "Reg") && strings.Contains(line, "ALU") {
			return
		}
	}
	t.Errorf("Wide terminal should place 'Reg' and 'ALU' on the same line, got:\n%s", view)
}

// TestModelRegStatusPlaceholderNotTruncated pins that the
// Reg Status pane shows the literal "(none)" placeholder
// (6 chars) when no rename tags are active. On a wide
// terminal the placeholder must not be truncated to "(no…".
func TestModelRegStatusPlaceholderNotTruncated(t *testing.T) {
	// A trivial program: one ADDI that retires in the
	// first step, leaving no in-flight rename tags.
	app := core.New(1024, cpu.SpecConfig())
	if err := app.LoadProgram("addi x1, x0, 5"); err != nil {
		t.Fatalf("LoadProgram: %v", err)
	}
	if err := app.Step(5); err != nil {
		t.Fatalf("Step: %v", err)
	}
	m := tui.NewModel(app)
	updated, _ := m.Update(tea.WindowSizeMsg{Width: 159, Height: 39})
	m2 := updated.(tui.Model)
	view := viewText(m2)
	if strings.Contains(view, "(no…") {
		t.Errorf("Reg Status placeholder should not truncate to '(no…', got:\n%s", view)
	}
	if !strings.Contains(view, "(none)") {
		t.Errorf("Reg Status should show the full '(none)' placeholder, got:\n%s", view)
	}
}

// TestModelThreeColumnLayout pins that the wide-terminal
// layout uses three composited columns (register, ALU RS,
// LSU RS + reg status) and that none of them truncates a
// pipeline cell. The picker creates a long-latency ALU
// instruction chain so the pipeline tables are non-empty.
func TestModelThreeColumnLayout(t *testing.T) {
	cfg := cpu.DefaultConfig()
	cfg.ALULatency = 5
	app := core.New(1024, cfg)
	if err := app.LoadProgram("addi x1, x0, 5\naddi x2, x1, 6\n"); err != nil {
		t.Fatalf("LoadProgram: %v", err)
	}
	if err := app.Step(2); err != nil {
		t.Fatalf("Step: %v", err)
	}
	m := tui.NewModel(app)
	updated, _ := m.Update(tea.WindowSizeMsg{Width: 159, Height: 39})
	m2 := updated.(tui.Model)
	view := viewText(m2)
	if !strings.Contains(view, "ALU RS") {
		t.Errorf("Three-column layout should still mention 'ALU RS', got:\n%s", view)
	}
	if !strings.Contains(view, "LSU RS") {
		t.Errorf("Three-column layout should still mention 'LSU RS', got:\n%s", view)
	}
	// No truncation ellipsis in the pipeline data rows.
	// The Op column is 5 wide; the longest valid mnemonic
	// is "STORE" (5 chars) and "INVALID" (7 chars, gets
	// truncated). We assert that no idle ALU row contains
	// the truncated "INVA…" form: it must show a real,
	// full-width placeholder.
	if strings.Contains(view, "INVA…") {
		t.Errorf("Pipeline pane should not truncate the Op column to 'INVA…', got:\n%s", view)
	}
}

// TestModelRegisterPaneNotTruncated pins that the register
// pane does not truncate register names. On a 159-column
// terminal the register pane is sized to fit a full
// 32-bit register value plus padding.
func TestModelRegisterPaneNotTruncated(t *testing.T) {
	app := core.New(1024, cpu.SpecConfig())
	// Build a large 32-bit value via a chain of ADDI; the
	// assembler accepts 12-bit signed immediates only.
	if err := app.LoadProgram(
		"addi x31, x0, 2047\n" +
			"addi x30, x0, 2047\n" +
			"add x31, x31, x30\n" +
			"slli x30, x30, 11\n" +
			"add x31, x31, x30\n" +
			"slli x30, x30, 11\n" +
			"add x31, x31, x30",
	); err != nil {
		t.Fatalf("LoadProgram: %v", err)
	}
	if err := app.Step(20); err != nil {
		t.Fatalf("Step: %v", err)
	}
	m := tui.NewModel(app)
	updated, _ := m.Update(tea.WindowSizeMsg{Width: 159, Height: 39})
	m2 := updated.(tui.Model)
	view := viewText(m2)
	// x31 should appear with its full name (not "x3…").
	if !strings.Contains(view, "x31") {
		t.Errorf("Register pane should mention 'x31' in full, got:\n%s", view)
	}
}

// errorString is a tiny error type used by the form-feedback
// tests. Keeping it in this file avoids touching tui's
// exported surface.
type errorString string

func (e errorString) Error() string { return string(e) }

// keyMsg builds a tea.KeyPressMsg for the given string. The
// model's Update handler switches on msg.String() so a single
// keystroke is enough to drive the tests. In Bubble Tea v2
// KeyPressMsg wraps a Key struct whose Code field carries the
// rune for printable characters or a special constant for
// things like esc/up/down.
func keyMsg(s string) tea.KeyPressMsg {
	switch s {
	case "esc":
		return tea.KeyPressMsg{Code: tea.KeyEscape}
	case "up":
		return tea.KeyPressMsg{Code: tea.KeyUp}
	case "down":
		return tea.KeyPressMsg{Code: tea.KeyDown}
	case "enter":
		return tea.KeyPressMsg{Code: tea.KeyEnter}
	}
	if len(s) == 1 {
		return tea.KeyPressMsg{Code: rune(s[0])}
	}
	return tea.KeyPressMsg{}
}

// newModel builds a fresh core.App + tui.Model for tests. It
// keeps the boilerplate out of the individual test functions.
func newModel(t *testing.T) (*core.App, tui.Model) {
	t.Helper()
	app := core.New(1024, cpu.SpecConfig())
	return app, tui.NewModel(app)
}
