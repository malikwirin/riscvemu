package assembler

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAssembleFile(t *testing.T) {
	tests := []struct {
		name string
		asm  string
		want []string
	}{
		{
			name: "Simple Program",
			asm:  "addi x1, x0, 42\naddi x2, x1, 1",
			want: []string{"addi x1, x0, 42", "addi x2, x1, 1"},
		},
		{
			name: "Example with SW and LW",
			asm: `
addi x1, x0, 42
addi x2, x0, 100
sw x1, 0(x2)
lw x3, 0(x2)
`,
			want: []string{"addi x1, x0, 42", "addi x2, x0, 100", "sw x1, 0(x2)", "lw x3, 0(x2)"},
		},
		{
			name: "Label and Branch",
			asm: `
start:  addi x1, x0, 1
        beq x1, x0, start
`,
			want: []string{"addi x1, x0, 1", "beq x1, x0, -4"},
		},
		{
			name: "BLT Instruction with Label",
			asm: `
start:  addi x1, x0, 1
        blt x1, x2, start
`,
			want: []string{"addi x1, x0, 1", "blt x1, x2, -4"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			filename := writeTempASM(t, tc.asm)
			prog, err := AssembleFile(filename)
			assert.NoError(t, err, "AssembleFile returned an error")
			checkInstructions(t, prog, tc.want)
		})
	}
}

func TestReplaceLabelOperandWithOffset(t *testing.T) {
	labelMap := map[string]int{
		"start": 0,
		"loop":  8,
	}
	tests := []struct {
		name      string
		line      string
		idx       int
		want      string
		shouldErr bool
	}{
		{"Valid BEQ with label", "beq x1, x0, start", 1, "beq x1, x0, -4", false},
		{"Valid BLT with label", "blt x1, x0, loop", 1, "blt x1, x0, 4", false},
		{"BEQ with numeric offset", "beq x1, x0, 12", 2, "beq x1, x0, 12", false},
		{"Missing label", "blt x1, x0, missing", 0, "", true},
		{"Instruction without label", "addi x1, x0, 5", 0, "addi x1, x0, 5", false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ReplaceLabelOperandWithOffset(tc.line, tc.idx, labelMap)
			if tc.shouldErr {
				assert.Error(t, err, "Expected an error")
			} else {
				assert.NoError(t, err, "Unexpected error")
				assert.Equal(t, tc.want, got, "ReplaceLabelOperandWithOffset returned unexpected result")
			}
		})
	}
}

func TestParseLabelsAndInstructions(t *testing.T) {
	lines := []string{
		"addi x1, x0, 5",
		"label_only:",
		"addi x2, x0, 9",
	}
	wantLabelMap := map[string]int{"label_only": 1 * INSTRUCTION_SIZE}
	wantInstructions := []string{"addi x1, x0, 5", "addi x2, x0, 9"}

	labelMap, instructions := parseLabelsAndInstructions(lines)
	assert.Equal(t, wantLabelMap, labelMap, "Label map mismatch")
	assert.Equal(t, wantInstructions, instructions, "Instructions mismatch")
}

func TestPreprocessPseudoInstructions(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"Valid jump pseudoinstruction", "j end", "jal x0, end"},
		{"Normal instruction untouched", "addi x1, x0, 5", "addi x1, x0, 5"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := preprocessPseudoInstructions(tc.in)
			assert.Equal(t, tc.want, got, "preprocessPseudoInstructions returned unexpected result")
		})
	}
}
