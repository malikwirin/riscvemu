package assembler

import "testing"

func TestInstructionRTypeFields(t *testing.T) {
    fields := []struct {
        name   string
        setter func(*Instruction, uint32)
        getter func(Instruction) uint32
        mask   uint32
    }{
        {"Opcode", func(i *Instruction, v uint32) { i.SetOpcode(v) }, func(i Instruction) uint32 { return i.Opcode() }, 0x7F},
        {"Rd",     func(i *Instruction, v uint32) { i.SetRd(v) },     func(i Instruction) uint32 { return i.Rd() },     0x1F},
        {"Funct3", func(i *Instruction, v uint32) { i.SetFunct3(v) }, func(i Instruction) uint32 { return i.Funct3() }, 0x7},
        {"Rs1",    func(i *Instruction, v uint32) { i.SetRs1(v) },    func(i Instruction) uint32 { return i.Rs1() },    0x1F},
        {"Rs2",    func(i *Instruction, v uint32) { i.SetRs2(v) },    func(i Instruction) uint32 { return i.Rs2() },    0x1F},
        {"Funct7", func(i *Instruction, v uint32) { i.SetFunct7(v) }, func(i Instruction) uint32 { return i.Funct7() }, 0x7F},
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
    inst.SetOpcode(0x33)
    inst.SetRd(5)
    inst.SetFunct3(0x0)
    inst.SetRs1(2)
    inst.SetRs2(3)
    inst.SetFunct7(0x20)

    if got := inst.Opcode(); got != 0x33 {
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

const (
    opcodeRType = 0x33
    opcodeIType = 0x13
    opcodeSType = 0x23
    opcodeBType = 0x63
    opcodeUType = 0x37
    opcodeJType = 0x6F
)

func TestInstructionType(t *testing.T) {
    cases := []struct {
        name     string
        opcode   uint32
        wantType string
    }{
        {"R-Type", opcodeRType, "R"},
        {"I-Type", opcodeIType, "I"},
        {"S-Type", opcodeSType, "S"},
        {"B-Type", opcodeBType, "B"},
        {"U-Type", opcodeUType, "U"},
        {"J-Type", opcodeJType, "J"},
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
