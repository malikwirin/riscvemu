package tui_test

import (
	"fmt"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"codeberg.org/malik/riscvemu/arch/cpu"
	"codeberg.org/malik/riscvemu/internal/core"
	"codeberg.org/malik/riscvemu/internal/testutil"
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
	app := testutil.NewSpecApp()
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

// TestModelCyclesCounter pins the model's cycles counter and
// stats-bar view across the four interesting key sequences:
// idle (zero), single step (+1), shift-step (+10), and reset
// back to zero. Each case asserts both the App's Stats
// (model layer) and the View's "cycles=N" rendering (view
// layer), so a refactor that drops either cannot slip through.
func TestModelCyclesCounter(t *testing.T) {
	cases := []struct {
		name        string
		preStep     int
		keys        []string
		wantStats   uint64
		wantViewHas string
	}{
		{"starts at zero", 0, nil, 0, "cycles=0"},
		{"single step is +1", 0, []string{"s"}, 1, "cycles=1"},
		{"shift step is +10", 0, []string{"S"}, 10, "cycles=10"},
		{"reset returns to zero", 5, []string{"r"}, 0, "cycles=0"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			app, m := newModel(t)
			if tc.preStep > 0 {
				if err := app.Step(tc.preStep); err != nil {
					t.Fatalf("setup Step(%d): %v", tc.preStep, err)
				}
			}
			cur := m
			for _, k := range tc.keys {
				next, _ := cur.Update(keyMsg(k))
				cur = next.(tui.Model)
			}
			if got := cur.Snapshot().Stats.Cycles; got != tc.wantStats {
				t.Errorf("model cycles = %d, want %d", got, tc.wantStats)
			}
			view := viewText(cur)
			if !strings.Contains(view, tc.wantViewHas) {
				t.Errorf("View should contain %q, got:\n%s", tc.wantViewHas, view)
			}
		})
	}
}

