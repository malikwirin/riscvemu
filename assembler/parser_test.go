package assembler

import (
    "testing"
)

func TestParseAddi(t *testing.T) {
    instr, err := ParseInstruction("addi x1, x0, 5")
    if err != nil {
        t.Fatalf("ParseInstruction error: %v", err)
    }

    if instr.Opcode() != OPCODE_I_TYPE {
        t.Errorf("Opcode: got 0x%X, want 0x%X", instr.Opcode(), OPCODE_I_TYPE)
    }
    if instr.Rd() != 1 {
        t.Errorf("Rd: got %d, want 1", instr.Rd())
    }
    if instr.Rs1() != 0 {
        t.Errorf("Rs1: got %d, want 0", instr.Rs1())
    }
    if instr.Funct3() != FUNCT3_ADDI {
        t.Errorf("Funct3: got %d, want %d", instr.Funct3(), FUNCT3_ADDI)
    }
    gotImm := (uint32(instr) >> 20) & 0xFFF
    if gotImm != 5 {
        t.Errorf("Immediate: got %d, want 5", gotImm)
    }
}

func TestParseInvalidInstruction(t *testing.T) {
    _, err := ParseInstruction("foo x1, x2, x3")
    if err == nil {
        t.Error("Expected error for invalid instruction, got nil")
    }
}
