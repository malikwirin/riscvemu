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

// TestModelQuitKeyReturnsTeaQuit pins that pressing 'q' asks
// Bubble Tea to quit the program.
func TestModelQuitKeyReturnsTeaQuit(t *testing.T) {
	app := core.New(1024, cpu.SpecConfig())
	m := tui.NewModel(app)
	_, cmd := m.Update(keyMsg("q"))
	if cmd == nil {
		t.Fatal("Update for 'q' should return a non-nil tea.Cmd")
	}
}

// keyMsg builds a tea.KeyMsg for the given rune. The model's
// Update handler switches on msg.String() so a single-rune
// keystroke is enough to drive the tests.
func keyMsg(s string) tea.KeyMsg {
	r := []rune(s)
	return tea.KeyMsg{Type: tea.KeyRunes, Runes: r}
}
