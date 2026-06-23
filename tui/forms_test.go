package tui_test

import (
	"testing"

	"charm.land/huh/v2"
	"codeberg.org/malik/riscvemu/tui"
)

// TestValidateLoadProgramPathRejectsEmpty pins that an empty
// path string fails validation. The form cannot proceed
// without a file to read.
func TestValidateLoadProgramPathRejectsEmpty(t *testing.T) {
	if err := tui.ValidateLoadProgramPath(""); err == nil {
		t.Fatal("ValidateLoadProgramPath(\"\"): expected error, got nil")
	}
}

// TestValidateLoadProgramPathAcceptsNonEmpty pins that any
// non-empty string passes validation. The actual file
// existence is checked later by LoadProgramFromFile, not by
// the form.
func TestValidateLoadProgramPathAcceptsNonEmpty(t *testing.T) {
	if err := tui.ValidateLoadProgramPath("program.s"); err != nil {
		t.Errorf("ValidateLoadProgramPath(\"program.s\"): %v", err)
	}
}

// TestValidatePositiveIntRejectsNonInt pins that a non-numeric
// string fails validation.
func TestValidatePositiveIntRejectsNonInt(t *testing.T) {
	if err := tui.ValidatePositiveInt("abc"); err == nil {
		t.Fatal("ValidatePositiveInt(\"abc\"): expected error, got nil")
	}
}

// TestValidatePositiveIntRejectsZero pins that zero fails
// validation.
func TestValidatePositiveIntRejectsZero(t *testing.T) {
	if err := tui.ValidatePositiveInt("0"); err == nil {
		t.Fatal("ValidatePositiveInt(\"0\"): expected error, got nil")
	}
}

// TestValidatePositiveIntRejectsNegative pins that negative
// values fail validation.
func TestValidatePositiveIntRejectsNegative(t *testing.T) {
	if err := tui.ValidatePositiveInt("-3"); err == nil {
		t.Fatal("ValidatePositiveInt(\"-3\"): expected error, got nil")
	}
}

// TestConfigFormRejectsNilApp pins the precondition. (Note
// the name is misleading because the value is captured at
// call time before the nil dereference; the guard simply
// gives a clearer error message than a panic.)
func TestConfigFormGuardsAgainstNilApp(t *testing.T) {
	if err := tui.ConfigForm(nil); err == nil {
		t.Fatal("ConfigForm(nil): expected error, got nil")
	}
}

// TestConfigFormLatenciesExposed pins that the ConfigForm
// surface accepts latency fields, not just RS counts. We
// don't run the form headlessly; we just make sure the
// surface advertises all nine configuration knobs. A future
// change to a separate LatencyForm would break this pin.
func TestConfigFormLatenciesExposed(t *testing.T) {
	names := tui.ConfigFormFieldNames()
	want := []string{
		"ALU RS count", "LSU RS count",
		"MUL RS count", "DIV RS count",
		"ALU latency", "MUL latency",
		"DIV latency", "Load latency", "Store latency",
	}
	if len(names) != len(want) {
		t.Errorf("ConfigForm exposes %d fields, want %d: %v", len(names), len(want), names)
	}
	for i, w := range want {
		if i < len(names) && names[i] != w {
			t.Errorf("ConfigForm field %d = %q, want %q", i, names[i], w)
		}
	}
}

// TestValidatePositiveIntAcceptsPositive pins that any
// positive integer passes validation.
func TestValidatePositiveIntAcceptsPositive(t *testing.T) {
	if err := tui.ValidatePositiveInt("4"); err != nil {
		t.Errorf("ValidatePositiveInt(\"4\"): %v", err)
	}
}

// TestLoadProgramFormRejectsNilApp pins the precondition.
func TestLoadProgramFormRejectsNilApp(t *testing.T) {
	if err := tui.LoadProgramForm(nil); err == nil {
		t.Fatal("LoadProgramForm(nil): expected error, got nil")
	}
}

// TestConfigFormRejectsNilApp pins the precondition.
func TestConfigFormRejectsNilApp(t *testing.T) {
	if err := tui.ConfigForm(nil); err == nil {
		t.Fatal("ConfigForm(nil): expected error, got nil")
	}
}

// TestIsCancelledDistinguishesCancellation pins that
// IsCancelled reports true only for the sentinel error
// returned by user-cancelled forms, not for other errors.
func TestIsCancelledDistinguishesCancellation(t *testing.T) {
	if !tui.IsCancelled(tui.ErrCancelled) {
		t.Error("IsCancelled(ErrCancelled) = false, want true")
	}
	if tui.IsCancelled(nil) {
		t.Error("IsCancelled(nil) = true, want false")
	}
	if tui.IsCancelled(huh.ErrUserAborted) {
		t.Error("IsCancelled on huh.ErrUserAborted = true, want false (huh v2 has its own sentinel)")
	}
}

type errString string

func (e errString) Error() string { return string(e) }
