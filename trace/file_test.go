package trace

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestParseTraceFileHappyPath writes a small trace to a tempfile
// and reads it back through ParseTraceFile. The function exercises
// both the file I/O path and the line-by-line parser.
func TestParseTraceFileHappyPath(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "spec-example.trace")
	input := `# Spec example, abridged to the supported subset
ADD R1, R2, R3
MUL R4, R1, R5
LOAD R6, 4(R2)
SUB R7, R6, R4
DIV R0, R7, R1
s
h
i
`
	if err := os.WriteFile(path, []byte(input), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
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
}

// TestParseTraceFileMissing wraps an *os.PathError-style failure
// in a *ParseError with LineNum 0. Callers can use the line number
// to distinguish "the file was unreadable" from "the file is
// syntactically broken".
func TestParseTraceFileMissing(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "does-not-exist.trace")
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
}

// TestParseTraceFileEmpty verifies that an empty file produces an
// empty line stream, not an error. The Spec does not forbid empty
// traces; the driver (later change) will report nothing happened.
func TestParseTraceFileEmpty(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "empty.trace")
	if err := os.WriteFile(path, nil, 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	lines, err := ParseTraceFile(path)
	if err != nil {
		t.Fatalf("ParseTraceFile: %v", err)
	}
	if len(lines) != 0 {
		t.Errorf("got %d lines, want 0", len(lines))
	}
}

// TestParseTraceFileSyntaxError verifies that a ParseError
// surfaces the line number of the offending entry. The error
// originates from ParseTrace but is wrapped through ParseTraceFile
// unchanged.
func TestParseTraceFileSyntaxError(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.trace")
	input := "ADD R1, R2, R3\nBEQ R1, R2, 4\n"
	if err := os.WriteFile(path, []byte(input), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
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
}
