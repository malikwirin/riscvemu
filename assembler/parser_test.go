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

func TestParseBeq(t *testing.T) {
	// Example: beq x1, x2, 32
	instr, err := ParseInstruction("beq x1, x2, 32")
	if err != nil {
		t.Fatalf("ParseInstruction error: %v", err)
	}

	// Check opcode for BRANCH type
	if instr.Opcode() != OPCODE_BRANCH {
		t.Errorf("Opcode: got 0x%X, want 0x%X", instr.Opcode(), OPCODE_BRANCH)
	}
	// Check rs1
	if instr.Rs1() != 1 {
		t.Errorf("Rs1: got %d, want 1", instr.Rs1())
	}
	// Check rs2
	if instr.Rs2() != 2 {
		t.Errorf("Rs2: got %d, want 2", instr.Rs2())
	}
	// Check funct3 for BEQ
	if instr.Funct3() != FUNCT3_BEQ {
		t.Errorf("Funct3: got %d, want %d", instr.Funct3(), FUNCT3_BEQ)
	}

	// Immediate for B-type: imm[12|10:5] in bits 31|30:25, imm[4:1|11] in bits 11:8|7
	// For offset 32: 0b000000000100000
	imm := int32(32)
	// B-type encoding splits the immediate as follows:
	// imm[12]    -> bit 31
	// imm[10:5]  -> bits 30:25
	// imm[4:1]   -> bits 11:8
	// imm[11]    -> bit 7
	// The actual encoded immediate is shifted right by 1 (since lowest bit is always zero in RISC-V branches)

	imm12 := (uint32(imm) >> 12) & 0x1
	imm10_5 := (uint32(imm) >> 5) & 0x3F
	imm4_1 := (uint32(imm) >> 1) & 0xF
	imm11 := (uint32(imm) >> 11) & 0x1

	got_imm12 := (uint32(instr) >> 31) & 0x1
	got_imm10_5 := (uint32(instr) >> 25) & 0x3F
	got_imm4_1 := (uint32(instr) >> 8) & 0xF
	got_imm11 := (uint32(instr) >> 7) & 0x1

	if got_imm12 != imm12 {
		t.Errorf("Immediate (bit 12): got %d, want %d", got_imm12, imm12)
	}
	if got_imm10_5 != imm10_5 {
		t.Errorf("Immediate (bits 10:5): got %d, want %d", got_imm10_5, imm10_5)
	}
	if got_imm4_1 != imm4_1 {
		t.Errorf("Immediate (bits 4:1): got %d, want %d", got_imm4_1, imm4_1)
	}
	if got_imm11 != imm11 {
		t.Errorf("Immediate (bit 11): got %d, want %d", got_imm11, imm11)
	}
}

func TestParseBne(t *testing.T) {
	// Example: bne x4, x5, 64
	instr, err := ParseInstruction("bne x4, x5, 64")
	if err != nil {
		t.Fatalf("ParseInstruction error: %v", err)
	}

	// Check opcode for BRANCH type
	if instr.Opcode() != OPCODE_BRANCH {
		t.Errorf("Opcode: got 0x%X, want 0x%X", instr.Opcode(), OPCODE_BRANCH)
	}
	// Check rs1
	if instr.Rs1() != 4 {
		t.Errorf("Rs1: got %d, want 4", instr.Rs1())
	}
	// Check rs2
	if instr.Rs2() != 5 {
		t.Errorf("Rs2: got %d, want 5", instr.Rs2())
	}
	// Check funct3 for BNE
	if instr.Funct3() != FUNCT3_BNE {
		t.Errorf("Funct3: got %d, want %d", instr.Funct3(), FUNCT3_BNE)
	}

	// Immediate for B-type: imm[12|10:5] in bits 31|30:25, imm[4:1|11] in bits 11:8|7
	imm := int32(64)
	imm12 := (uint32(imm) >> 12) & 0x1
	imm10_5 := (uint32(imm) >> 5) & 0x3F
	imm4_1 := (uint32(imm) >> 1) & 0xF
	imm11 := (uint32(imm) >> 11) & 0x1

	got_imm12 := (uint32(instr) >> 31) & 0x1
	got_imm10_5 := (uint32(instr) >> 25) & 0x3F
	got_imm4_1 := (uint32(instr) >> 8) & 0xF
	got_imm11 := (uint32(instr) >> 7) & 0x1

	if got_imm12 != imm12 {
		t.Errorf("Immediate (bit 12): got %d, want %d", got_imm12, imm12)
	}
	if got_imm10_5 != imm10_5 {
		t.Errorf("Immediate (bits 10:5): got %d, want %d", got_imm10_5, imm10_5)
	}
	if got_imm4_1 != imm4_1 {
		t.Errorf("Immediate (bits 4:1): got %d, want %d", got_imm4_1, imm4_1)
	}
	if got_imm11 != imm11 {
		t.Errorf("Immediate (bit 11): got %d, want %d", got_imm11, imm11)
	}
}

