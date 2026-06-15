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
	// RegisterCount is the number of architectural registers
	// exposed by this configuration. Reads and writes to a
	// register index >= RegisterCount return zero (reads) or
	// are dropped (writes), so a trace driver or test that
	// only uses the Spec-mandated R0..R7 sees a clean
	// boundary even though the underlying storage remains
	// 32-wide.
	RegisterCount int
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
		RegisterCount:        32,
	}
}

func SpecConfig() Config {
	return Config{
		ALURSCount:           4,
		LSURSCount:           3,
		ALULatency:           1,
		LoadLatency:          2,
		StoreLatency:         2,
		MulRSCount:           1,
		DivRSCount:           1,
		MulLatency:           3,
		DivLatency:           5,
		InstructionQueueSize: 8,
		RegisterCount:        8,
	}
}
