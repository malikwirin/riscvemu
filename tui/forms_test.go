package tui_test

import (
	"testing"

	"charm.land/huh/v2"
	"codeberg.org/malik/riscvemu/tui"
)

// TestValidateLoadProgramPath pins the path validator: any
// non-empty path is accepted (file existence is checked later
// by LoadProgramFromFile), and the empty string is rejected.
func TestValidateLoadProgramPath(t *testing.T) {
	cases := []struct {
		path    string
		wantErr bool
	}{
		{"", true},
		{"program.s", false},
	}
	for _, tc := range cases {
		err := tui.ValidateLoadProgramPath(tc.path)
		if tc.wantErr && err == nil {
			t.Errorf("ValidateLoadProgramPath(%q): want error, got nil", tc.path)
		}
		if !tc.wantErr && err != nil {
			t.Errorf("ValidateLoadProgramPath(%q): %v", tc.path, err)
		}
	}
}

// TestValidatePositiveInt pins the integer validator: only
// strictly positive integers are accepted. Zero, negatives,
// and non-numeric input are rejected.
func TestValidatePositiveInt(t *testing.T) {
	cases := []struct {
		input   string
		wantErr bool
	}{
		{"abc", true},
		{"0", true},
		{"-3", true},
		{"4", false},
	}
	for _, tc := range cases {
		err := tui.ValidatePositiveInt(tc.input)
		if tc.wantErr && err == nil {
			t.Errorf("ValidatePositiveInt(%q): want error, got nil", tc.input)
		}
		if !tc.wantErr && err != nil {
			t.Errorf("ValidatePositiveInt(%q): %v", tc.input, err)
		}
	}
}

// TestFormsRejectNilApp pins the precondition that both
// LoadProgramForm and ConfigForm refuse a nil app. The guard
// gives a clearer error message than a panic on the first
// dereference inside the form.
func TestFormsRejectNilApp(t *testing.T) {
	if err := tui.LoadProgramForm(nil); err == nil {
		t.Error("LoadProgramForm(nil): want error, got nil")
	}
	if err := tui.ConfigForm(nil); err == nil {
		t.Error("ConfigForm(nil): want error, got nil")
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
