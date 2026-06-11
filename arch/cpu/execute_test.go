package cpu

import "testing"

func TestExecuteStartsReadyALUEntry(t *testing.T) {
	core := NewCPU(DefaultConfig())
	if err := core.ReceiveInstruction(encode(t, "addi x1, x0, 5"), 0); err != nil {
		t.Fatalf("ReceiveInstruction: %v", err)
	}
	core.RunCycle()
	if busy := countBusyALUs(core); busy != 1 {
		t.Errorf("busy ALUs = %d, want 1", busy)
	}
}

func TestExecuteSkipsEntryWithPendingQj(t *testing.T) {
	core := NewCPU(DefaultConfig())
	// First issue: addi x1, x0, 7  -> sets Qi[x1]
	if err := core.ReceiveInstruction(encode(t, "addi x1, x0, 7"), 0); err != nil {
		t.Fatalf("issue 1: %v", err)
	}
	// Second issue: add x2, x1, x0  -> operand x1 is pending (Qj != NoTag)
	if err := core.ReceiveInstruction(encode(t, "add x2, x1, x0"), 0); err != nil {
		t.Fatalf("issue 2: %v", err)
	}
	core.RunCycle()
	// The first addi is ready, the second is not. Only one ALU should be busy.
	if busy := countBusyALUs(core); busy != 1 {
		t.Errorf("busy ALUs = %d, want 1 (only the first addi should start)", busy)
	}
}

func TestExecuteSkipsEntryWithPendingQk(t *testing.T) {
	core := NewCPU(DefaultConfig())
	if err := core.ReceiveInstruction(encode(t, "addi x1, x0, 7"), 0); err != nil {
		t.Fatalf("issue 1: %v", err)
	}
	// add x2, x0, x1 -> operand x1 (the second source) is pending (Qk != NoTag)
	if err := core.ReceiveInstruction(encode(t, "add x2, x0, x1"), 0); err != nil {
		t.Fatalf("issue 2: %v", err)
	}
	core.RunCycle()
	if busy := countBusyALUs(core); busy != 1 {
		t.Errorf("busy ALUs = %d, want 1 (second operand still pending)", busy)
	}
}

func TestExecuteFreezesOnNoFreeALU(t *testing.T) {
	cfg := DefaultConfig()
	cfg.ALURSCount = 1
	cfg.ALULatency = 3
	core := NewCPU(cfg)
	// First addi occupies the only ALU and will run for 3 cycles
	if err := core.ReceiveInstruction(encode(t, "addi x1, x0, 1"), 0); err != nil {
		t.Fatalf("issue 1: %v", err)
	}
	// Second addi is ready but cannot start: no free ALU
	if err := core.ReceiveInstruction(encode(t, "addi x2, x0, 2"), 0); err != nil {
		t.Fatalf("issue 2: %v", err)
	}
	core.RunCycle()
	if busy := countBusyALUs(core); busy != 1 {
		t.Errorf("busy ALUs = %d, want 1 (only one ALU exists)", busy)
	}
}

func TestExecuteOutOfOrderDispatch(t *testing.T) {
	cfg := DefaultConfig()
	cfg.ALURSCount = 2
	cfg.ALULatency = 1
	core := NewCPU(cfg)
	// Independent: addi x1 and addi x2 can both start at the same cycle.
	if err := core.ReceiveInstruction(encode(t, "addi x1, x0, 1"), 0); err != nil {
		t.Fatalf("issue 1: %v", err)
	}
	if err := core.ReceiveInstruction(encode(t, "addi x2, x0, 2"), 0); err != nil {
		t.Fatalf("issue 2: %v", err)
	}
	core.RunCycle()
	if busy := countBusyALUs(core); busy != 2 {
		t.Errorf("busy ALUs = %d, want 2 (both ready entries should start)", busy)
	}
}

func TestExecuteLatencyCountdownCompletes(t *testing.T) {
	cfg := DefaultConfig()
	cfg.ALULatency = 2
	core := NewCPU(cfg)
	if err := core.ReceiveInstruction(encode(t, "addi x1, x0, 5"), 0); err != nil {
		t.Fatalf("ReceiveInstruction: %v", err)
	}
	core.RunCycle() // dispatch (remain=2, busy)
	if busy := countBusyALUs(core); busy != 1 {
		t.Fatalf("busy ALUs after cycle 1 = %d, want 1", busy)
	}
	core.RunCycle() // step (remain=2->1, still busy)
	if busy := countBusyALUs(core); busy != 1 {
		t.Errorf("busy ALUs after cycle 2 = %d, want 1 (latency=2 still running)", busy)
	}
	core.RunCycle() // step (remain=1->0, done)
	if busy := countBusyALUs(core); busy != 0 {
		t.Errorf("busy ALUs after cycle 3 = %d, want 0 (latency=2 complete)", busy)
	}
}

func TestExecuteSecondEntryStartsAfterFirstCompletes(t *testing.T) {
	cfg := DefaultConfig()
	cfg.ALURSCount = 2
	cfg.ALULatency = 1
	core := NewCPU(cfg)
	if err := core.ReceiveInstruction(encode(t, "addi x1, x0, 1"), 0); err != nil {
		t.Fatalf("issue 1: %v", err)
	}
	if err := core.ReceiveInstruction(encode(t, "addi x2, x0, 2"), 0); err != nil {
		t.Fatalf("issue 2: %v", err)
	}
	core.RunCycle() // ALU[0] starts with x1; ALU[1] starts with x2 (both latency 1)
	if busy := countBusyALUs(core); busy != 2 {
		t.Fatalf("busy ALUs = %d, want 2", busy)
	}
	// After 1 more cycle, both should be done; then issue a third and verify it dispatches.
	if err := core.ReceiveInstruction(encode(t, "addi x3, x0, 3"), 0); err != nil {
		t.Fatalf("issue 3: %v", err)
	}
	core.RunCycle() // both finish; new entry x3 dispatches into the freed ALU
	if busy := countBusyALUs(core); busy != 1 {
		t.Errorf("busy ALUs = %d, want 1 (only x3 should be running)", busy)
	}
}

func TestExecuteNoNewDispatchAfterAllDone(t *testing.T) {
	cfg := DefaultConfig()
	cfg.ALULatency = 1
	core := NewCPU(cfg)
	if err := core.ReceiveInstruction(encode(t, "addi x1, x0, 1"), 0); err != nil {
		t.Fatal(err)
	}
	core.RunCycle()
	core.RunCycle()
	core.RunCycle()
	if busy := countBusyALUs(core); busy != 0 {
		t.Errorf("busy ALUs = %d, want 0", busy)
	}
}

// countBusyALUs returns the number of ALUs currently in the busy state.
func countBusyALUs(c *CPU) int {
	n := 0
	for _, a := range c.alus {
		if a.IsBusy() {
			n++
		}
	}
	return n
}
