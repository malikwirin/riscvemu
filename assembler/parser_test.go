package assembler

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseInstructionTable(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		expectedOpcode Opcode
		expectedFields map[string]interface{}
		expectError    bool
	}{
		{
			name:           "ADD",
			input:          "add x3, x4, x5",
			expectedOpcode: OPCODE_R_TYPE,
			expectedFields: map[string]interface{}{
				"Rd":     uint32(3),
				"Rs1":    uint32(4),
				"Rs2":    uint32(5),
				"Funct3": FUNCT3_ADD_SUB,
				"Funct7": FUNCT7_ADD,
			},
		},
		{
			name:           "ADDI",
			input:          "addi x1, x0, 5",
			expectedOpcode: OPCODE_I_TYPE,
			expectedFields: map[string]interface{}{
				"Rd":     uint32(1),
				"Rs1":    uint32(0),
				"Funct3": FUNCT3_ADDI,
				"ImmI":   int32(5),
			},
		},
		{
			name:           "BEQ",
			input:          "beq x1, x2, 32",
			expectedOpcode: OPCODE_BRANCH,
			expectedFields: map[string]interface{}{
				"Rs1":    uint32(1),
				"Rs2":    uint32(2),
				"Funct3": FUNCT3_BEQ,
				"ImmB":   int32(32),
			},
		},
		{
			name:           "BLT",
			input:          "blt x6, x7, 128",
			expectedOpcode: OPCODE_BRANCH,
			expectedFields: map[string]interface{}{
				"Rs1":    uint32(6),
				"Rs2":    uint32(7),
				"Funct3": FUNCT3_BLT,
				"ImmB":   int32(128),
			},
		},
		{
			name:           "BNE",
			input:          "bne x4, x5, 64",
			expectedOpcode: OPCODE_BRANCH,
			expectedFields: map[string]interface{}{
				"Rs1":    uint32(4),
				"Rs2":    uint32(5),
				"Funct3": FUNCT3_BNE,
				"ImmB":   int32(64),
			},
		},
		{
			name:           "JAL",
			input:          "jal x1, 2048",
			expectedOpcode: OPCODE_JAL,
			expectedFields: map[string]interface{}{
				"Rd":   uint32(1),
				"ImmJ": int32(2048),
			},
		},
		{
			name:           "JALR",
			input:          "jalr x5, 0(x1)",
			expectedOpcode: OPCODE_JALR,
			expectedFields: map[string]interface{}{
				"Rd":     uint32(5),
				"Rs1":    uint32(1),
				"Funct3": FUNCT3_JALR,
				"ImmI":   int32(0),
			},
		},
		{
			name:           "LW",
			input:          "lw x5, 16(x6)",
			expectedOpcode: OPCODE_LOAD,
			expectedFields: map[string]interface{}{
				"Rd":     uint32(5),
				"Rs1":    uint32(6),
				"Funct3": FUNCT3_LW,
				"ImmI":   int32(16),
			},
		},
		{
			name:           "SLLI",
			input:          "slli x7, x8, 2",
			expectedOpcode: OPCODE_I_TYPE,
			expectedFields: map[string]interface{}{
				"Rd":     uint32(7),
				"Rs1":    uint32(8),
				"Funct3": FUNCT3_SLLI,
				"ImmI":   int32(2),
				"Funct7": uint32(0),
			},
		},
		{
			name:           "SLT",
			input:          "slt x10, x11, x12",
			expectedOpcode: OPCODE_R_TYPE,
			expectedFields: map[string]interface{}{
				"Rd":     uint32(10),
				"Rs1":    uint32(11),
				"Rs2":    uint32(12),
				"Funct3": FUNCT3_SLT,
				"Funct7": uint32(0),
			},
		},
		{
			name:           "SUB",
			input:          "sub x8, x6, x7",
			expectedOpcode: OPCODE_R_TYPE,
			expectedFields: map[string]interface{}{
				"Rd":     uint32(8),
				"Rs1":    uint32(6),
				"Rs2":    uint32(7),
				"Funct3": FUNCT3_ADD_SUB,
				"Funct7": FUNCT7_SUB,
			},
		},
		{
			name:           "SW",
			input:          "sw x7, 12(x8)",
			expectedOpcode: OPCODE_STORE,
			expectedFields: map[string]interface{}{
				"Rs1":    uint32(8),
				"Rs2":    uint32(7),
				"Funct3": FUNCT3_SW,
				"ImmS":   int32(12),
			},
		},
		{
			name:        "Invalid Instruction",
			input:       "foo x1, x2, x3",
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			instr, err := ParseInstruction(tc.input)

			if tc.expectError {
				assert.Error(t, err, "Expected an error for input: %q", tc.input)
				return
			}
			assert.NoError(t, err, "Unexpected error for input: %q", tc.input)

			assert.Equal(t, tc.expectedOpcode, instr.Opcode(), "Opcode mismatch for input: %q", tc.input)

			for field, expected := range tc.expectedFields {
				switch field {
				case "Rd":
					assert.Equal(t, expected, instr.Rd(), "Rd mismatch for input: %q", tc.input)
				case "Rs1":
					assert.Equal(t, expected, instr.Rs1(), "Rs1 mismatch for input: %q", tc.input)
				case "Rs2":
					assert.Equal(t, expected, instr.Rs2(), "Rs2 mismatch for input: %q", tc.input)
				case "Funct3":
					assert.Equal(t, expected, instr.Funct3(), "Funct3 mismatch for input: %q", tc.input)
				case "Funct7":
					assert.Equal(t, expected, instr.Funct7(), "Funct7 mismatch for input: %q", tc.input)
				case "ImmI":
					assert.Equal(t, expected, instr.ImmI(), "ImmI mismatch for input: %q", tc.input)
				case "ImmB":
					assert.Equal(t, expected, instr.ImmB(), "ImmB mismatch for input: %q", tc.input)
				case "ImmS":
					assert.Equal(t, expected, instr.ImmS(), "ImmS mismatch for input: %q", tc.input)
				case "ImmJ":
					assert.Equal(t, expected, instr.ImmJ(), "ImmJ mismatch for input: %q", tc.input)
				default:
					t.Fatalf("Unknown field: %s", field)
				}
			}
		})
	}
}

func TestParseInstruction_WhitespaceAndComments(t *testing.T) {
	cases := []struct {
		in   string
		want Instruction
	}{
		{"  addi x1, x0, 5   # comment", mustParse("addi x1, x0, 5")},
		{"\tadd x3,   x4, x5 ", mustParse("add x3, x4, x5")},
		{"addi x1, x0, 5 ; inline", mustParse("addi x1, x0, 5")},
	}
	for _, c := range cases {
		instr, err := ParseInstruction(removeCommentAndTrim(c.in))
		if err != nil {
			t.Errorf("ParseInstruction(%q) error: %v", c.in, err)
			continue
		}
		if instr != c.want {
			t.Errorf("ParseInstruction(%q) = 0x%08x, want 0x%08x", c.in, instr, c.want)
		}
	}
}
