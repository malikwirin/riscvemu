package cpu

import (
	"testing"

	"codeberg.org/malik/riscvemu/arch/cpu/cputest"
)

func TestStatisticsIPC(t *testing.T) {
	core := NewCPU(DefaultConfig())
	// Issue 2 addi, let them complete
	if !core.Fetch(cputest.Encode(t, "addi x1, x0, 1"), 0) {
		t.Fatal("Fetch failed")
	}
	if !core.Fetch(cputest.Encode(t, "addi x2, x0, 2"), 0) {
		t.Fatal("Fetch failed")
	}
	core.RunCycle() // dispatch both
	core.RunCycle() // first broadcast
	core.RunCycle() // second broadcast
	s := core.Stats()
	if got := s.IPC(); got < 0.5 || got > 1.0 {
		t.Errorf("IPC = %f, want ~0.67 (2 retired / 3 cycles)", got)
	}
}

func TestStatisticsFUUtil(t *testing.T) {
	core := NewCPU(DefaultConfig())
	if !core.Fetch(cputest.Encode(t, "addi x1, x0, 1"), 0) {
		t.Fatal("Fetch failed")
	}
	core.RunCycle() // dispatch
	core.RunCycle() // step, writeback
	s := core.Stats()
	// ADDI has latency 1, so it occupies the FU for exactly 1 cycle;
	// 1 busy cycle / 2 total cycles = 0.5.
	if got := s.FUUtil(OpADDI); got != 0.5 {
		t.Errorf("FUUtil(OpADDI) = %f, want 0.5", got)
	}
}

// TestStatisticsFUBusyCyclesMatchLatency is a regression test for the
// FU-utilisation counter. A functional unit must be counted as busy
// for every cycle it actually holds an in-flight instruction, not just
// once at dispatch time. With MUL latency 3 and two independent MULs
// the MUL FU is busy for 6 cycles in total.
func TestStatisticsFUBusyCyclesMatchLatency(t *testing.T) {
	core := makeCPU(t, withMulLatency(3), withMulRSCount(1))
	fetchProgram(t, core,
		"addi x1, x0, 2",
		"addi x2, x0, 3",
		"mul x3, x1, x2", // MUL #1
		"addi x4, x0, 4",
		"addi x5, x0, 5",
		"mul x6, x4, x5", // MUL #2
	)
	runUntilSettled(t, core, 30, func() bool { return core.Stats().Retired >= 6 })
	if got := core.Stats().FunctionalBusyCycles[OpMUL]; got != 6 {
		t.Errorf("FunctionalBusyCycles[OpMUL] = %d, want 6 (2 MULs * latency 3)", got)
	}
}

// TestStatisticsFUUtilPerUnit ensures the busy-cycle counter is
// tracked per OpKind so that activity in one FU pool does not
// contaminate the utilisation reported for another.
func TestStatisticsFUUtilPerUnit(t *testing.T) {
	core := makeCPU(t, withMulLatency(3), withLoadLatency(2))
	mem := &cputest.MockWordHandler{Mem: map[uint32]uint32{100: 0xDEADBEEF}}
	core.AttachMemory(mem)
	fetchProgram(t, core,
		"addi x2, x0, 100",
		"lw x1, 0(x2)", // LOAD: 2 busy cycles
		"addi x3, x1, 1",
		"mul x4, x3, x1", // MUL: 3 busy cycles
	)
	runUntilSettled(t, core, 30, func() bool { return core.Stats().Retired >= 4 })
	if got := core.Stats().FunctionalBusyCycles[OpLOAD]; got != 2 {
		t.Errorf("FunctionalBusyCycles[OpLOAD] = %d, want 2 (1 LOAD * latency 2)", got)
	}
	if got := core.Stats().FunctionalBusyCycles[OpMUL]; got != 3 {
		t.Errorf("FunctionalBusyCycles[OpMUL] = %d, want 3 (1 MUL * latency 3)", got)
	}
}
