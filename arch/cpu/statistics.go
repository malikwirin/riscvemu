package cpu

// Statistics tracks execution metrics for the Tomasulo CPU.
type Statistics struct {
	// Cycles is the total number of clock cycles elapsed.
	Cycles uint64

	// Retired is the number of instructions that completed writeback.
	Retired uint64

	// Issued is the number of instructions dispatched to reservation stations.
	Issued uint64

	// StructuralStalls counts cycles where issue was blocked by full reservation stations.
	StructuralStalls uint64

	// RAWResolved counts dependencies that were resolved via the common data bus.
	RAWResolved uint64

	// FunctionalBusyCycles counts how many cycles each functional unit was busy.
	FunctionalBusyCycles map[OpKind]uint64
}

func newStatistics() Statistics {
	return Statistics{
		FunctionalBusyCycles: make(map[OpKind]uint64),
	}
}

// IPC returns the instructions-per-cycle ratio (Retired / Cycles).
// Returns 0 when no cycles have been executed.
func (s Statistics) IPC() float64 {
	if s.Cycles == 0 {
		return 0
	}
	return float64(s.Retired) / float64(s.Cycles)
}

// FUUtil returns the utilisation of a given functional unit (busy cycles / total cycles).
// Returns 0 when no cycles have been executed.
func (s Statistics) FUUtil(kind OpKind) float64 {
	if s.Cycles == 0 {
		return 0
	}
	return float64(s.FunctionalBusyCycles[kind]) / float64(s.Cycles)
}
