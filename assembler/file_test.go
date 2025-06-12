package assembler

import (
	"os"
	"testing"
)

func TestAssembleFile_SimpleProgram(t *testing.T) {
	// Prepare a simple assembler source code
	asm := "addi x1, x0, 42\naddi x2, x1, 1"

	// Create a temporary file with the assembler code
	tmpfile, err := os.CreateTemp("", "test-*.asm")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.WriteString(asm); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	tmpfile.Close()

	// Call AssembleFile
	prog, err := AssembleFile(tmpfile.Name())
	if err != nil {
		t.Fatalf("AssembleFile returned error: %v", err)
	}

	// We expect two instructions
	if len(prog) != 2 {
		t.Fatalf("Expected 2 instructions, got %d", len(prog))
	}

	// Check the first instruction is correct
	expected, _ := ParseInstruction("addi x1, x0, 42")
	if prog[0] != expected {
		t.Errorf("First instruction mismatch. Got %08x, want %08x", prog[0], expected)
	}
}
