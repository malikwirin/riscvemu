package assembler

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

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
		assert.Equal(t, tc.want, uint32(tc.opcode), tc.name)
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
		{"FUNCT3_SLLI", FUNCT3_SLLI, 0x1},
		{"FUNCT3_SLT", FUNCT3_SLT, 0x2},
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
		assert.Equal(t, tc.want, tc.val, tc.name)
	}
}

func TestFunct7Constants(t *testing.T) {
	assert.Equal(t, uint32(0x00), FUNCT7_ADD, "FUNCT7_ADD")
	assert.Equal(t, uint32(0x20), FUNCT7_SUB, "FUNCT7_SUB")
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
		assert.Equal(t, tc.want, tc.opcode.String(), "Opcode("+tc.opcode.String()+").String()")
	}
}

func TestIsValidOpcode(t *testing.T) {
	validOpcodes := []Opcode{
		OPCODE_R_TYPE,
		OPCODE_I_TYPE,
		OPCODE_JALR,
		OPCODE_LOAD,
		OPCODE_STORE,
		OPCODE_BRANCH,
		OPCODE_JAL,
	}
	for _, op := range validOpcodes {
		assert.Equal(t, true, IsValidOpcode(op), "IsValidOpcode valid")
	}

	invalidOpcodes := []Opcode{
		0x0, 0x1, 0x2, 0x5, 0x7, 0x12, 0x14, 0x20, 0xF0, 0xFF, 0x80, 0xDEADBEEF,
	}
	for _, op := range invalidOpcodes {
		assert.Equal(t, false, IsValidOpcode(op), "IsValidOpcode invalid")
	}
}
