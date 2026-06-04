package cpu

type Config struct {
	ALURSCount   int
	LSURSCount   int
	ALULatency   int
	LoadLatency  int
	StoreLatency int
}

func DefaultConfig() Config {
	return Config{
		ALURSCount:   3,
		LSURSCount:   2,
		ALULatency:   1,
		LoadLatency:  2,
		StoreLatency: 2,
	}
}
