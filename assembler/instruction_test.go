package assembler

import "testing"

// Test all R-type instruction fields, including the special case for Opcode which may return OPCODE_INVALID for unknown opcodes.
func TestInstructionRTypeFields(t *testing.T) {
	fields := []struct {
		name   string
		setter func(*Instruction, uint32)
		getter func(Instruction) uint32
		mask   uint32
	}{
		{"Opcode",
			func(i *Instruction, v uint32) { i.SetOpcode(Opcode(v)) },
			func(i Instruction) uint32 { return uint32(i.Opcode()) },
			0x7F,
		},
		{"Rd",
			func(i *Instruction, v uint32) { i.SetRd(v) },
			func(i Instruction) uint32 { return i.Rd() },
			0x1F},
		{"Funct3",
			func(i *Instruction, v uint32) { i.SetFunct3(v) },
			func(i Instruction) uint32 { return i.Funct3() },
			0x7},
		{"Rs1",
			func(i *Instruction, v uint32) { i.SetRs1(v) },
			func(i Instruction) uint32 { return i.Rs1() },
			0x1F},
		{"Rs2",
			func(i *Instruction, v uint32) { i.SetRs2(v) },
			func(i Instruction) uint32 { return i.Rs2() },
			0x1F},
		{"Funct7",
			func(i *Instruction, v uint32) { i.SetFunct7(v) },
			func(i Instruction) uint32 { return i.Funct7() },
			0x7F},
	}
	for _, f := range fields {
		for try := uint32(0); try <= f.mask; try++ {
			var inst Instruction = 0
			f.setter(&inst, try)
			got := f.getter(inst)
			if f.name == "Opcode" {
				if IsValidOpcode(Opcode(try)) {
					checkField(t, f.name, got, try)
				} else {
					checkField(t, f.name, got, uint32(OPCODE_INVALID))
				}
			} else {
				checkField(t, f.name, got, try)
			}
		}
	}

	var inst Instruction
	inst.SetOpcode(OPCODE_R_TYPE)
	inst.SetRd(5)
	inst.SetFunct3(0x0)
	inst.SetRs1(2)
	inst.SetRs2(3)
	inst.SetFunct7(0x20)

	checkField(t, "Opcode", inst.Opcode(), OPCODE_R_TYPE)
	checkField(t, "Rd", inst.Rd(), uint32(5))
	checkField(t, "Funct3", inst.Funct3(), uint32(0))
	checkField(t, "Rs1", inst.Rs1(), uint32(2))
	checkField(t, "Rs2", inst.Rs2(), uint32(3))
	checkField(t, "Funct7", inst.Funct7(), uint32(0x20))
}

// Test the Type() method for several opcode cases, including an unknown opcode.
func TestInstructionType(t *testing.T) {
	cases := []struct {
		name     string
		opcode   Opcode
		wantType string
	}{
		{"R-Type", OPCODE_R_TYPE, "R"},
		{"I-Type", OPCODE_I_TYPE, "I"},
		{"S-Type", OPCODE_STORE, "S"},
		{"B-Type", OPCODE_BRANCH, "B"},
		{"J-Type", OPCODE_JAL, "J"},
		{"Unknown", 0x7F, "unknown"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var inst Instruction = Instruction(tc.opcode)
			got := inst.Type()
			checkField(t, "Type()", got, tc.wantType)
		})
	}
}

// Test I-type immediate encoding and decoding.
func TestInstructionITypeImmediate(t *testing.T) {
	var inst Instruction
	inst.SetImmI(0x7FF) // max positive 12bit
	checkField(t, "ImmI", inst.ImmI(), int32(0x7FF))
	inst.SetImmI(-1)
	checkField(t, "ImmI", inst.ImmI(), int32(-1))
	inst.SetImmI(-2048) // min negative 12bit
	checkField(t, "ImmI", inst.ImmI(), int32(-2048))
}

// Test S-type immediate encoding and decoding.
func TestInstructionSTypeImmediate(t *testing.T) {
	var inst Instruction
	inst.SetImmS(0x7FF)
	checkField(t, "ImmS", inst.ImmS(), int32(0x7FF))
	inst.SetImmS(-1)
	checkField(t, "ImmS", inst.ImmS(), int32(-1))
	inst.SetImmS(-2048)
	checkField(t, "ImmS", inst.ImmS(), int32(-2048))
}

// Test B-type immediate encoding and decoding.
func TestInstructionBTypeImmediate(t *testing.T) {
	var inst Instruction
	inst.SetImmB(0xFFE) // max positive even 13bit
	checkField(t, "ImmB", inst.ImmB(), int32(0xFFE))
	inst.SetImmB(-2)
	checkField(t, "ImmB", inst.ImmB(), int32(-2))
	inst.SetImmB(-4096)
	checkField(t, "ImmB", inst.ImmB(), int32(-4096))
}

// Test J-type immediate encoding and decoding.
func TestInstructionJTypeImmediate(t *testing.T) {
	var inst Instruction
	inst.SetImmJ(0xFFFFE)
	checkField(t, "ImmJ", inst.ImmJ(), int32(0xFFFFE))
	inst.SetImmJ(-2)
	checkField(t, "ImmJ", inst.ImmJ(), int32(-2))
	inst.SetImmJ(-1048576)
	checkField(t, "ImmJ", inst.ImmJ(), int32(-1048576))
}

// Test that Opcode() returns the correct value or OPCODE_INVALID for a variety of raw instruction values.
func TestInstruction_OpcodeReturnsExpectedValue(t *testing.T) {
	for v := uint32(0); v <= 0x7F; v++ {
		inst := Instruction(v)
		got := inst.Opcode()
		if IsValidOpcode(Opcode(v)) {
			checkField(t, "Opcode valid", got, Opcode(v))
		} else {
			checkField(t, "Opcode invalid", got, OPCODE_INVALID)
		}
	}
}

// Test that Opcode() returns OPCODE_INVALID for raw inputs that would result in unknown opcodes after masking.
func TestInstruction_OpcodeReturnsInvalidForUnknownOpcode(t *testing.T) {
	tests := []struct {
		name string
		raw  uint32
		want Opcode
	}{
		{"Known opcode (R-Type)", uint32(OPCODE_R_TYPE), OPCODE_R_TYPE},
		{"Known opcode (I-Type)", uint32(OPCODE_I_TYPE), OPCODE_I_TYPE},
		{"Known opcode (STORE)", uint32(OPCODE_STORE), OPCODE_STORE},
		{"Unknown opcode (0x00)", 0x00, OPCODE_INVALID},
		{"Unknown opcode (0x12)", 0x12, OPCODE_INVALID},
		{"Unknown opcode (0x7F)", 0x7F, OPCODE_INVALID},
		{"Random large value", 0xDEADBEEF, func() Opcode {
			op := Opcode(0xDEADBEEF & 0x7F)
			if IsValidOpcode(op) {
				return op
			}
			return OPCODE_INVALID
		}()},
	}
	for _, tc := range tests {
		inst := Instruction(tc.raw)
		got := inst.Opcode()
		checkField(t, tc.name, got, tc.want)
	}
}
