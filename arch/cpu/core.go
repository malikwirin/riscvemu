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
	// rs is the pool of reservation stations (ALU and LSU entries).
	rs *ReservationStation
	// rf is the architectural register file together with the rename tags.
	rf *RegisterStatus
	// alus is the pool of arithmetic/logic functional units.
	alus []*ALU
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
	// captured by ReceiveInstruction so dispatched entries know their origin.
	instrPC uint32
}

func NewCPU(cfg Config) *CPU {
	alus := make([]*ALU, cfg.ALURSCount)
	for i := range alus {
		alus[i] = newALU(cfg.ALULatency)
	}
	lsus := make([]*LSU, cfg.LSURSCount)
	for i := range lsus {
		lsus[i] = newLSU(cfg.LoadLatency, cfg.StoreLatency)
	}
	return &CPU{
		cfg:   cfg,
		rs:    NewReservationStation(cfg.ALURSCount, cfg.LSURSCount),
		rf:    NewRegisterStatus(),
		alus:  alus,
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

// ReceiveInstruction feeds an encoded instruction word into the issue queue.
// pc is the program counter of the instruction, needed for branch and jump
// target computation. Pass 0 if the caller has no PC information.
func (c *CPU) ReceiveInstruction(word uint32, pc uint32) error {
	c.instrPC = pc
	return c.issue(word, pc)
}

// RunCycle advances the CPU by one clock cycle: issue, execute, writeback.
func (c *CPU) RunCycle() {
	c.stats.Cycles++
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
