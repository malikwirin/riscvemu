package cpu

// ALU executes integer arithmetic and logical operations.
type ALU struct {
    // latency is the configured execution time in clock cycles.
    latency int
    // busy reports whether the ALU is currently producing a result.
    busy bool
    // remain counts down from latency to zero; the result is ready when it hits zero.
    remain int
    // tag is the rename tag of the RS entry that was dispatched into this FU.
    tag RSTag
    // kind is the operation kind being computed.
    kind OpKind
    // vj is the resolved value of the first source operand.
    vj uint32
    // vk is the resolved value of the second source operand.
    vk uint32
    // imm carries the immediate value for I-type operations.
    imm int32
    // rd is the architectural destination register of the in-flight operation.
    rd uint32
    // pcOfBranch is the program counter of the branch instruction; used
    // to compute the target when the branch is taken.
    pcOfBranch uint32
    // justDone is set the cycle the countdown reaches zero and cleared on result consumption.
    justDone bool
    // resultValue is the computed result, available once justDone is true.
    resultValue uint32
    // branchTaken is set when the completed operation is a taken branch or jump.
    branchTaken bool
}

func newALU(latency int) *ALU {
	return &ALU{latency: latency}
}

// IsBusy reports whether the ALU is currently executing an instruction.
func (a *ALU) IsBusy() bool { return a.busy }

// Start dispatches a ready RS entry into the ALU and arms the latency countdown.
func (a *ALU) Start(tag RSTag, kind OpKind, rd uint32, vj, vk uint32, imm int32) {
    a.StartAtPC(tag, kind, rd, vj, vk, imm, 0)
}

// StartAtPC behaves like Start but additionally records the program counter
// of the dispatched instruction so branch resolution can compute the target.
func (a *ALU) StartAtPC(tag RSTag, kind OpKind, rd uint32, vj, vk uint32, imm int32, pc uint32) {
    a.busy = true
    a.remain = a.latency
    a.tag = tag
    a.kind = kind
    a.rd = rd
    a.vj = vj
    a.vk = vk
    a.imm = imm
    a.pcOfBranch = pc
    a.justDone = false
    a.branchTaken = false
}

// Step advances the latency countdown by one cycle and precomputes the result
// when the FU completes. The result becomes available via TakeResult.
func (a *ALU) Step() {
    if !a.busy {
        return
    }
    a.remain--
    if a.remain <= 0 {
        a.resultValue, a.branchTaken = a.compute()
        a.busy = false
        a.justDone = true
    }
}

// TakeResult returns the produced value, tag, and destination register if the
// ALU finished this cycle. It clears the justDone flag so the same result is
// not consumed twice.
func (a *ALU) TakeResult() (RSTag, uint32, uint32, bool) {
    if !a.justDone {
        return NoTag, 0, 0, false
    }
    a.justDone = false
    return a.tag, a.resultValue, a.rd, true
}

// IsBranchTaken reports whether the most recently completed operation was a
// taken branch or jump.
func (a *ALU) IsBranchTaken() bool {
    return a.branchTaken
}

// Kind returns the operation kind of the most recently completed instruction
// (or the one currently in flight).
func (a *ALU) Kind() OpKind {
    return a.kind
}

// BranchTarget returns the resolved target PC for the most recently completed
// branch. Only meaningful when IsBranchTaken() is true.
func (a *ALU) BranchTarget() uint32 {
    return uint32(int32(a.pcOfBranch) + a.imm)
}

// compute evaluates the integer operation and reports whether the result
// is a taken branch. For non-branch operations the boolean is false and
// the result is the arithmetic value. For branches and jumps the result
// is undefined and only the boolean is meaningful.
func (a *ALU) compute() (uint32, bool) {
    switch a.kind {
    case OpADD:
        return a.vj + a.vk, false
    case OpADDI:
        return a.vj + uint32(a.imm), false
    case OpSUB:
        return a.vj - a.vk, false
    case OpSLT:
        if int32(a.vj) < int32(a.vk) {
            return 1, false
        }
        return 0, false
    case OpSLLI:
        return a.vj << (a.imm & 0x1F), false
    case OpBEQ:
        return 0, a.vj == a.vk
    case OpBNE:
        return 0, a.vj != a.vk
    case OpBLT:
        return 0, int32(a.vj) < int32(a.vk)
    }
    return 0, false
}
