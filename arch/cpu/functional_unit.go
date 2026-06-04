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
	// justDone is set the cycle the countdown reaches zero and cleared on result consumption.
	justDone bool
	// resultValue is the computed result, available once justDone is true.
	resultValue uint32
}

func newALU(latency int) *ALU {
	return &ALU{latency: latency}
}

// IsBusy reports whether the ALU is currently executing an instruction.
func (a *ALU) IsBusy() bool { return a.busy }

// Start dispatches a ready RS entry into the ALU and arms the latency countdown.
func (a *ALU) Start(tag RSTag, kind OpKind, rd uint32, vj, vk uint32, imm int32) {
	a.busy = true
	a.remain = a.latency
	a.tag = tag
	a.kind = kind
	a.rd = rd
	a.vj = vj
	a.vk = vk
	a.imm = imm
	a.justDone = false
}

// Step advances the latency countdown by one cycle and precomputes the result
// when the FU completes. The result becomes available via TakeResult.
func (a *ALU) Step() {
	if !a.busy {
		return
	}
	a.remain--
	if a.remain <= 0 {
		a.resultValue = a.compute()
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

// compute evaluates the integer operation. ALU only handles pure register-register
// and register-immediate forms; branches and loads are handled elsewhere.
func (a *ALU) compute() uint32 {
	switch a.kind {
	case OpADD:
		return a.vj + a.vk
	case OpADDI:
		return a.vj + uint32(a.imm)
	case OpSUB:
		return a.vj - a.vk
	case OpSLT:
		if int32(a.vj) < int32(a.vk) {
			return 1
		}
		return 0
	case OpSLLI:
		return a.vj << (a.imm & 0x1F)
	}
	return 0
}
