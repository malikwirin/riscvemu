package trace

import (
	"path/filepath"
	"strings"
	"testing"
)

// TestParseTraceFile pins the four observable behaviours of
// ParseTraceFile: a happy path with one of each line kind,
// an empty file (returns zero lines, no error), a missing
// file (wraps a *ParseError with LineNum 0 and a path
// mention), and a syntax error (wraps a *ParseError with
// the offending line number and a mnemonic hint).
func TestParseTraceFile(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		path := WriteTempTrace(t, "spec-example.trace", `# Spec example, abridged to the supported subset
ADD R1, R2, R3
MUL R4, R1, R5
LOAD R6, 4(R2)
SUB R7, R6, R4
DIV R0, R7, R1
s
h
i
`)
		lines, err := ParseTraceFile(path)
		if err != nil {
			t.Fatalf("ParseTraceFile: %v", err)
		}
		if len(lines) != 8 {
			t.Fatalf("got %d lines, want 8", len(lines))
		}
		wantMnems := []string{"ADD", "MUL", "LOAD", "SUB", "DIV"}
		for i, want := range wantMnems {
			if lines[i].Instr.Mnemonic != want {
				t.Errorf("line %d: Mnemonic = %q, want %q", i, lines[i].Instr.Mnemonic, want)
			}
		}
		if lines[5].Kind != LineControl || lines[5].Control.Op != 's' {
			t.Errorf("line 5: kind=%d op=%c, want LineControl/'s'", lines[5].Kind, lines[5].Control.Op)
		}
		if lines[6].Control.Op != 'h' {
			t.Errorf("line 6: op = %c, want h", lines[6].Control.Op)
		}
		if lines[7].Control.Op != 'i' {
			t.Errorf("line 7: op = %c, want i", lines[7].Control.Op)
		}
	})

	t.Run("empty file", func(t *testing.T) {
		path := WriteTempTrace(t, "empty.trace", "")
		lines, err := ParseTraceFile(path)
		if err != nil {
			t.Fatalf("ParseTraceFile: %v", err)
		}
		if len(lines) != 0 {
			t.Errorf("got %d lines, want 0", len(lines))
		}
	})

	t.Run("missing file", func(t *testing.T) {
		// The file does not need to exist; t.TempDir gives
		// us a path under a real directory so the
		// missing-file branch of ParseTraceFile fires.
		path := filepath.Join(t.TempDir(), "does-not-exist.trace")
		_, err := ParseTraceFile(path)
		if err == nil {
			t.Fatal("expected error for missing file, got nil")
		}
		pe, ok := err.(*ParseError)
		if !ok {
			t.Fatalf("error type = %T, want *ParseError", err)
		}
		if pe.LineNum != 0 {
			t.Errorf("LineNum = %d, want 0 (I/O error has no line number)", pe.LineNum)
		}
		if !strings.Contains(pe.Reason, "does-not-exist.trace") {
			t.Errorf("Reason %q should mention the path", pe.Reason)
		}
	})

	t.Run("syntax error", func(t *testing.T) {
		path := WriteTempTrace(t, "bad.trace", "ADD R1, R2, R3\nBEQ R1, R2, 4\n")
		_, err := ParseTraceFile(path)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		pe, ok := err.(*ParseError)
		if !ok {
			t.Fatalf("error type = %T, want *ParseError", err)
		}
		if pe.LineNum != 2 {
			t.Errorf("LineNum = %d, want 2", pe.LineNum)
		}
		if !strings.Contains(pe.Reason, "unsupported mnemonic") {
			t.Errorf("Reason = %q, want substring %q", pe.Reason, "unsupported mnemonic")
		}
	})
}
