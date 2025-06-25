package assembler

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseAdd(t *testing.T) {
	instr, err := ParseInstruction("add x3, x4, x5")
	if err != nil {
		t.Fatalf("ParseInstruction error: %v", err)
	}
	assert.Equal(t, OPCODE_R_TYPE, instr.Opcode(), "Opcode")
	assert.Equal(t, uint32(3), instr.Rd(), "Rd")
	assert.Equal(t, uint32(4), instr.Rs1(), "Rs1")
	assert.Equal(t, uint32(5), instr.Rs2(), "Rs2")
	assert.Equal(t, FUNCT3_ADD_SUB, instr.Funct3(), "Funct3")
	assert.Equal(t, FUNCT7_ADD, instr.Funct7(), "Funct7")
}

func TestParseAddi(t *testing.T) {
	instr, err := ParseInstruction("addi x1, x0, 5")
	if err != nil {
		t.Fatalf("ParseInstruction error: %v", err)
	}
	assert.Equal(t, OPCODE_I_TYPE, instr.Opcode(), "Opcode")
	assert.Equal(t, uint32(1), instr.Rd(), "Rd")
	assert.Equal(t, uint32(0), instr.Rs1(), "Rs1")
	assert.Equal(t, FUNCT3_ADDI, instr.Funct3(), "Funct3")
	assert.Equal(t, int32(5), instr.ImmI(), "ImmI")
}

func TestParseBeq(t *testing.T) {
	instr, err := ParseInstruction("beq x1, x2, 32")
	if err != nil {
		t.Fatalf("ParseInstruction error: %v", err)
	}
	assert.Equal(t, OPCODE_BRANCH, instr.Opcode(), "Opcode")
	assert.Equal(t, uint32(1), instr.Rs1(), "Rs1")
	assert.Equal(t, uint32(2), instr.Rs2(), "Rs2")
	assert.Equal(t, FUNCT3_BEQ, instr.Funct3(), "Funct3")
	assert.Equal(t, int32(32), instr.ImmB(), "ImmB")
}

func TestParseBne(t *testing.T) {
	instr, err := ParseInstruction("bne x4, x5, 64")
	if err != nil {
		t.Fatalf("ParseInstruction error: %v", err)
	}
	assert.Equal(t, OPCODE_BRANCH, instr.Opcode(), "Opcode")
	assert.Equal(t, uint32(4), instr.Rs1(), "Rs1")
	assert.Equal(t, uint32(5), instr.Rs2(), "Rs2")
	assert.Equal(t, FUNCT3_BNE, instr.Funct3(), "Funct3")
	assert.Equal(t, int32(64), instr.ImmB(), "ImmB")
}

func TestParseJal(t *testing.T) {
	instr, err := ParseInstruction("jal x1, 2048")
	if err != nil {
		t.Fatalf("ParseInstruction error: %v", err)
	}
	assert.Equal(t, OPCODE_JAL, instr.Opcode(), "Opcode")
	assert.Equal(t, uint32(1), instr.Rd(), "Rd")
	assert.Equal(t, int32(2048), instr.ImmJ(), "ImmJ")
}

func TestParseJalr(t *testing.T) {
	instr, err := ParseInstruction("jalr x5, 0(x1)")
	if err != nil {
		t.Fatalf("ParseInstruction error: %v", err)
	}
	assert.Equal(t, OPCODE_JALR, instr.Opcode(), "Opcode")
	assert.Equal(t, uint32(5), instr.Rd(), "Rd")
	assert.Equal(t, uint32(1), instr.Rs1(), "Rs1")
	assert.Equal(t, FUNCT3_JALR, instr.Funct3(), "Funct3")
	assert.Equal(t, int32(0), instr.ImmI(), "ImmI")
}

