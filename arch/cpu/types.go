package cpu

// OpKind classifies the Tomasulo operation kind for a decoded instruction.
type OpKind uint8

const (
	OpADD OpKind = iota + 1
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
)

// OpInvalid is the sentinel returned for unsupported or unknown opcodes.
const OpInvalid OpKind = 0

// RSTag identifies a reservation station entry.
type RSTag uint32

// NoTag is the sentinel for an absent reservation station tag.
const NoTag RSTag = 0

// InstrMeta is the decoded view of an instruction used inside the CPU.
type InstrMeta struct {
	Kind OpKind
	Rd   uint32
	Rs1  uint32
	Rs2  uint32
	Imm  int32
}
