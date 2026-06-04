package cpu

// WordHandler is the memory interface required by the load/store unit.
type WordHandler interface {
	ReadWord(addr uint32) (uint32, error)
	WriteWord(addr uint32, value uint32) error
}

// CPU is the Tomasulo-style execution core.
type CPU struct {
	cfg   Config
	rs    *ReservationStation
	rf    *RegisterStatus
	alus  []*ALU
	lsu   *LSU
	cdb   *CommonDataBus
	mem   WordHandler
	stats Statistics
}

func NewCPU(cfg Config) *CPU {
	return &CPU{
		cfg:   cfg,
		rs:    NewReservationStation(cfg.ALURSCount, cfg.LSURSCount),
		rf:    NewRegisterStatus(),
		alus:  make([]*ALU, cfg.ALURSCount),
		lsu:   NewLSU(),
		cdb:   NewCommonDataBus(),
		stats: newStatistics(),
	}
}

// AttachMemory binds the load/store unit to a memory backend.
func (c *CPU) AttachMemory(mem WordHandler) {
	c.mem = mem
	c.lsu.AttachMemory(mem)
}

// ReceiveInstruction feeds an encoded instruction word into the issue queue.
func (c *CPU) ReceiveInstruction(word uint32) error {
	return c.issue(word)
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
