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

func TestParseSw(t *testing.T) {
    // Example: sw x7, 12(x8)
    instr, err := ParseInstruction("sw x7, 12(x8)")
    if err != nil {
        t.Fatalf("ParseInstruction error: %v", err)
    }

    // Check opcode for STORE type
    if instr.Opcode() != OPCODE_STORE {
        t.Errorf("Opcode: got 0x%X, want 0x%X", instr.Opcode(), OPCODE_STORE)
    }
    // Check rs1 (base register)
    if instr.Rs1() != 8 {
        t.Errorf("Rs1: got %d, want 8", instr.Rs1())
    }
    // Check rs2 (source register)
    if instr.Rs2() != 7 {
        t.Errorf("Rs2: got %d, want 7", instr.Rs2())
    }
    // Check funct3 for SW
    if instr.Funct3() != FUNCT3_SW {
        t.Errorf("Funct3: got %d, want %d", instr.Funct3(), FUNCT3_SW)
    }

    // Immediate for S-type: bits 11:5 in bits 25:31, bits 4:0 in bits 7:11
    imm := uint32(12)
    imm_low := (imm >> 0) & 0x1F  // bits 0-4 -> bits 7-11
    imm_high := (imm >> 5) & 0x7F // bits 5-11 -> bits 25-31

    got_imm_low := (uint32(instr) >> 7) & 0x1F
    got_imm_high := (uint32(instr) >> 25) & 0x7F

    if got_imm_low != imm_low {
        t.Errorf("Immediate (low bits): got %d, want %d", got_imm_low, imm_low)
    }
    if got_imm_high != imm_high {
        t.Errorf("Immediate (high bits): got %d, want %d", got_imm_high, imm_high)
    }
}

func TestParseInvalidInstruction(t *testing.T) {
    _, err := ParseInstruction("foo x1, x2, x3")
    if err == nil {
        t.Error("Expected error for invalid instruction, got nil")
    }
}
