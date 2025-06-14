package assembler

import (
	"testing"
)

func checkField(t *testing.T, name string, got, want interface{}) {
	if got != want {
		t.Errorf("%s: got %v, want %v", name, got, want)
	}
}

func TestParseAdd(t *testing.T) {
	instr, err := ParseInstruction("add x3, x4, x5")
	if err != nil {
		t.Fatalf("ParseInstruction error: %v", err)
	}
	checkField(t, "Opcode", instr.Opcode(), OPCODE_R_TYPE)
	checkField(t, "Rd", instr.Rd(), uint32(3))
	checkField(t, "Rs1", instr.Rs1(), uint32(4))
	checkField(t, "Rs2", instr.Rs2(), uint32(5))
	checkField(t, "Funct3", instr.Funct3(), FUNCT3_ADD_SUB)
	checkField(t, "Funct7", instr.Funct7(), FUNCT7_ADD)
}

func TestParseAddi(t *testing.T) {
	instr, err := ParseInstruction("addi x1, x0, 5")
	if err != nil {
		t.Fatalf("ParseInstruction error: %v", err)
	}
	checkField(t, "Opcode", instr.Opcode(), OPCODE_I_TYPE)
	checkField(t, "Rd", instr.Rd(), uint32(1))
	checkField(t, "Rs1", instr.Rs1(), uint32(0))
	checkField(t, "Funct3", instr.Funct3(), FUNCT3_ADDI)
	checkField(t, "ImmI", instr.ImmI(), int32(5))
}

func TestParseBeq(t *testing.T) {
	instr, err := ParseInstruction("beq x1, x2, 32")
	if err != nil {
		t.Fatalf("ParseInstruction error: %v", err)
	}
	checkField(t, "Opcode", instr.Opcode(), OPCODE_BRANCH)
	checkField(t, "Rs1", instr.Rs1(), uint32(1))
	checkField(t, "Rs2", instr.Rs2(), uint32(2))
	checkField(t, "Funct3", instr.Funct3(), FUNCT3_BEQ)
	checkField(t, "ImmB", instr.ImmB(), int32(32))
}

func TestParseBne(t *testing.T) {
	instr, err := ParseInstruction("bne x4, x5, 64")
	if err != nil {
		t.Fatalf("ParseInstruction error: %v", err)
	}
	checkField(t, "Opcode", instr.Opcode(), OPCODE_BRANCH)
	checkField(t, "Rs1", instr.Rs1(), uint32(4))
	checkField(t, "Rs2", instr.Rs2(), uint32(5))
	checkField(t, "Funct3", instr.Funct3(), FUNCT3_BNE)
	checkField(t, "ImmB", instr.ImmB(), int32(64))
}

func TestParseJal(t *testing.T) {
	instr, err := ParseInstruction("jal x1, 2048")
	if err != nil {
		t.Fatalf("ParseInstruction error: %v", err)
	}
	checkField(t, "Opcode", instr.Opcode(), OPCODE_JAL)
	checkField(t, "Rd", instr.Rd(), uint32(1))
	checkField(t, "ImmJ", instr.ImmJ(), int32(2048))
}

func TestParseJalr(t *testing.T) {
	instr, err := ParseInstruction("jalr x5, 0(x1)")
	if err != nil {
		t.Fatalf("ParseInstruction error: %v", err)
	}
	checkField(t, "Opcode", instr.Opcode(), OPCODE_JALR)
	checkField(t, "Rd", instr.Rd(), uint32(5))
	checkField(t, "Rs1", instr.Rs1(), uint32(1))
	checkField(t, "Funct3", instr.Funct3(), FUNCT3_JALR)
	checkField(t, "ImmI", instr.ImmI(), int32(0))
}

func TestParseLw(t *testing.T) {
	instr, err := ParseInstruction("lw x5, 16(x6)")
	if err != nil {
		t.Fatalf("ParseInstruction error: %v", err)
	}
	checkField(t, "Opcode", instr.Opcode(), OPCODE_LOAD)
	checkField(t, "Rd", instr.Rd(), uint32(5))
	checkField(t, "Rs1", instr.Rs1(), uint32(6))
	checkField(t, "Funct3", instr.Funct3(), FUNCT3_LW)
	checkField(t, "ImmI", instr.ImmI(), int32(16))
}

func TestParseSlli(t *testing.T) {
	instr, err := ParseInstruction("slli x7, x8, 2")
	if err != nil {
		t.Fatalf("ParseInstruction error: %v", err)
	}
	checkField(t, "Opcode", instr.Opcode(), OPCODE_I_TYPE)
	checkField(t, "Rd", instr.Rd(), uint32(7))
	checkField(t, "Rs1", instr.Rs1(), uint32(8))
	checkField(t, "Funct3", instr.Funct3(), FUNCT3_SLLI)
	// shamt is in ImmI for slli
	checkField(t, "Shamt (ImmI)", instr.ImmI(), int32(2))
	checkField(t, "Funct7", instr.Funct7(), uint32(0))
}

func TestParseSlt(t *testing.T) {
	instr, err := ParseInstruction("slt x10, x11, x12")
	if err != nil {
		t.Fatalf("ParseInstruction error: %v", err)
	}
	checkField(t, "Opcode", instr.Opcode(), OPCODE_R_TYPE)
	checkField(t, "Rd", instr.Rd(), uint32(10))
	checkField(t, "Rs1", instr.Rs1(), uint32(11))
	checkField(t, "Rs2", instr.Rs2(), uint32(12))
	checkField(t, "Funct3", instr.Funct3(), FUNCT3_SLT)
	checkField(t, "Funct7", instr.Funct7(), uint32(0))
}

func TestParseSub(t *testing.T) {
	instr, err := ParseInstruction("sub x8, x6, x7")
	if err != nil {
		t.Fatalf("ParseInstruction error: %v", err)
	}
	checkField(t, "Opcode", instr.Opcode(), OPCODE_R_TYPE)
	checkField(t, "Rd", instr.Rd(), uint32(8))
	checkField(t, "Rs1", instr.Rs1(), uint32(6))
	checkField(t, "Rs2", instr.Rs2(), uint32(7))
	checkField(t, "Funct3", instr.Funct3(), FUNCT3_ADD_SUB)
	checkField(t, "Funct7", instr.Funct7(), FUNCT7_SUB)
}

func TestParseSw(t *testing.T) {
	instr, err := ParseInstruction("sw x7, 12(x8)")
	if err != nil {
		t.Fatalf("ParseInstruction error: %v", err)
	}
	checkField(t, "Opcode", instr.Opcode(), OPCODE_STORE)
	checkField(t, "Rs1", instr.Rs1(), uint32(8))
	checkField(t, "Rs2", instr.Rs2(), uint32(7))
	checkField(t, "Funct3", instr.Funct3(), FUNCT3_SW)
	checkField(t, "ImmS", instr.ImmS(), int32(12))
}

func TestParseInstruction_SW(t *testing.T) {
    instr, err := ParseInstruction("sw x1, 0(x2)")
    if err != nil {
        t.Fatalf("ParseInstruction: %v", err)
    }
    t.Logf("sw: 0x%08x", uint32(instr))
    if instr == 0x53544F52 || instr == 0x544F5245 {
        t.Fatal("Assembler returned ASCII STORE-like, not a valid Instruction")
    }
}

func TestParseInvalidInstruction(t *testing.T) {
	_, err := ParseInstruction("foo x1, x2, x3")
	if err == nil {
		t.Error("Expected error for invalid instruction, got nil")
	}
}
