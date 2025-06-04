package assembler

import "testing"

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
			if got != try {
				t.Errorf("%s: set %d, got %d", f.name, try, got)
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

	if got := inst.Opcode(); got != OPCODE_R_TYPE {
		t.Errorf("Opcode: expected 0x33, got 0x%X", got)
	}
	if got := inst.Rd(); got != 5 {
		t.Errorf("Rd: expected 5, got %d", got)
	}
	if got := inst.Funct3(); got != 0x0 {
		t.Errorf("Funct3: expected 0, got %d", got)
	}
	if got := inst.Rs1(); got != 2 {
		t.Errorf("Rs1: expected 2, got %d", got)
	}
	if got := inst.Rs2(); got != 3 {
		t.Errorf("Rs2: expected 3, got %d", got)
	}
	if got := inst.Funct7(); got != 0x20 {
		t.Errorf("Funct7: expected 0x20, got 0x%X", got)
	}
}

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
			if got != tc.wantType {
				t.Errorf("Type(): want %q, got %q (opcode=0x%02X)", tc.wantType, got, tc.opcode)
			}
		})
	}
}

func TestInstructionITypeImmediate(t *testing.T) {
	var inst Instruction
	inst.SetImmI(0x7FF) // max positive 12bit
	if got := inst.ImmI(); got != 0x7FF {
		t.Errorf("ImmI: expected 0x7FF, got 0x%X", got)
	}
	inst.SetImmI(-1)
	if got := inst.ImmI(); got != -1 {
		t.Errorf("ImmI: expected -1, got %d", got)
	}
	inst.SetImmI(-2048) // min negative 12bit
	if got := inst.ImmI(); got != -2048 {
		t.Errorf("ImmI: expected -2048, got %d", got)
	}
}

func TestInstructionSTypeImmediate(t *testing.T) {
	var inst Instruction
	inst.SetImmS(0x7FF)
	if got := inst.ImmS(); got != 0x7FF {
		t.Errorf("ImmS: expected 0x7FF, got 0x%X", got)
	}
	inst.SetImmS(-1)
	if got := inst.ImmS(); got != -1 {
		t.Errorf("ImmS: expected -1, got %d", got)
	}
	inst.SetImmS(-2048)
	if got := inst.ImmS(); got != -2048 {
		t.Errorf("ImmS: expected -2048, got %d", got)
	}
}

func TestInstructionBTypeImmediate(t *testing.T) {
	var inst Instruction
	inst.SetImmB(0xFFE) // max positive even 13bit
	if got := inst.ImmB(); got != 0xFFE {
		t.Errorf("ImmB: expected 0xFFE, got 0x%X", got)
	}
	inst.SetImmB(-2)
	if got := inst.ImmB(); got != -2 {
		t.Errorf("ImmB: expected -2, got %d", got)
	}
	inst.SetImmB(-4096)
	if got := inst.ImmB(); got != -4096 {
		t.Errorf("ImmB: expected -4096, got %d", got)
	}
}

func TestInstructionJTypeImmediate(t *testing.T) {
	var inst Instruction
	inst.SetImmJ(0xFFFFE)
	if got := inst.ImmJ(); got != 0xFFFFE {
		t.Errorf("ImmJ: expected 0xFFFFE, got 0x%X", got)
	}
	inst.SetImmJ(-2)
	if got := inst.ImmJ(); got != -2 {
		t.Errorf("ImmJ: expected -2, got %d", got)
	}
	inst.SetImmJ(-1048576)
	if got := inst.ImmJ(); got != -1048576 {
		t.Errorf("ImmJ: expected -1048576, got %d", got)
	}
}
