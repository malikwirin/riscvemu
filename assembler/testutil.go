package assembler

import (
	"os"
	"testing"
)

// checkInstructions compares a slice of Instructions with expected instruction strings.
// It fails the test if the length or any instruction does not match.
func checkInstructions(t *testing.T, got []Instruction, want []string) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("Expected %d instructions, got %d", len(want), len(got))
	}
	for i, line := range want {
		expected, err := ParseInstruction(line)
		if err != nil {
			t.Fatalf("ParseInstruction failed for %q: %v", line, err)
		}
		if got[i] != expected {
			t.Errorf("Instruction %d mismatch. Got %08x, want %08x", i, got[i], expected)
		}
	}
}

// mustParse parses a single assembler instruction and panics on error.
// Use only in tests where failure is not expected.
func mustParse(line string) Instruction {
	instr, err := ParseInstruction(line)
	if err != nil {
		panic(err)
	}
	return instr
}

// writeTempASM writes the provided assembly code to a temporary file and returns its name.
// The file is scheduled for cleanup after the test.
func writeTempASM(t *testing.T, code string) string {
	t.Helper()
	tmpfile, err := os.CreateTemp("", "test-*.asm")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	t.Cleanup(func() { os.Remove(tmpfile.Name()) })
	if _, err := tmpfile.WriteString(code); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	tmpfile.Close()
	return tmpfile.Name()
}
