package assembler

import (
	"testing"
)

func TestAssembleFile_SimpleProgram(t *testing.T) {
	asm := "addi x1, x0, 42\naddi x2, x1, 1"
	filename := writeTempASM(t, asm)
	prog, err := AssembleFile(filename)
	if err != nil {
		t.Fatalf("AssembleFile returned error: %v", err)
	}
	checkInstructions(t, prog, []string{"addi x1, x0, 42", "addi x2, x1, 1"})
}

func TestAssembleFile_Example2asm(t *testing.T) {
	asm := `
addi x1, x0, 42
addi x2, x0, 100
sw x1, 0(x2)
lw x3, 0(x2)
`
	filename := writeTempASM(t, asm)
	prog, err := AssembleFile(filename)
	if err != nil {
		t.Fatalf("AssembleFile returned error: %v", err)
	}
	checkInstructions(t, prog, []string{
		"addi x1, x0, 42",
		"addi x2, x0, 100",
		"sw x1, 0(x2)",
		"lw x3, 0(x2)",
	})
}

func TestAssembleFile_LabelAndBranch(t *testing.T) {
	asm := `
start:  addi x1, x0, 1
        beq x1, x0, start
`
	filename := writeTempASM(t, asm)
	prog, err := AssembleFile(filename)
	if err != nil {
		t.Fatalf("AssembleFile returned error: %v", err)
	}
	checkInstructions(t, prog, []string{
		"addi x1, x0, 1",
		"beq x1, x0, -4",
	})
}

func TestAssembleFile_BLTInstruction(t *testing.T) {
	asm := `
start:  addi x1, x0, 1
        blt x1, x2, start
`
	filename := writeTempASM(t, asm)
	prog, err := AssembleFile(filename)
	if err != nil {
		t.Fatalf("AssembleFile returned error: %v", err)
	}
	checkInstructions(t, prog, []string{
		"addi x1, x0, 1",
		"blt x1, x2, -4",
	})
}

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
		{"blt x1, x0, loop", 1, "blt x1, x0, 4", false},
		{"beq x1, x0, 12", 2, "beq x1, x0, 12", false},
		{"blt x1, x0, missing", 0, "", true},
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

func TestParseLabelsAndInstructions_LabelOnOwnLine(t *testing.T) {
	lines := []string{
		"addi x1, x0, 5",
		"label_only:",
		"addi x2, x0, 9",
	}
	labelMap, instructions := parseLabelsAndInstructions(lines)
	wantInstr := []string{"addi x1, x0, 5", "addi x2, x0, 9"}
	if len(instructions) != len(wantInstr) {
		t.Fatalf("Expected %d instructions, got %d", len(wantInstr), len(instructions))
	}
	for i, instr := range wantInstr {
		if instructions[i] != instr {
			t.Errorf("Instruction %d mismatch: got %q, want %q", i, instructions[i], instr)
		}
	}
	addr, ok := labelMap["label_only"]
	if !ok {
		t.Errorf("Label 'label_only' not found in labelMap")
	}
	if addr != 1*INSTRUCTION_SIZE {
		t.Errorf("Label 'label_only' points to address %d, want %d", addr, 1*INSTRUCTION_SIZE)
	}
}

func TestPreprocessPseudoInstructions_Jump(t *testing.T) {
	got := preprocessPseudoInstructions("j end")
	want := "jal x0, end"
	if got != want {
		t.Errorf("preprocessPseudoInstructions(%q) = %q, want %q", "j end", got, want)
	}
	// Should not touch normal instructions
	normal := "addi x1, x0, 5"
	if preprocessPseudoInstructions(normal) != normal {
		t.Errorf("preprocessPseudoInstructions(%q) should be unchanged", normal)
	}
}
