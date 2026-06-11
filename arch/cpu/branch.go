package cpu

// BranchInfo reports the outcome of the most recently resolved branch or
// jump instruction. The Machine uses it after RunCycle() to update the PC
// and to write the link register for JAL/JALR.
type BranchInfo struct {
	// IsBranch is true if the last retired instruction was a branch or jump.
	IsBranch bool
	// Taken is true when a conditional branch was taken, or always true
	// for unconditional jumps.
	Taken bool
	// Target is the new PC value the branch or jump should move to.
	// Only meaningful when IsBranch and Taken are both true.
	Target uint32
	// LinkReg is the architectural destination register that JAL/JALR
	// writes its link value into. Zero for branches and for JAL x0.
	LinkReg uint32
	// LinkValue is the value written to LinkReg on JAL/JALR, typically
	// PC+4 of the jumping instruction. Zero for branches and for JAL x0.
	LinkValue uint32
}

func isBranchKind(k OpKind) bool {
	switch k {
	case OpBEQ, OpBNE, OpBLT, OpJAL, OpJALR:
		return true
	}
	return false
}
