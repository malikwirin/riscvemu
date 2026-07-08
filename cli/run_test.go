package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"codeberg.org/malik/riscvemu/internal/testutil"
)

// TestRunScriptLoadsAsm pins the assembly path: a one-line
// asm file (addi x1, x0, 5) is loaded, stepped 3 cycles,
// and the one-line stats summary is printed. The test
// captures stdout and asserts the summary line is
// present.
func TestRunScriptLoadsAsm(t *testing.T) {
	app := testutil.NewSpecApp()
	path := writeTemp(t, "add.s", "addi x1, x0, 5\n")
	out := captureOutput(func() {
		if err := RunScript(app, path, 3); err != nil {
			t.Errorf("RunScript: %v", err)
		}
	})
	for _, want := range []string{"cycles=3", "retired=", "IPC="} {
		if !contains(out, want) {
			t.Errorf("output should contain %q, got:\n%s", want, out)
		}
	}
	if got := app.Snapshot().Registers[1]; got != 5 {
		t.Errorf("R1 = %d, want 5", got)
	}
}

// TestRunScriptFallsBackToTrace pins the trace path: a
// trace file (Spec-format ADD) is loaded after the
// assembler rejects the same content, and the run
// succeeds.
func TestRunScriptFallsBackToTrace(t *testing.T) {
	app := testutil.NewSpecApp()
	path := writeTemp(t, "add.trace", "ADD R1, R2, R3\n")
	out := captureOutput(func() {
		if err := RunScript(app, path, 5); err != nil {
			t.Errorf("RunScript: %v", err)
		}
	})
	if !contains(out, "cycles=5") {
		t.Errorf("output should contain cycles=5, got:\n%s", out)
	}
}

// TestRunScriptRejectsEmptyPath pins that a missing
// program name produces a clear error without ever
// touching the filesystem.
func TestRunScriptRejectsEmptyPath(t *testing.T) {
	app := testutil.NewSpecApp()
	err := RunScript(app, "", 10)
	if err == nil {
		t.Fatal("RunScript(\"\"): want error, got nil")
	}
	if !contains(err.Error(), "no program") {
		t.Errorf("error = %q, want 'no program'", err)
	}
}

// TestRunScriptRejectsMissingFile pins that a path that
// cannot be read is reported with the path embedded in
// the error message.
func TestRunScriptRejectsMissingFile(t *testing.T) {
	app := testutil.NewSpecApp()
	err := RunScript(app, "/no/such/file.s", 10)
	if err == nil {
		t.Fatal("RunScript(missing): want error, got nil")
	}
	if !contains(err.Error(), "/no/such/file.s") {
		t.Errorf("error = %q, want to mention the path", err)
	}
}

// TestRunScriptRejectsNegativeCycles pins that the
// safety check on the cycle count fires before any
// step. The app is left untouched.
func TestRunScriptRejectsNegativeCycles(t *testing.T) {
	app := testutil.NewSpecApp()
	path := writeTemp(t, "add.s", "addi x1, x0, 5\n")
	if err := RunScript(app, path, -1); err == nil {
		t.Fatal("RunScript(-1): want error, got nil")
	}
	if got := app.Snapshot().Stats.Cycles; got != 0 {
		t.Errorf("cycles = %d after rejected negative count, want 0", got)
	}
}

// TestRunScriptRejectsInvalidContent pins the "not asm
// and not trace" branch: a file that is neither valid
// assembly nor a valid trace must surface a clear error
// mentioning both attempts. The exact substrings are
// "not assembly" (the wrapped LoadProgram error) and
// "trace" (the wrapped LoadTrace error, whose format
// string is "<mnemonic> at line N: ..." so the word
// "trace" is in the second wrapping).
func TestRunScriptRejectsInvalidContent(t *testing.T) {
	app := testutil.NewSpecApp()
	path := writeTemp(t, "garbage.s", "totally not a program\n")
	err := RunScript(app, path, 5)
	if err == nil {
		t.Fatal("RunScript(garbage): want error, got nil")
	}
	if !contains(err.Error(), "not assembly") {
		t.Errorf("error = %q, want 'not assembly'", err)
	}
	if !contains(err.Error(), "trace") {
		t.Errorf("error = %q, want 'trace' (from the trace-fallback error)", err)
	}
}

// writeTemp writes content to a fresh file under t.TempDir
// and returns the path. Used by every TestRunScript* case.
func writeTemp(t *testing.T, name, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	return path
}

// avoid unused-import warnings if strings is not used in
// some future edit.
var _ = strings.Contains
