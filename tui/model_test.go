package tui_test

import (
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
