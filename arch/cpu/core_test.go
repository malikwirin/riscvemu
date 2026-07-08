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

// TestSpecConfigDefaults pins every field of the Spec-mandated
// configuration: 4 ALU RS, 3 LSU RS, 8 registers, and the
// five canonical latencies.
func TestSpecConfigDefaults(t *testing.T) {
	c := SpecConfig()
	cases := []struct {
		field string
		got   int
		want  int
	}{
		{"ALURSCount", c.ALURSCount, 4},
		{"LSURSCount", c.LSURSCount, 3},
		{"ALULatency", c.ALULatency, 1},
		{"MulLatency", c.MulLatency, 3},
		{"DivLatency", c.DivLatency, 5},
		{"LoadLatency", c.LoadLatency, 2},
		{"StoreLatency", c.StoreLatency, 2},
		{"RegisterCount", c.RegisterCount, 8},
	}
	for _, tc := range cases {
		if tc.got != tc.want {
			t.Errorf("%s = %d, want %d", tc.field, tc.got, tc.want)
		}
	}
}

// TestRegisterLimit pins the register-file size limit
// behaviour across the three interesting configurations: the
// 8-register Spec layout, the historical 32-register default,
// and an uninitialised Config (RegisterCount=0) that must
// fall back to 32.
func TestRegisterLimit(t *testing.T) {
	cases := []struct {
		name     string
		cfg      Config
		writeReg uint32
		wantRead uint32
	}{
		{"Spec layout drops R8+", SpecConfig(), 8, 0},
		{"Spec layout keeps R7", SpecConfig(), 7, 42},
		{"default 32-register keeps R31", DefaultConfig(), 31, 0xCAFE},
		{"zero RegisterCount falls back to 32", Config{}, 31, 0xBEEF},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			core := NewCPU(tc.cfg)
			if tc.wantRead != 0 {
				core.rf.Write(tc.writeReg, tc.wantRead)
			} else {
				core.rf.Write(tc.writeReg, 0xDEADBEEF)
			}
			if got := core.Reg(tc.writeReg); got != tc.wantRead {
				t.Errorf("Reg(%d) = %#x, want %#x", tc.writeReg, got, tc.wantRead)
			}
		})
	}
}
