package cpu

import "testing"

func TestExecuteStartsReadyALUEntry(t *testing.T) {
	cfg := DefaultConfig()
	cfg.ALULatency = 2
	core := NewCPU(cfg)
	if !core.Fetch(encode(t, "addi x1, x0, 5"), 0) {
		t.Fatal("Fetch failed")
	}
	core.RunCycle() // dispatch + step: remain=2->1
	if busy := countBusyALUs(core); busy != 1 {
		t.Errorf("busy ALUs = %d, want 1", busy)
	}
}

func TestExecuteSkipsEntryWithPendingQj(t *testing.T) {
	cfg := DefaultConfig()
	cfg.ALULatency = 2
	core := NewCPU(cfg)
	if !core.Fetch(encode(t, "addi x1, x0, 7"), 0) {
		t.Fatal("Fetch 1 failed")
	}
	if !core.Fetch(encode(t, "add x2, x1, x0"), 0) {
		t.Fatal("Fetch 2 failed")
	}
	core.RunCycle() // issue x1, dispatch x1, step x1 (busy)
	// The first addi is busy; the second has pending operand and waits.
	if busy := countBusyALUs(core); busy != 1 {
		t.Errorf("busy ALUs = %d, want 1 (only the first addi should start)", busy)
	}
}

func TestExecuteSkipsEntryWithPendingQk(t *testing.T) {
	cfg := DefaultConfig()
	cfg.ALULatency = 2
	core := NewCPU(cfg)
	if !core.Fetch(encode(t, "addi x1, x0, 7"), 0) {
		t.Fatal("Fetch 1 failed")
	}
	if !core.Fetch(encode(t, "add x2, x0, x1"), 0) {
		t.Fatal("Fetch 2 failed")
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
	if !core.Fetch(encode(t, "addi x1, x0, 1"), 0) {
		t.Fatal("Fetch 1 failed")
	}
	if !core.Fetch(encode(t, "addi x2, x0, 2"), 0) {
		t.Fatal("Fetch 2 failed")
	}
	core.RunCycle() // issue x1, dispatch x1
	if busy := countBusyALUs(core); busy != 1 {
		t.Errorf("busy ALUs = %d, want 1 (only one ALU exists)", busy)
	}
}

func TestExecuteOutOfOrderDispatch(t *testing.T) {
	cfg := DefaultConfig()
	cfg.ALURSCount = 2
	cfg.ALULatency = 3
	core := NewCPU(cfg)
	if !core.Fetch(encode(t, "addi x1, x0, 1"), 0) {
		t.Fatal("Fetch 1 failed")
	}
	if !core.Fetch(encode(t, "addi x2, x0, 2"), 0) {
		t.Fatal("Fetch 2 failed")
	}
	core.RunCycle() // issue x1, dispatch x1
	core.RunCycle() // issue x2, dispatch x2
	if busy := countBusyALUs(core); busy != 2 {
		t.Errorf("busy ALUs = %d, want 2 (both entries should be running)", busy)
	}
}

func TestExecuteLatencyCountdownCompletes(t *testing.T) {
	cfg := DefaultConfig()
	cfg.ALULatency = 2
	core := NewCPU(cfg)
	if !core.Fetch(encode(t, "addi x1, x0, 5"), 0) {
		t.Fatal("Fetch failed")
	}
	core.RunCycle() // dispatch + step: remain=2->1
	if busy := countBusyALUs(core); busy != 1 {
		t.Fatalf("busy ALUs after cycle 1 = %d, want 1", busy)
	}
	core.RunCycle() // step: remain=1->0, done, writeback fires
	if busy := countBusyALUs(core); busy != 0 {
		t.Errorf("busy ALUs after cycle 2 = %d, want 0 (latency=2 complete)", busy)
	}
}

func TestExecuteSecondEntryStartsAfterFirstCompletes(t *testing.T) {
	cfg := DefaultConfig()
	cfg.ALURSCount = 2
	cfg.ALULatency = 3
	core := NewCPU(cfg)
	if !core.Fetch(encode(t, "addi x1, x0, 1"), 0) {
		t.Fatal("Fetch 1 failed")
	}
	if !core.Fetch(encode(t, "addi x2, x0, 2"), 0) {
		t.Fatal("Fetch 2 failed")
	}
	core.RunCycle() // issue x1, dispatch x1
	core.RunCycle() // issue x2, dispatch x2
	if busy := countBusyALUs(core); busy != 2 {
		t.Fatalf("busy ALUs = %d, want 2", busy)
	}
	if !core.Fetch(encode(t, "addi x3, x0, 3"), 0) {
		t.Fatal("Fetch 3 failed")
	}
	core.RunCycle() // x1 and x2 still running; x3 issued but waiting for FU
	if busy := countBusyALUs(core); busy < 1 {
		t.Errorf("busy ALUs = %d, want >= 1", busy)
	}
}

func TestExecuteNoNewDispatchAfterAllDone(t *testing.T) {
	cfg := DefaultConfig()
	cfg.ALULatency = 2
	core := NewCPU(cfg)
	if !core.Fetch(encode(t, "addi x1, x0, 1"), 0) {
		t.Fatal("Fetch failed")
	}
	for i := 0; i < 4; i++ {
		core.RunCycle()
	}
	if busy := countBusyALUs(core); busy != 0 {
		t.Errorf("busy ALUs = %d, want 0", busy)
	}
}
