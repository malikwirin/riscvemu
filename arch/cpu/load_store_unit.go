package cpu

// LSU executes load and store operations. It bridges the CPU core to memory
// and counts down an internal latency per operation.
type LSU struct {
	// loadLatency is the configured execution time for a load in clock cycles.
	loadLatency int
	// storeLatency is the configured execution time for a store in clock cycles.
	storeLatency int

	// busy reports whether the LSU is currently executing an operation.
	busy bool
	// remain counts down from the chosen latency to zero; the operation
	// completes when the countdown hits zero.
	remain int
	// kind identifies whether this is a load (OpLOAD) or a store (OpSTORE).
	kind OpKind
	// vj is the resolved value of the base address register.
	vj uint32
	// vk is the resolved value of the data register (used by stores; ignored for loads).
	vk uint32
	// imm is the sign-extended immediate offset added to vj to form the effective address.
	imm int32
	// rd is the architectural destination register (set for loads, zero for stores).
	rd uint32
	// tag is the rename tag of the RS entry that was dispatched into this FU.
	tag RSTag
	// mem is the memory backend used for the actual load or store.
	mem WordHandler
	// justDone is set the cycle the countdown reaches zero and cleared on result consumption.
	justDone bool
	// resultValue holds the value read from memory (LOAD) or zero (STORE).
	resultValue uint32
	// pcOfBranch is unused for LSU (kept for symmetry with ALU).
	pcOfBranch uint32
}

func newLSU(loadLatency, storeLatency int) *LSU {
	return &LSU{
		loadLatency:  loadLatency,
		storeLatency: storeLatency,
	}
}

// AttachMemory binds the LSU to a memory backend.
func (l *LSU) AttachMemory(mem WordHandler) {
	l.mem = mem
}

// IsBusy reports whether the LSU is currently executing an operation.
func (l *LSU) IsBusy() bool { return l.busy }

// IsFree reports whether the LSU can accept a new operation. An LSU that has
// just finished is not free until its result has been consumed by writeback,
// otherwise a subsequent dispatch would clobber the pending justDone flag
// and the completed value would be lost.
func (l *LSU) IsFree() bool { return !l.busy && !l.justDone }

// Start dispatches a load or store into the LSU and arms the latency countdown.
func (l *LSU) Start(tag RSTag, kind OpKind, rd uint32, vj, vk uint32, imm int32) {
	l.StartAtPC(tag, kind, rd, vj, vk, imm, 0)
}

// StartAtPC behaves like Start but additionally records the program counter
// of the dispatched instruction.
func (l *LSU) StartAtPC(tag RSTag, kind OpKind, rd uint32, vj, vk uint32, imm int32, pc uint32) {
	l.busy = true
	l.tag = tag
	l.kind = kind
	l.rd = rd
	l.vj = vj
	l.vk = vk
	l.imm = imm
	l.pcOfBranch = pc
	l.justDone = false
	if kind == OpLOAD {
		l.remain = l.loadLatency
	} else {
		l.remain = l.storeLatency
	}
}

// Step advances the latency countdown by one cycle. When the countdown reaches
// zero the memory access is performed and the result becomes available.
func (l *LSU) Step() {
	if !l.busy {
		return
	}
	l.remain--
	if l.remain <= 0 {
		l.completeOperation()
	}
}

// completeOperation performs the actual load or store and marks the FU done.
func (l *LSU) completeOperation() {
	addr := l.vj + uint32(l.imm)
	if l.kind == OpLOAD {
		if l.mem != nil {
			value, err := l.mem.ReadWord(addr)
			if err == nil {
				l.resultValue = value
			}
		}
	} else {
		if l.mem != nil {
			_ = l.mem.WriteWord(addr, l.vk)
		}
		l.resultValue = 0
	}
	l.busy = false
	l.justDone = true
}

// TakeResult returns the produced value, tag, and destination register if the
// LSU finished this cycle. It clears the justDone flag so the same result is
// not consumed twice.
func (l *LSU) TakeResult() (RSTag, uint32, uint32, bool) {
	if !l.justDone {
		return NoTag, 0, 0, false
	}
	l.justDone = false
	return l.tag, l.resultValue, l.rd, true
}
