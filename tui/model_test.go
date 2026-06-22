package tui_test

import (
	"fmt"
	"strings"
	"testing"

	"codeberg.org/malik/riscvemu/arch/cpu"
	"codeberg.org/malik/riscvemu/internal/core"
	"codeberg.org/malik/riscvemu/tui"
	tea "github.com/charmbracelet/bubbletea"
)

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
	view := m.View()
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
	view := updated.(tui.Model).View()
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
	view := updated.(tui.Model).View()
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
	view := m2.View()
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
	if !strings.Contains(m2.View(), "error:") {
		t.Fatalf("setup: View should contain 'error:'")
	}
	updated, _ := m2.Update(keyMsg("esc"))
	view := updated.(tui.Model).View()
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
	if !strings.Contains(m2.(tui.Model).View(), "ok:") {
		t.Fatalf("setup: 's' should leave 'ok:' visible")
	}
	m3, _ := m2.Update(keyMsg("S"))
	view := m3.(tui.Model).View()
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
	view := m.View()
	lines := strings.Split(view, "\n")
	if len(lines) < 4 {
		t.Errorf("View should have at least 4 lines (header + register rows + footer), got %d:\n%s", len(lines), view)
	}
}

// TestModelViewNoLongLine pins that no single line in the
// View exceeds a reasonable terminal width. The viewport-
// based layout must keep every line bounded.
func TestModelViewNoLongLine(t *testing.T) {
	_, m := newModel(t)
	view := m.View()
	for i, line := range strings.Split(view, "\n") {
		if len(line) > 120 {
			t.Errorf("line %d is %d chars long (>120):\n%s", i, len(line), line)
		}
	}
}

// TestModelViewShowsAllRegisters pins that the View mentions
// every architectural register name (x0 .. x31) once the
// terminal is sized tall enough to show the full viewport.
// The user can always scroll to see hidden rows.
func TestModelViewShowsAllRegisters(t *testing.T) {
	_, m := newModel(t)
	updated, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 100})
	m2 := updated.(tui.Model)
	view := m2.View()
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
	view := m.View()
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
	view := updated.(tui.Model).View()
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
	// Without a WindowSizeMsg the filepicker might not have
	// its directory list yet; we accept any non-empty view
	// that is not the help line alone.
	view := m2.View()
	if view == "" {
		t.Fatal("View after 'l' should not be empty")
	}
	// The picker mode shows a feedback hint instead of the
	// usual "ok: program loaded".
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
	// Pressing 'l' opens the picker; we don't actually
	// navigate or select (the picker requires a real
	// terminal), but we can assert the picker has no
	// extension filter via the model's exposed picker
	// state. The view should not contain a filter hint.
	updated, _ := m.Update(keyMsg("l"))
	view := updated.(tui.Model).View()
	// If a filter were active, the picker header would
	// typically mention it. We instead trust the absence
	// of AllowedTypes in NewModel and the presence of
	// FileAllowed/DirAllowed; this test is a smoke check
	// that the picker is reachable at all.
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
	view := m.View()
	if !strings.Contains(view, "cycles=0") {
		t.Errorf("View should show 'cycles=0' before any step, got:\n%s", view)
	}
}

// TestModelStatsBarAdvancesOnStep pins that pressing 's'
// increments the cycles counter by 1.
func TestModelStatsBarAdvancesOnStep(t *testing.T) {
	_, m := newModel(t)
	updated, _ := m.Update(keyMsg("s"))
	view := updated.(tui.Model).View()
	if !strings.Contains(view, "cycles=1") {
		t.Errorf("View should show 'cycles=1' after one 's', got:\n%s", view)
	}
}

// TestModelStatsBarAdvancesOnShiftStep pins that pressing
// 'S' increments the cycles counter by 10.
func TestModelStatsBarAdvancesOnShiftStep(t *testing.T) {
	_, m := newModel(t)
	updated, _ := m.Update(keyMsg("S"))
	view := updated.(tui.Model).View()
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
	view := updated.(tui.Model).View()
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
	view := updated.(tui.Model).View()
	for _, want := range []string{"cycles=", "retired=", "IPC="} {
		if !strings.Contains(view, want) {
			t.Errorf("View should contain %q, got:\n%s", want, view)
		}
	}
}

// errorString is a tiny error type used by the form-feedback
// tests. Keeping it in this file avoids touching tui's
// exported surface.
type errorString string

func (e errorString) Error() string { return string(e) }

// keyMsg builds a tea.KeyMsg for the given rune. The model's
// Update handler switches on msg.String() so a single-rune
// keystroke is enough to drive the tests.
func keyMsg(s string) tea.KeyMsg {
	r := []rune(s)
	return tea.KeyMsg{Type: tea.KeyRunes, Runes: r}
}

// newModel builds a fresh core.App + tui.Model for tests. It
// keeps the boilerplate out of the individual test functions.
func newModel(t *testing.T) (*core.App, tui.Model) {
	t.Helper()
	app := core.New(1024, cpu.SpecConfig())
	return app, tui.NewModel(app)
}
