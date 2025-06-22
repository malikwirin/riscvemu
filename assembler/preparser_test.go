package assembler

import (
	"os"
	"testing"
)

// Test that AssembleFile correctly assembles a simple program from a file.
func TestAssembleFile_SimpleProgram(t *testing.T) {
	asm := "addi x1, x0, 42\naddi x2, x1, 1"
	tmpfile, err := os.CreateTemp("", "test-*.asm")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpfile.Name())
	if _, err := tmpfile.WriteString(asm); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	tmpfile.Close()

	prog, err := AssembleFile(tmpfile.Name())
	if err != nil {
		t.Fatalf("AssembleFile returned error: %v", err)
	}
	if len(prog) != 2 {
		t.Fatalf("Expected 2 instructions, got %d", len(prog))
	}
	expected, _ := ParseInstruction("addi x1, x0, 42")
	if prog[0] != expected {
		t.Errorf("First instruction mismatch. Got %08x, want %08x", prog[0], expected)
	}
}

// Test that AssembleFile correctly assembles a more complex example program.
func TestAssembleFile_Example2asm(t *testing.T) {
	asm := `
addi x1, x0, 42
addi x2, x0, 100
sw x1, 0(x2)
lw x3, 0(x2)
`
	tmpfile, err := os.CreateTemp("", "test-ex2-*.asm")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpfile.Name())
	if _, err := tmpfile.WriteString(asm); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	tmpfile.Close()

	prog, err := AssembleFile(tmpfile.Name())
	if err != nil {
		t.Fatalf("AssembleFile returned error: %v", err)
	}
	if len(prog) != 4 {
		t.Fatalf("Expected 4 instructions, got %d", len(prog))
	}
	want := []string{
		"addi x1, x0, 42",
		"addi x2, x0, 100",
		"sw x1, 0(x2)",
		"lw x3, 0(x2)",
	}
	for i, line := range want {
		expected, err := ParseInstruction(line)
		if err != nil {
			t.Fatalf("ParseInstruction failed for %q: %v", line, err)
		}
		if prog[i] != expected {
			t.Errorf("Instruction %d mismatch. Got %08x, want %08x", i, prog[i], expected)
		}
	}
}

// Test AssembleFile with labels and branches, including label resolution.
func TestAssembleFile_LabelAndBranch(t *testing.T) {
	asm := `
start:  addi x1, x0, 1
        beq x1, x0, start
`
	tmpfile, err := os.CreateTemp("", "test-label-*.asm")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpfile.Name())
	if _, err := tmpfile.WriteString(asm); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	tmpfile.Close()

	prog, err := AssembleFile(tmpfile.Name())
	if err != nil {
		t.Fatalf("AssembleFile returned error: %v", err)
	}
	if len(prog) != 2 {
		t.Fatalf("Expected 2 instructions, got %d", len(prog))
	}
	expected0, _ := ParseInstruction("addi x1, x0, 1")
	if prog[0] != expected0 {
		t.Errorf("First instruction mismatch. Got %08x, want %08x", prog[0], expected0)
	}
	expected1, _ := ParseInstruction("beq x1, x0, -4")
	if prog[1] != expected1 {
		t.Errorf("Second instruction mismatch. Got %08x, want %08x", prog[1], expected1)
	}
}

// Test label replacement with offsets before instruction parsing.
func TestReplaceLabelOperandWithOffset_Preparse(t *testing.T) {
	labelMap := map[string]int{
		"start": 0,
		"loop":  8,
	}
	cases := []struct {
		line      string
		idx       int // instruction index (not byte address)
		want      string
		shouldErr bool
	}{
		{"beq x1, x0, start", 1, "beq x1, x0, -4", false},
		{"beq x1, x0, loop", 1, "beq x1, x0, 4", false},
		{"beq x1, x0, 12", 2, "beq x1, x0, 12", false},
		{"beq x1, x0, missing", 0, "", true},
		{"addi x1, x0, 5", 0, "addi x1, x0, 5", false},
	}
	for _, tc := range cases {
		got, err := ReplaceLabelOperandWithOffset(tc.line, tc.idx, labelMap)
		if tc.shouldErr {
			if err == nil {
				t.Errorf("Expected error for %q, got nil", tc.line)
			}
		} else {
			if err != nil {
				t.Errorf("Unexpected error for %q: %v", tc.line, err)
			}
			if got != tc.want {
				t.Errorf("ReplaceLabelOperandWithOffset(%q) = %q, want %q", tc.line, got, tc.want)
			}
		}
	}
}
