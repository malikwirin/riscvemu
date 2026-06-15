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
	return &CPU{
		cfg:   cfg,
		iq:    newInstructionQueue(cfg.InstructionQueueSize),
		rs:    NewReservationStation(cfg.ALURSCount, cfg.LSURSCount),
		rf:    NewRegisterStatus(),
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