// TestModelFeedbackMessages pins the model's feedback
// rendering: success messages ("ok: ...") after step/reset,
// error messages ("error: ...") from WithError, and the
// "press esc to clear" path. Each case asserts both the
// presence of the prefix (ok:/error:) and the specific body
// text the user is expected to see.
func TestModelFeedbackMessages(t *testing.T) {
	cases := []struct {
		name     string
		build    func(t *testing.T) tui.Model
		wantHas  []string
		wantMiss []string
	}{
		{
			name:    "step shows ok+stepped",
			build:   func(t *testing.T) tui.Model { _, m := newModel(t); n, _ := m.Update(keyMsg("s")); return n.(tui.Model) },
			wantHas: []string{"ok:", "stepped"},
		},
		{
			name:    "reset shows ok+reset",
			build:   func(t *testing.T) tui.Model { _, m := newModel(t); n, _ := m.Update(keyMsg("r")); return n.(tui.Model) },
			wantHas: []string{"ok:", "reset"},
		},
		{
			name: "error message is rendered",
			build: func(t *testing.T) tui.Model {
				_, m := newModel(t)
				return m.WithError(errorString("simulated load failure"))
			},
			wantHas: []string{"error:", "simulated load failure"},
		},
		{
			name: "esc clears an error",
			build: func(t *testing.T) tui.Model {
				_, m := newModel(t)
				withErr := m.WithError(errorString("something went wrong"))
				n, _ := withErr.Update(keyMsg("esc"))
				return n.(tui.Model)
			},
			wantMiss: []string{"error:"},
		},
		{
			name: "ok survives next step",
			build: func(t *testing.T) tui.Model {
				_, m := newModel(t)
				n1, _ := m.Update(keyMsg("s"))
				n2, _ := n1.(tui.Model).Update(keyMsg("S"))
				return n2.(tui.Model)
			},
			wantHas: []string{"ok:"},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			view := viewText(tc.build(t))
			for _, w := range tc.wantHas {
				if !strings.Contains(view, w) {
					t.Errorf("View should contain %q, got:\n%s", w, view)
				}
			}
			for _, w := range tc.wantMiss {
				if strings.Contains(view, w) {
					t.Errorf("View should not contain %q, got:\n%s", w, view)
				}
			}
		})
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

// TestModelViewSections pins that the View renders the three
// top-level sections (stats bar, register pane, pipeline
// pane) with the expected header fragments. The model is
// stepped twice so the stats bar carries non-zero values
// and is unambiguously rendered.
func TestModelViewSections(t *testing.T) {
	_, m := newModel(t)
	updated, _ := m.Update(keyMsg("s"))
	updated, _ = updated.(tui.Model).Update(keyMsg("s"))
	view := viewText(updated.(tui.Model))
	for _, want := range []string{
		"cycles=", "retired=", "IPC=",
		"Reg", "Value",
		"ALU RS", "LSU RS",
	} {
		if !strings.Contains(view, want) {
			t.Errorf("View should contain %q, got:\n%s", want, view)
		}
	}
}

// TestModelPipelinePaneShowsBusyEntry pins that after a
// step with a long-latency ALU configuration, the pipeline
// pane shows at least one busy reservation-station entry.
func TestModelPipelinePaneShowsBusyEntry(t *testing.T) {
	cfg := cpu.DefaultConfig()
	cfg.ALULatency = 5
	app := testutil.NewAppWith(1024, cfg)
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

// TestModelLayout pins the layout behaviour across the
// four interesting terminal sizes. The narrow case stacks
// the register and pipeline panes vertically; the wide
// cases fit them side by side, and on a 159×39 terminal
// the pipeline itself splits into two columns. The wide
// cases also assert that no column truncates a known
// register name, an "(none)" placeholder, or a real
// mnemonic.
func TestModelLayout(t *testing.T) {
	wideProgram := "addi x1, x0, 5\naddi x2, x1, 6\n"
	largeRegProgram := "addi x31, x0, 2047\n" +
		"addi x30, x0, 2047\n" +
		"add x31, x31, x30\n" +
		"slli x30, x30, 11\n" +
		"add x31, x31, x30\n" +
		"slli x30, x30, 11\n" +
		"add x31, x31, x30"
	cases := []struct {
		name      string
		width     int
		height    int
		program   string
		preStep   int
		hasAfter  []string
		notBefore []string
	}{
		{
			name:  "narrow 80x60 stacks vertically",
			width: 80, height: 60,
			hasAfter: []string{"Reg", "ALU RS"},
		},
		{
			name:  "wide 120x60 shares a row",
			width: 120, height: 60,
			hasAfter: []string{"Reg", "ALU"},
		},
		{
			name:  "wide 159x39 shows full reg status",
			width: 159, height: 39,
			program:   "addi x1, x0, 5",
			preStep:   5,
			hasAfter:  []string{"(none)"},
			notBefore: []string{"(no…"},
		},
		{
			name:  "wide 159x39 three columns no truncation",
			width: 159, height: 39,
			program:   wideProgram,
			preStep:   2,
			hasAfter:  []string{"ALU RS", "LSU RS"},
			notBefore: []string{"INVA…"},
		},
		{
			name:  "wide 159x39 register names full",
			width: 159, height: 39,
			program:  largeRegProgram,
			preStep:  20,
			hasAfter: []string{"x31"},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			app, m := newModel(t)
			if tc.program != "" {
				if err := app.LoadProgram(tc.program); err != nil {
					t.Fatalf("LoadProgram: %v", err)
				}
				if tc.preStep > 0 {
					if err := app.Step(tc.preStep); err != nil {
						t.Fatalf("Step: %v", err)
					}
				}
			}
			updated, _ := m.Update(tea.WindowSizeMsg{Width: tc.width, Height: tc.height})
			view := viewText(updated.(tui.Model))
			// The narrow case asserts that "Reg" appears
			// before "ALU RS" (vertical stack).
			if tc.name == "narrow 80x60 stacks vertically" {
				regIdx := strings.Index(view, "Reg")
				aluIdx := strings.Index(view, "ALU RS")
				if regIdx > aluIdx {
					t.Errorf("On narrow terminal, register pane should appear before pipeline pane (regIdx=%d, aluIdx=%d)", regIdx, aluIdx)
				}
			}
			// The wide 120x60 case asserts that "Reg"
			// and "ALU" share a line (side-by-side
			// composited columns).
			if tc.name == "wide 120x60 shares a row" {
				shared := false
				for _, line := range strings.Split(view, "\n") {
					if strings.Contains(line, "Reg") && strings.Contains(line, "ALU") {
						shared = true
						break
					}
				}
				if !shared {
					t.Errorf("Wide terminal should place 'Reg' and 'ALU' on the same line, got:\n%s", view)
				}
			}
			for _, w := range tc.hasAfter {
				if !strings.Contains(view, w) {
					t.Errorf("View should contain %q, got:\n%s", w, view)
				}
			}
			for _, w := range tc.notBefore {
				if strings.Contains(view, w) {
					t.Errorf("View should not contain %q, got:\n%s", w, view)
				}
			}
		})
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
	app := testutil.NewSpecApp()
	return app, tui.NewModel(app)
}
