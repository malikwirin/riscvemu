package tui_test

import (
	"testing"

	"codeberg.org/malik/riscvemu/tui"
)

// TestValidateLoadProgramPathRejectsEmpty pins that an empty
// path string fails validation. The form cannot proceed without
// a file to read.
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
// string fails validation. The Config form requires integers.
func TestValidatePositiveIntRejectsNonInt(t *testing.T) {
	if err := tui.ValidatePositiveInt("abc"); err == nil {
		t.Fatal("ValidatePositiveInt(\"abc\"): expected error, got nil")
	}
}

// TestValidatePositiveIntRejectsZero pins that zero fails
// validation. A reservation station with zero slots makes no
// sense, so the form rejects it.
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

// TestValidatePositiveIntAcceptsPositive pins that any
// positive integer passes validation.
func TestValidatePositiveIntAcceptsPositive(t *testing.T) {
	if err := tui.ValidatePositiveInt("4"); err != nil {
		t.Errorf("ValidatePositiveInt(\"4\"): %v", err)
	}
}

// TestLoadProgramFormRejectsNilApp pins the precondition that
// LoadProgramForm refuses a nil app. Without this check the
// Huh form would nil-deref on submission.
func TestLoadProgramFormRejectsNilApp(t *testing.T) {
	if err := tui.LoadProgramForm(nil); err == nil {
		t.Fatal("LoadProgramForm(nil): expected error, got nil")
	}
}

// TestConfigFormRejectsNilApp pins the precondition that
// ConfigForm refuses a nil app.
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
	if tui.IsCancelled(errString("some other error")) {
		t.Error("IsCancelled on a non-cancellation error = true, want false")
	}
}

type errString string

func (e errString) Error() string { return string(e) }
