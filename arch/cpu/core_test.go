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

func TestSpecConfigDefaults(t *testing.T) {
	c := SpecConfig()
	if c.ALURSCount != 4 {
		t.Errorf("ALURSCount = %d, want 4", c.ALURSCount)
	}
	if c.LSURSCount != 3 {
		t.Errorf("LSURSCount = %d, want 3", c.LSURSCount)
	}
	if c.ALULatency != 1 {
		t.Errorf("ALULatency = %d, want 1", c.ALULatency)
	}
	if c.MulLatency != 3 {
		t.Errorf("MulLatency = %d, want 3", c.MulLatency)
	}
	if c.DivLatency != 5 {
		t.Errorf("DivLatency = %d, want 5", c.DivLatency)
	}
	if c.LoadLatency != 2 {
		t.Errorf("LoadLatency = %d, want 2", c.LoadLatency)
	}
	if c.StoreLatency != 2 {
		t.Errorf("StoreLatency = %d, want 2", c.StoreLatency)
	}
	if c.RegisterCount != 8 {
		t.Errorf("RegisterCount = %d, want 8", c.RegisterCount)
	}
}

// TestDefaultConfigIs32Register guards the historical default
// layout: 32 registers so the existing RV32I examples in
// examples/ continue to work without explicit configuration.
// A future change to the default must update this test.
func TestDefaultConfigIs32Register(t *testing.T) {
	c := DefaultConfig()
	if c.RegisterCount != 32 {
		t.Errorf("DefaultConfig RegisterCount = %d, want 32 (historical default)", c.RegisterCount)
	}
}

// TestSpecRegisterLimit checks that a CPU built with SpecConfig
// silently drops writes to registers R8..R31 and reads from them
// return zero, so the trace subset (R0..R7) sees a clean boundary.
func TestSpecRegisterLimit(t *testing.T) {
	core := NewCPU(SpecConfig())
	// A value destined for R8 is dropped.
	core.rf.Write(8, 0xDEADBEEF)
	if got := core.Reg(8); got != 0 {
		t.Errorf("Reg(8) = %#x, want 0 (out of range for SpecConfig)", got)
	}
	// R7 is in range; the write sticks.
	core.rf.Write(7, 42)
	if got := core.Reg(7); got != 42 {
		t.Errorf("Reg(7) = %d, want 42", got)
	}
	// R0 is hard-wired to zero even in the Spec layout.
	if got := core.Reg(0); got != 0 {
		t.Errorf("Reg(0) = %d, want 0", got)
	}
}

// TestDefaultRegisterLimitIs32 confirms that the historical
// 32-register default still allows writes to high indices.
func TestDefaultRegisterLimitIs32(t *testing.T) {
	core := NewCPU(DefaultConfig())
	core.rf.Write(31, 0xCAFE)
	if got := core.Reg(31); got != 0xCAFE {
		t.Errorf("Reg(31) = %#x, want 0xCAFE (default 32-register layout)", got)
	}
}

// TestZeroRegisterCountFallsBackTo32 guards a defensive check:
// an uninitialised Config (RegisterCount=0) must not silently
// zero the whole register file. The fix in NewCPU bumps the
// limit back to 32 in that case.
func TestZeroRegisterCountFallsBackTo32(t *testing.T) {
	core := NewCPU(Config{}) // RegisterCount defaults to zero
	core.rf.Write(31, 0xBEEF)
	if got := core.Reg(31); got != 0xBEEF {
		t.Errorf("Reg(31) = %#x, want 0xBEEF (zero RegisterCount should fall back to 32)", got)
	}
}
