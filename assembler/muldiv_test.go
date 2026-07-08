package assembler

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestParseMULAndDIV covers the RV32M extension. All six instructions
// share OPCODE_R_TYPE with funct7=0x01; the funct3 field selects the
// specific operation.
func TestParseMULAndDIV(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		funct3   uint32
		expected map[string]interface{}
	}{
		{
			name:   "MUL",
			input:  "mul x3, x4, x5",
			funct3: FUNCT3_MUL,
			expected: map[string]interface{}{
				"Rd":     uint32(3),
				"Rs1":    uint32(4),
				"Rs2":    uint32(5),
				"Funct3": FUNCT3_MUL,
				"Funct7": FUNCT7_MULDIV,
			},
		},
		{
			name:   "MULH",
			input:  "mulh x3, x4, x5",
			funct3: FUNCT3_MULH,
			expected: map[string]interface{}{
				"Rd":     uint32(3),
				"Rs1":    uint32(4),
				"Rs2":    uint32(5),
				"Funct3": FUNCT3_MULH,
				"Funct7": FUNCT7_MULDIV,
			},
		},
		{
			name:   "DIV",
			input:  "div x3, x4, x5",
			funct3: FUNCT3_DIV,
			expected: map[string]interface{}{
				"Rd":     uint32(3),
				"Rs1":    uint32(4),
				"Rs2":    uint32(5),
				"Funct3": FUNCT3_DIV,
				"Funct7": FUNCT7_MULDIV,
			},
		},
		{
			name:   "DIVU",
			input:  "divu x3, x4, x5",
			funct3: FUNCT3_DIVU,
			expected: map[string]interface{}{
				"Rd":     uint32(3),
				"Rs1":    uint32(4),
				"Rs2":    uint32(5),
				"Funct3": FUNCT3_DIVU,
				"Funct7": FUNCT7_MULDIV,
			},
		},
		{
			name:   "REM",
			input:  "rem x3, x4, x5",
			funct3: FUNCT3_REM,
			expected: map[string]interface{}{
				"Rd":     uint32(3),
				"Rs1":    uint32(4),
				"Rs2":    uint32(5),
				"Funct3": FUNCT3_REM,
				"Funct7": FUNCT7_MULDIV,
			},
		},
		{
			name:   "REMU",
			input:  "remu x3, x4, x5",
			funct3: FUNCT3_REMU,
			expected: map[string]interface{}{
				"Rd":     uint32(3),
				"Rs1":    uint32(4),
				"Rs2":    uint32(5),
				"Funct3": FUNCT3_REMU,
				"Funct7": FUNCT7_MULDIV,
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			instr, err := ParseInstruction(tc.input)
			assert.NoError(t, err)
			assert.Equal(t, OPCODE_R_TYPE, instr.Opcode(), "opcode")
			assert.Equal(t, tc.expected["Rd"], instr.Rd(), "Rd")
			assert.Equal(t, tc.expected["Rs1"], instr.Rs1(), "Rs1")
			assert.Equal(t, tc.expected["Rs2"], instr.Rs2(), "Rs2")
			assert.Equal(t, tc.expected["Funct3"], instr.Funct3(), "Funct3")
			assert.Equal(t, tc.expected["Funct7"], instr.Funct7(), "Funct7")
		})
	}
}
