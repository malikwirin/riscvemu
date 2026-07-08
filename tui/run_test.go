package tui_test

import (
	"strings"
	"testing"

	"codeberg.org/malik/riscvemu/tui"
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
