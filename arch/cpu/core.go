package cpu

// WordHandler is the memory interface required by the load/store unit.
type WordHandler interface {
	ReadWord(addr uint32) (uint32, error)
	WriteWord(addr uint32, value uint32) error
}

// CPU is the Tomasulo-style execution core.
type CPU struct {
	// cfg is the configuration that was passed to NewCPU.
	cfg Config
	// iq is the instruction queue between fetch and issue.
	iq *instructionQueue
	// rs is the pool of reservation stations (ALU and LSU entries).
	rs *ReservationStation
	// rf is the architectural register file together with the rename tags.
	rf *RegisterStatus
	// alus is the pool of arithmetic/logic functional units.
	alus []*ALU
	// muls is the pool of MUL functional units (RV32M).
	muls []*ALU
	// divs is the pool of DIV/REM functional units (RV32M).
	divs []*ALU
	// lsus is the pool of load/store functional units.
	lsus []*LSU
	// cdb is the common data bus that broadcasts completed results.
	cdb *CommonDataBus
	// mem is the memory backend used by the load/store units.
	mem WordHandler
	// stats accumulates execution metrics across cycles.
	stats Statistics
	// lastBranch records the outcome of the most recently retired branch.
	lastBranch BranchInfo
	// instrPC is the program counter of the instruction currently in issue;
	// captured by Fetch so dispatched entries know their origin.
	instrPC uint32
	// hasUnresolvedBranch is true between issuing a branch and resolving
	// it in writeback; it gates the issue stage (stall-on-branch).
	hasUnresolvedBranch bool
}

func NewCPU(cfg Config) *CPU {
	alus := make([]*ALU, cfg.ALURSCount)
	for i := range alus {
		alus[i] = newALU(cfg.ALULatency)
	}
	muls := make([]*ALU, cfg.MulRSCount)
	for i := range muls {
		muls[i] = newALU(cfg.MulLatency)
	}
	divs := make([]*ALU, cfg.DivRSCount)
	for i := range divs {
		divs[i] = newALU(cfg.DivLatency)
	}
	lsus := make([]*LSU, cfg.LSURSCount)
	for i := range lsus {
		lsus[i] = newLSU(cfg.LoadLatency, cfg.StoreLatency)
	}
	rf := NewRegisterStatus()
	// The Spec-mandated 8-register layout uses Config.RegisterCount=8.
	// DefaultConfig keeps the historical 32-register layout. A
	// zero RegisterCount falls back to 32 so an uninitialised
	// Config does not silently zero the whole register file.
	limit := cfg.RegisterCount
	if limit == 0 {
		limit = 32
	}
	rf.SetLimit(uint32(limit))
	return &CPU{
		cfg:   cfg,
		iq:    newInstructionQueue(cfg.InstructionQueueSize),
		rs:    NewReservationStation(cfg.ALURSCount, cfg.LSURSCount),
		rf:    rf,
		alus:  alus,
		muls:  muls,
		divs:  divs,
		lsus:  lsus,
		cdb:   NewCommonDataBus(),
		stats: newStatistics(),
	}
}

// AttachMemory binds the load/store units to a memory backend.
func (c *CPU) AttachMemory(mem WordHandler) {
	c.mem = mem
	for _, l := range c.lsus {
		l.AttachMemory(mem)
	}
}

// Fetch enqueues an encoded instruction word for later issue. The instruction
// lives in the CPU's instruction queue until the issue stage has capacity and
// no unresolved branch is in flight. Returns true when the word was accepted
// into the queue, false when the queue is full (the caller should stall the
// PC and retry on the next cycle). pc is the program counter of the
// instruction, needed for branch and jump target computation.
func (c *CPU) Fetch(word uint32, pc uint32) bool {
	return c.iq.Enqueue(word, pc)
}

// RunCycle advances the CPU by one clock cycle: issue, execute, writeback.
func (c *CPU) RunCycle() {
	c.stats.Cycles++
	c.issueStage()
	c.execute()
	c.writeback()
}

func (c *CPU) Stats() Statistics {
	return c.stats
}

// Reg returns the architectural value of register idx.
func (c *CPU) Reg(idx uint32) uint32 {
	return c.rf.Read(idx)
}

// SetReg writes a value directly into the architectural register
// file. The rename tag is cleared so any subsequent read sees the
// new value, not a stale tag. R0 is silently ignored. Tests use
// this to pre-stage base addresses and increment constants.
func (c *CPU) SetReg(idx uint32, value uint32) {
	if idx == 0 {
		return
	}
	c.rf.Write(idx, value)
}

// LastBranch returns the outcome of the most recently retired branch or
// jump instruction. IsBranch is false for non-branch instructions.
func (c *CPU) LastBranch() BranchInfo {
	return c.lastBranch
}

// IQFull reports whether the instruction queue is currently full.
func (c *CPU) IQFull() bool {
	return c.iq.Len() >= c.iq.Cap()
}

// HasUnresolvedBranch reports whether there is a branch in flight that has
// not yet been resolved by the writeback stage.
func (c *CPU) HasUnresolvedBranch() bool {
	return c.hasUnresolvedBranch
}

// RSSnapshot returns a copy of every reservation-station entry across
// the integer (ALU/MUL/DIV) and load/store pools. The trace driver
// uses the snapshot for the 's' (state) control command, so it
// must remain consistent with the CPU's internal state without
// holding any locks. The two slices are returned in a single
// struct so a caller can render both pools side by side.
//
// Note: MUL and DIV share the integer reservation station (see
// arch/cpu/issue.go); the dedicated MUL and DIV functional-unit
// pools live in c.muls and c.divs but do not carry reservation
// state. Two slices are enough.
type RSSnapshot struct {
	ALU []RSEntry
	LSU []RSEntry
}

// SnapshotRS returns the current state of every reservation-station
// pool. Each entry is a value copy, so the caller cannot mutate the
// CPU by writing through the returned slice.
func (c *CPU) SnapshotRS() RSSnapshot {
	return RSSnapshot{
		ALU: append([]RSEntry(nil), c.rs.alu...),
		LSU: append([]RSEntry(nil), c.rs.lsu...),
	}
}

// RegisterSnapshot is a copy of the architectural register file and
// the rename tags. The driver renders R0..R7 from the V slice and
// the Qi slice explains the rename state for the 's' control
// command. The RegisterCount field limits the rendered window; a
// follow-up change (PR C) wires this to the Spec-mandated 8
// registers. For now the driver passes its own limit and the
// snapshot itself is not bounded.
type RegisterSnapshot struct {
	V  [32]uint32
	Qi [32]RSTag
}

// SnapshotRegisters returns a copy of the architectural register
// file and the rename tags.
func (c *CPU) SnapshotRegisters() RegisterSnapshot {
	return RegisterSnapshot{V: c.rf.V, Qi: c.rf.Qi}
}
