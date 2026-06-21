package cpu

import "github.com/malikwirin/riscvemu/assembler"

// mulDivKinds maps the RV32M funct3 field to its OpKind. Used by
// decode() for the MUL/DIV family (funct7=0x01).
var mulDivKinds = map[uint32]OpKind{
	assembler.FUNCT3_MUL:  OpMUL,
	assembler.FUNCT3_MULH: OpMULH,
	assembler.FUNCT3_DIV:  OpDIV,
	assembler.FUNCT3_DIVU: OpDIVU,
	assembler.FUNCT3_REM:  OpREM,
	assembler.FUNCT3_REMU: OpREMU,
}

// decode maps an encoded instruction word to its Tomasulo InstrMeta.
// Returns InstrMeta{Kind: OpInvalid} for unsupported opcodes.
func decode(instr assembler.Instruction) InstrMeta {
	switch instr.Opcode() {
	case assembler.OPCODE_R_TYPE:
		switch instr.Funct3() {
		case assembler.FUNCT3_ADD_SUB:
			switch instr.Funct7() {
			case assembler.FUNCT7_ADD:
				return InstrMeta{Kind: OpADD, Rd: instr.Rd(), Rs1: instr.Rs1(), Rs2: instr.Rs2()}
			case assembler.FUNCT7_SUB:
				return InstrMeta{Kind: OpSUB, Rd: instr.Rd(), Rs1: instr.Rs1(), Rs2: instr.Rs2()}
			}
		case assembler.FUNCT3_SLT:
			return InstrMeta{Kind: OpSLT, Rd: instr.Rd(), Rs1: instr.Rs1(), Rs2: instr.Rs2()}
		}
		// RV32M: funct7=0x01 selects the MUL/DIV family.
		if instr.Funct7() == assembler.FUNCT7_MULDIV {
			if k, ok := mulDivKinds[instr.Funct3()]; ok {
				return InstrMeta{Kind: k, Rd: instr.Rd(), Rs1: instr.Rs1(), Rs2: instr.Rs2()}
			}
		}
	case assembler.OPCODE_I_TYPE:
		switch instr.Funct3() {
		case assembler.FUNCT3_ADDI:
			return InstrMeta{Kind: OpADDI, Rd: instr.Rd(), Rs1: instr.Rs1(), Imm: instr.ImmI()}
		case assembler.FUNCT3_SLLI:
			return InstrMeta{Kind: OpSLLI, Rd: instr.Rd(), Rs1: instr.Rs1(), Imm: instr.ImmI()}
		}
	case assembler.OPCODE_LOAD:
		if instr.Funct3() == assembler.FUNCT3_LW {
			return InstrMeta{Kind: OpLOAD, Rd: instr.Rd(), Rs1: instr.Rs1(), Imm: instr.ImmI()}
		}
	case assembler.OPCODE_STORE:
		if instr.Funct3() == assembler.FUNCT3_SW {
			return InstrMeta{Kind: OpSTORE, Rs1: instr.Rs1(), Rs2: instr.Rs2(), Imm: instr.ImmS()}
		}
	case assembler.OPCODE_BRANCH:
		switch instr.Funct3() {
		case assembler.FUNCT3_BEQ:
			return InstrMeta{Kind: OpBEQ, Rs1: instr.Rs1(), Rs2: instr.Rs2(), Imm: instr.ImmB()}
		case assembler.FUNCT3_BNE:
			return InstrMeta{Kind: OpBNE, Rs1: instr.Rs1(), Rs2: instr.Rs2(), Imm: instr.ImmB()}
		case assembler.FUNCT3_SLT:
			return InstrMeta{Kind: OpBLT, Rs1: instr.Rs1(), Rs2: instr.Rs2(), Imm: instr.ImmB()}
		}
	case assembler.OPCODE_JAL:
		return InstrMeta{Kind: OpJAL, Rd: instr.Rd(), Imm: instr.ImmJ()}
	case assembler.OPCODE_JALR:
		return InstrMeta{Kind: OpJALR, Rd: instr.Rd(), Rs1: instr.Rs1(), Imm: instr.ImmI()}
	}
	return InstrMeta{Kind: OpInvalid}
}

// isLSUOp reports whether the operation belongs to the load/store unit.
func isLSUOp(k OpKind) bool {
	return k == OpLOAD || k == OpSTORE
}
