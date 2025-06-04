package assembler

import (
    "testing"
)

func TestParseAdd(t *testing.T) {
    instr, err := ParseInstruction("add x3, x4, x5")
    if err != nil {
        t.Fatalf("ParseInstruction error: %v", err)
    }
    if instr.Opcode() != OPCODE_R_TYPE {
        t.Errorf("Opcode: got 0x%X, want 0x%X", instr.Opcode(), OPCODE_R_TYPE)
    }
    if instr.Rd() != 3 {
        t.Errorf("Rd: got %d, want 3", instr.Rd())
    }
    if instr.Rs1() != 4 {
        t.Errorf("Rs1: got %d, want 4", instr.Rs1())
    }
    if instr.Rs2() != 5 {
        t.Errorf("Rs2: got %d, want 5", instr.Rs2())
    }
    if instr.Funct3() != FUNCT3_ADD_SUB {
        t.Errorf("Funct3: got %d, want %d", instr.Funct3(), FUNCT3_ADD_SUB)
    }
    if instr.Funct7() != FUNCT7_ADD {
        t.Errorf("Funct7: got %d, want %d", instr.Funct7(), FUNCT7_ADD)
    }
}

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
