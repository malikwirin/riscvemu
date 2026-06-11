package cpu

import "testing"

func TestNewCPUHasDefaultConfig(t *testing.T) {
	core := NewCPU(DefaultConfig())
	if core == nil {
		t.Fatal("NewCPU returned nil")
	}
	if core.rs == nil {
		t.Fatal("ReservationStation not initialized")
	}
	if core.rf == nil {
		t.Fatal("RegisterStatus not initialized")
	}
	if core.cdb == nil {
		t.Fatal("CommonDataBus not initialized")
	}
	if len(core.lsus) != 2 {
		t.Errorf("LSU pool size = %d, want 2", len(core.lsus))
	}
	if got := core.rs.ALURSCapacity(); got != 3 {
		t.Errorf("ALU RS capacity = %d, want 3", got)
	}
	if got := core.rs.LSURSCapacity(); got != 2 {
		t.Errorf("LSU RS capacity = %d, want 2", got)
	}
}

func TestRunCycleIncrementsCycleCount(t *testing.T) {
	core := NewCPU(DefaultConfig())
	core.RunCycle()
	core.RunCycle()
	core.RunCycle()
	s := core.Stats()
	if s.Cycles != 3 {
		t.Errorf("Cycles = %d, want 3", s.Cycles)
	}
}

func TestRegZeroStaysZero(t *testing.T) {
	core := NewCPU(DefaultConfig())
	if got := core.Reg(0); got != 0 {
		t.Errorf("x0 = %d, want 0", got)
	}
}
