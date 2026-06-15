package cpu

type Config struct {
	ALURSCount   int
	LSURSCount   int
	ALULatency   int
	LoadLatency  int
	StoreLatency int
	// RV32M: MUL and DIV run in their own functional units so that
	// their higher latencies do not stall the integer ALU pool. They
	// share the ALU reservation station (one integer rename space),
	// but dispatch to MULFU or DIVFU based on the operation kind.
	MulRSCount int
	DivRSCount int
	MulLatency int
	DivLatency int
	// InstructionQueueSize is the number of entries in the fetch-to-issue
	// instruction queue.
	InstructionQueueSize int
}

func DefaultConfig() Config {
	return Config{
		ALURSCount:           3,
		LSURSCount:           2,
		ALULatency:           1,
		LoadLatency:          2,
		StoreLatency:         2,
		MulRSCount:           1,
		DivRSCount:           1,
		MulLatency:           3,
		DivLatency:           8,
		InstructionQueueSize: 8,
	}
}
