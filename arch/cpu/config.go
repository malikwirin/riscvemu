package cpu

type Config struct {
	ALURSCount   int
	LSURSCount   int
	ALULatency   int
	LoadLatency  int
	StoreLatency int
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
		InstructionQueueSize: 8,
	}
}
