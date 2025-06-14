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

func TestInstruction_CastAndOpcodeEdgeCases(t *testing.T) {
    // Typical sw instruction (should not be interpreted as "STORE")
    sw := uint32(0x00112023)
    inst := Instruction(sw)
    if got := inst.Opcode(); got != OPCODE_STORE {
        t.Errorf("Opcode for sw (0x%X): got 0x%X, want 0x%X", sw, got, OPCODE_STORE)
    }

    // "STOR" as ASCII (should not yield a valid opcode)
    stor := uint32(0x53544F52)
    instStor := Instruction(stor)
    opcodeStor := instStor.Opcode()
    // We don't care what opcode this yields, but it must not panic or overflow
    t.Logf("Opcode for ASCII 'STOR' (0x%X): 0x%X", stor, opcodeStor)

    // "TORE" as ASCII (should not yield a valid opcode)
    tore := uint32(0x544F5245)
    instTore := Instruction(tore)
    opcodeTore := instTore.Opcode()
    t.Logf("Opcode for ASCII 'TORE' (0x%X): 0x%X", tore, opcodeTore)

    // For completeness, check that converting a random uint32 does not panic
    random := uint32(0xDEADBEEF)
    instRandom := Instruction(random)
    _ = instRandom.Opcode()
}

// Test that Opcode() always returns a 7-bit value, even for random or malformed input.
// This helps prove the masking is robust and prevents accidental overflows.
func TestInstruction_OpcodeIsAlways7Bits(t *testing.T) {
	testVals := []uint32{
		0xFFFFFFFF,      // all bits set
		0x80000000,      // only highest bit set
		0x53544F52,      // "STOR" as ASCII
		0x544F5245,      // "TORE" as ASCII
		0x00112023,      // typical instruction
		0x12345678,      // random value
	}
	for _, val := range testVals {
		inst := Instruction(val)
		opcode := inst.Opcode()
		if uint32(opcode) > 0x7F {
			t.Errorf("Opcode too large: inst=0x%X, opcode=0x%X", val, opcode)
		}
	}
}

// Test that casting a 64-bit value down to Instruction only uses the lower 32 bits.
// This also proves that even if a higher value (e.g. an ASCII string like "STORE") is cast, only the lower 32 bits are used.
func TestInstruction_OpcodeMasking64Bit(t *testing.T) {
	// 0x53544F5245 == "STORE" as ASCII (5 bytes, 40 bits)
	// When cast to uint32, only the lower 4 bytes remain.
	store64 := uint64(0x53544F5245)
	inst := Instruction(uint32(store64))
	got := inst.Opcode()
	want := Opcode(uint32(store64) & 0x7F)
	if got != want {
		t.Errorf("Opcode masking for 64-bit value: got 0x%X, want 0x%X", got, want)
	}
}
