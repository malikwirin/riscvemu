package assembler

import "testing"

func TestOpcodeConstants(t *testing.T) {
	cases := []struct {
		name   string
		opcode Opcode
		want   uint32
	}{
		{"OPCODE_R_TYPE", OPCODE_R_TYPE, 0x33},
		{"OPCODE_I_TYPE", OPCODE_I_TYPE, 0x13},
		{"OPCODE_JALR", OPCODE_JALR, 0x67},
		{"OPCODE_LOAD", OPCODE_LOAD, 0x03},
		{"OPCODE_STORE", OPCODE_STORE, 0x23},
		{"OPCODE_BRANCH", OPCODE_BRANCH, 0x63},
		{"OPCODE_JAL", OPCODE_JAL, 0x6F},
	}

	for _, tc := range cases {
		if uint32(tc.opcode) != tc.want {
			t.Errorf("%s: expected 0x%X, got 0x%X", tc.name, tc.want, tc.opcode)
		}
	}
}

func TestFunct3Constants(t *testing.T) {
	consts := []struct {
		name string
		val  uint32
		want uint32
	}{
		{"FUNCT3_ADD_SUB", FUNCT3_ADD_SUB, 0x0},
		{"FUNCT3_AND", FUNCT3_AND, 0x7},
		{"FUNCT3_OR", FUNCT3_OR, 0x6},
		{"FUNCT3_XOR", FUNCT3_XOR, 0x4},
		{"FUNCT3_BEQ", FUNCT3_BEQ, 0x0},
		{"FUNCT3_BNE", FUNCT3_BNE, 0x1},
		{"FUNCT3_LW", FUNCT3_LW, 0x2},
		{"FUNCT3_SW", FUNCT3_SW, 0x2},
		{"FUNCT3_ADDI", FUNCT3_ADDI, 0x0},
		{"FUNCT3_ANDI", FUNCT3_ANDI, 0x7},
		{"FUNCT3_ORI", FUNCT3_ORI, 0x6},
		{"FUNCT3_JALR", FUNCT3_JALR, 0x0},
	}

	for _, tc := range consts {
		if tc.val != tc.want {
			t.Errorf("%s: expected 0x%X, got 0x%X", tc.name, tc.want, tc.val)
		}
	}
}

func TestFunct7Constants(t *testing.T) {
	if FUNCT7_ADD != 0x00 {
		t.Errorf("FUNCT7_ADD: expected 0x00, got 0x%X", FUNCT7_ADD)
	}
	if FUNCT7_SUB != 0x20 {
		t.Errorf("FUNCT7_SUB: expected 0x20, got 0x%X", FUNCT7_SUB)
	}
}

func TestOpcodeStringer(t *testing.T) {
	tests := []struct {
		opcode Opcode
		want   string
	}{
		{OPCODE_R_TYPE, "R-Type"},
		{OPCODE_I_TYPE, "I-Type"},
		{OPCODE_JALR, "JALR"},
		{Opcode(0xFF), "Unknown(0xFF)"},
	}
	for _, tc := range tests {
		got := tc.opcode.String()
		if got != tc.want {
			t.Errorf("Opcode(%#x).String(): want %q, got %q", tc.opcode, tc.want, got)
		}
	}
}