func TestParseJal(t *testing.T) {
	// Example: jal x1, 2048
	instr, err := ParseInstruction("jal x1, 2048")
	if err != nil {
		t.Fatalf("ParseInstruction error: %v", err)
	}

	// Check opcode for JAL type
	if instr.Opcode() != OPCODE_JAL {
		t.Errorf("Opcode: got 0x%X, want 0x%X", instr.Opcode(), OPCODE_JAL)
	}
	// Check rd (destination register)
	if instr.Rd() != 1 {
		t.Errorf("Rd: got %d, want 1", instr.Rd())
	}

	// Immediate for J-type: imm[20|10:1|11|19:12] in bits 31|30:21|20|19:12
	imm := int32(2048)
	imm20 := (uint32(imm) >> 20) & 0x1
	imm10_1 := (uint32(imm) >> 1) & 0x3FF
	imm11 := (uint32(imm) >> 11) & 0x1
	imm19_12 := (uint32(imm) >> 12) & 0xFF

	got_imm20 := (uint32(instr) >> 31) & 0x1
	got_imm10_1 := (uint32(instr) >> 21) & 0x3FF
	got_imm11 := (uint32(instr) >> 20) & 0x1
	got_imm19_12 := (uint32(instr) >> 12) & 0xFF

	if got_imm20 != imm20 {
		t.Errorf("Immediate (bit 20): got %d, want %d", got_imm20, imm20)
	}
	if got_imm10_1 != imm10_1 {
		t.Errorf("Immediate (bits 10:1): got %d, want %d", got_imm10_1, imm10_1)
	}
	if got_imm11 != imm11 {
		t.Errorf("Immediate (bit 11): got %d, want %d", got_imm11, imm11)
	}
	if got_imm19_12 != imm19_12 {
		t.Errorf("Immediate (bits 19:12): got %d, want %d", got_imm19_12, imm19_12)
	}
}

func TestParseJalr(t *testing.T) {
	// Example: jalr x5, 0(x1)
	instr, err := ParseInstruction("jalr x5, 0(x1)")
	if err != nil {
		t.Fatalf("ParseInstruction error: %v", err)
	}

	if instr.Opcode() != OPCODE_JALR {
		t.Errorf("Opcode: got 0x%X, want 0x%X", instr.Opcode(), OPCODE_JALR)
	}
	if instr.Rd() != 5 {
		t.Errorf("Rd: got %d, want 5", instr.Rd())
	}
	if instr.Rs1() != 1 {
		t.Errorf("Rs1: got %d, want 1", instr.Rs1())
	}
	if instr.Funct3() != FUNCT3_JALR {
		t.Errorf("Funct3: got %d, want %d", instr.Funct3(), FUNCT3_JALR)
	}
	imm := uint32(0)
	gotImm := (uint32(instr) >> 20) & 0xFFF
	if gotImm != imm {
		t.Errorf("Immediate: got %d, want %d", gotImm, imm)
	}
}

func TestParseLw(t *testing.T) {
	// Example: lw x5, 16(x6)
	instr, err := ParseInstruction("lw x5, 16(x6)")
	if err != nil {
		t.Fatalf("ParseInstruction error: %v", err)
	}

	// Check opcode for LOAD type
	if instr.Opcode() != OPCODE_LOAD {
		t.Errorf("Opcode: got 0x%X, want 0x%X", instr.Opcode(), OPCODE_LOAD)
	}
	// Check destination register
	if instr.Rd() != 5 {
		t.Errorf("Rd: got %d, want 5", instr.Rd())
	}
	// Check base register
	if instr.Rs1() != 6 {
		t.Errorf("Rs1: got %d, want 6", instr.Rs1())
	}
	// Check funct3 for LW
	if instr.Funct3() != FUNCT3_LW {
		t.Errorf("Funct3: got %d, want %d", instr.Funct3(), FUNCT3_LW)
	}
	// Immediate for I-type: bits 20-31
	imm := uint32(16)
	gotImm := (uint32(instr) >> 20) & 0xFFF
	if gotImm != imm {
		t.Errorf("Immediate: got %d, want %d", gotImm, imm)
	}
}

func TestParseSlli(t *testing.T) {
	instr, err := ParseInstruction("slli x7, x8, 2")
	if err != nil {
		t.Fatalf("ParseInstruction error: %v", err)
	}
	if instr.Opcode() != OPCODE_I_TYPE {
		t.Errorf("Opcode: got 0x%X, want 0x%X", instr.Opcode(), OPCODE_I_TYPE)
	}
	if instr.Rd() != 7 {
		t.Errorf("Rd: got %d, want 7", instr.Rd())
	}
	if instr.Rs1() != 8 {
		t.Errorf("Rs1: got %d, want 8", instr.Rs1())
	}
	if instr.Funct3() != FUNCT3_SLLI {
		t.Errorf("Funct3: got %d, want %d", instr.Funct3(), FUNCT3_SLLI)
	}
	if (uint32(instr)>>20)&0x1F != 2 {
		t.Errorf("Shamt: got %d, want 2", (uint32(instr)>>20)&0x1F)
	}
	// Funct7 für slli ist 0x00
	if (uint32(instr)>>25)&0x7F != 0 {
		t.Errorf("Funct7: got %d, want 0", (uint32(instr)>>25)&0x7F)
	}
}

func TestParseSlt(t *testing.T) {
	instr, err := ParseInstruction("slt x10, x11, x12")
	if err != nil {
		t.Fatalf("ParseInstruction error: %v", err)
	}
	if instr.Opcode() != OPCODE_R_TYPE {
		t.Errorf("Opcode: got 0x%X, want 0x%X", instr.Opcode(), OPCODE_R_TYPE)
	}
	if instr.Rd() != 10 {
		t.Errorf("Rd: got %d, want 10", instr.Rd())
	}
	if instr.Rs1() != 11 {
		t.Errorf("Rs1: got %d, want 11", instr.Rs1())
	}
	if instr.Rs2() != 12 {
		t.Errorf("Rs2: got %d, want 12", instr.Rs2())
	}
	if instr.Funct3() != FUNCT3_SLT {
		t.Errorf("Funct3: got %d, want %d", instr.Funct3(), FUNCT3_SLT)
	}
	if instr.Funct7() != 0 {
		t.Errorf("Funct7: got %d, want 0", instr.Funct7())
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
