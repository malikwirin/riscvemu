package tui_test

import (
	"strings"
	"testing"

	"codeberg.org/malik/riscvemu/arch/cpu"
	"codeberg.org/malik/riscvemu/internal/core"
	"codeberg.org/malik/riscvemu/tui"
	tea "github.com/charmbracelet/bubbletea"
)

// TestRunRejectsNilApp pins the precondition that Run refuses
// a nil app. Without this check the Bubble-Tea program would
// nil-deref on the first Update.
func TestRunRejectsNilApp(t *testing.T) {
	err := tui.Run(nil)
	if err == nil {
		t.Fatal("Run(nil): expected error, got nil")
	}
	if !strings.Contains(err.Error(), "nil") {
		t.Errorf("Run(nil) error %q should mention nil", err)
	}
}

// TestModelIsBubbleTeaModel is a compile-time guarantee that
// Model satisfies tea.Model. If Init/Update/View are ever
// removed the test file fails to compile.
func TestModelIsBubbleTeaModel(t *testing.T) {
	var _ tea.Model = tui.NewModel(core.New(1024, cpu.SpecConfig()))
}

// TestModelViewContainsProjectName pins that the View output
// contains a header line.
func TestModelViewContainsProjectName(t *testing.T) {
	app := core.New(1024, cpu.SpecConfig())
	m := tui.NewModel(app)
	view := m.View()
	if !strings.Contains(view, "riscvemu") {
		t.Errorf("View output should mention the project name, got: %q", view)
	}
}
