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

	// We expect two instructions
	if len(prog) != 2 {
		t.Fatalf("Expected 2 instructions, got %d", len(prog))
	}

	// First instruction: addi x1, x0, 1
	expected0, _ := ParseInstruction("addi x1, x0, 1")
	if prog[0] != expected0 {
		t.Errorf("First instruction mismatch. Got %08x, want %08x", prog[0], expected0)
	}

	// Second instruction: beq x1, x0, -4 (branch back to start)
	expected1, _ := ParseInstruction("beq x1, x0, -4")
	if prog[1] != expected1 {
		t.Errorf("Second instruction mismatch. Got %08x, want %08x", prog[1], expected1)
	}
}

func TestParseLabelsAndInstructions_Simple(t *testing.T) {
	src := []string{
		"start: addi x1, x0, 1",
		"loop: addi x2, x2, 1",
		"      beq x1, x2, loop",
		"end:  add x3, x1, x2",
	}
	labelMap, instructions := parseLabelsAndInstructions(src)

	// Check label addresses (should be instruction index * 4)
	expectedLabels := map[string]int{
		"start": 0,
		"loop":  4,
		"end":   12,
	}
	for label, wantAddr := range expectedLabels {
		got, ok := labelMap[label]
		if !ok {
			t.Errorf("Label %q not found", label)
			continue
		}
		if got != wantAddr {
			t.Errorf("Label %q: got %d, want %d", label, got, wantAddr)
		}
	}

	// Check instructions (labels should be stripped)
	wantInstr := []string{
		"addi x1, x0, 1",
		"addi x2, x2, 1",
		"beq x1, x2, loop",
		"add x3, x1, x2",
	}
	if len(instructions) != len(wantInstr) {
		t.Fatalf("Expected %d instructions, got %d", len(wantInstr), len(instructions))
	}
	for i, want := range wantInstr {
		if instructions[i] != want {
			t.Errorf("Instruction %d: got %q, want %q", i, instructions[i], want)
		}
	}
}
