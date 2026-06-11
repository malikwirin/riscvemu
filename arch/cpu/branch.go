package cpu

// BranchInfo reports the outcome of the most recently resolved branch or
// jump instruction. The Machine uses it after RunCycle() to update the PC.
type BranchInfo struct {
    // IsBranch is true if the last retired instruction was a branch or jump.
    IsBranch bool
    // Taken is true when a conditional branch was taken, or always true
    // for unconditional jumps.
    Taken bool
    // Target is the new PC value the branch or jump should move to.
    // Only meaningful when IsBranch and Taken are both true.
    Target uint32
}

func isBranchKind(k OpKind) bool {
    switch k {
    case OpBEQ, OpBNE, OpBLT, OpJAL, OpJALR:
        return true
    }
    return false
}