func TestParseLw(t *testing.T) {
	instr, err := ParseInstruction("lw x5, 16(x6)")
	if err != nil {
		t.Fatalf("ParseInstruction error: %v", err)
	}
	assert.Equal(t, OPCODE_LOAD, instr.Opcode(), "Opcode")
	assert.Equal(t, uint32(5), instr.Rd(), "Rd")
	assert.Equal(t, uint32(6), instr.Rs1(), "Rs1")
	assert.Equal(t, FUNCT3_LW, instr.Funct3(), "Funct3")
	assert.Equal(t, int32(16), instr.ImmI(), "ImmI")
}

func TestParseSlli(t *testing.T) {
	instr, err := ParseInstruction("slli x7, x8, 2")
	if err != nil {
		t.Fatalf("ParseInstruction error: %v", err)
	}
	assert.Equal(t, OPCODE_I_TYPE, instr.Opcode(), "Opcode")
	assert.Equal(t, uint32(7), instr.Rd(), "Rd")
	assert.Equal(t, uint32(8), instr.Rs1(), "Rs1")
	assert.Equal(t, FUNCT3_SLLI, instr.Funct3(), "Funct3")
	// shamt is in ImmI for slli
	assert.Equal(t, int32(2), instr.ImmI(), "Shamt (ImmI)")
	assert.Equal(t, uint32(0), instr.Funct7(), "Funct7")
}

func TestParseSlt(t *testing.T) {
	instr, err := ParseInstruction("slt x10, x11, x12")
	if err != nil {
		t.Fatalf("ParseInstruction error: %v", err)
	}
	assert.Equal(t, OPCODE_R_TYPE, instr.Opcode(), "Opcode")
	assert.Equal(t, uint32(10), instr.Rd(), "Rd")
	assert.Equal(t, uint32(11), instr.Rs1(), "Rs1")
	assert.Equal(t, uint32(12), instr.Rs2(), "Rs2")
	assert.Equal(t, FUNCT3_SLT, instr.Funct3(), "Funct3")
	assert.Equal(t, uint32(0), instr.Funct7(), "Funct7")
}

func TestParseSub(t *testing.T) {
	instr, err := ParseInstruction("sub x8, x6, x7")
	if err != nil {
		t.Fatalf("ParseInstruction error: %v", err)
	}
	assert.Equal(t, OPCODE_R_TYPE, instr.Opcode(), "Opcode")
	assert.Equal(t, uint32(8), instr.Rd(), "Rd")
	assert.Equal(t, uint32(6), instr.Rs1(), "Rs1")
	assert.Equal(t, uint32(7), instr.Rs2(), "Rs2")
	assert.Equal(t, FUNCT3_ADD_SUB, instr.Funct3(), "Funct3")
	assert.Equal(t, FUNCT7_SUB, instr.Funct7(), "Funct7")
}

func TestParseSw(t *testing.T) {
	instr, err := ParseInstruction("sw x7, 12(x8)")
	if err != nil {
		t.Fatalf("ParseInstruction error: %v", err)
	}
	assert.Equal(t, OPCODE_STORE, instr.Opcode(), "Opcode")
	assert.Equal(t, uint32(8), instr.Rs1(), "Rs1")
	assert.Equal(t, uint32(7), instr.Rs2(), "Rs2")
	assert.Equal(t, FUNCT3_SW, instr.Funct3(), "Funct3")
	assert.Equal(t, int32(12), instr.ImmS(), "ImmS")
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

func TestParseInstruction_WhitespaceAndComments(t *testing.T) {
	cases := []struct {
		in   string
		want Instruction
	}{
		{"  addi x1, x0, 5   # comment", mustParse("addi x1, x0, 5")},
		{"\tadd x3,   x4, x5 ", mustParse("add x3, x4, x5")},
		{"addi x1, x0, 5 ; inline", mustParse("addi x1, x0, 5")},
	}
	for _, c := range cases {
		instr, err := ParseInstruction(removeCommentAndTrim(c.in))
		if err != nil {
			t.Errorf("ParseInstruction(%q) error: %v", c.in, err)
			continue
		}
		if instr != c.want {
			t.Errorf("ParseInstruction(%q) = 0x%08x, want 0x%08x", c.in, instr, c.want)
		}
	}
}
