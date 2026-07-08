package cpu

import "fmt"

// OpKind classifies the Tomasulo operation kind for a decoded instruction.
type OpKind uint8

const (
	OpADD OpKind = iota + 1
	OpADDI
	OpSUB
	OpSLT
	OpSLLI
	OpLOAD
	OpSTORE
	OpBEQ
	OpBNE
	OpBLT
	OpJAL
	OpJALR
	// RV32M extension
	OpMUL
	OpMULH
	OpDIV
	OpDIVU
	OpREM
	OpREMU
)

// OpInvalid is the sentinel returned for unsupported or unknown opcodes.
const OpInvalid OpKind = 0

// String returns a short, human-readable name for the operation kind.
// Used by the CLI stats output to label FU utilisation rows.
func (k OpKind) String() string {
	switch k {
	case OpADD:
		return "ADD"
	case OpADDI:
		return "ADDI"
	case OpSUB:
		return "SUB"
	case OpSLT:
		return "SLT"
	case OpSLLI:
		return "SLLI"
	case OpLOAD:
		return "LOAD"
	case OpSTORE:
		return "STORE"
	case OpBEQ:
		return "BEQ"
	case OpBNE:
		return "BNE"
	case OpBLT:
		return "BLT"
	case OpJAL:
		return "JAL"
	case OpJALR:
		return "JALR"
	case OpMUL:
		return "MUL"
	case OpMULH:
		return "MULH"
	case OpDIV:
		return "DIV"
	case OpDIVU:
		return "DIVU"
	case OpREM:
		return "REM"
	case OpREMU:
		return "REMU"
	}
	return "INVALID"
}

// RSTag identifies a reservation station entry.
type RSTag uint32

// NoTag is the sentinel for an absent reservation station tag.
const NoTag RSTag = 0

// String returns a short, human-readable name for the
// reservation-station tag. Tags in the ALU namespace
// render as "ALU#N"; tags in the LSU namespace render as
// "LSU#N"; the NoTag sentinel renders as a dash.
func (t RSTag) String() string {
	const lsuBase = 1 << 16
	if t == NoTag {
		return "-"
	}
	if t >= lsuBase {
		return fmt.Sprintf("LSU#%d", int(t-lsuBase))
	}
	return fmt.Sprintf("ALU#%d", int(t))
}

// InstrMeta is the decoded view of an instruction used inside the CPU.
type InstrMeta struct {
	Kind OpKind
	Rd   uint32
	Rs1  uint32
	Rs2  uint32
	Imm  int32
}
