package cpu

// OpKind classifies the Tomasulo operation kind for a decoded instruction.
type OpKind uint8

const (
	OpADD OpKind = iota
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

// RSTag identifies a reservation station entry.
type RSTag uint32

// NoTag is the sentinel for an absent reservation station tag.
const NoTag RSTag = 0
