package cpu

import "testing"

func TestExecuteStartsReadyALUEntry(t *testing.T) {
	core := makeCPU(t, withALULatency(2))
	fetchProgram(t, core, "addi x1, x0, 5")
	core.RunCycle() // dispatch + step: remain=2->1
	if busy := countBusyALUs(core); busy != 1 {
		t.Errorf("busy ALUs = %d, want 1", busy)
	}
}

func TestExecuteSkipsEntriesWithPendingOperand(t *testing.T) {
	// Both pending-Qj and pending-Qk cases are structurally identical
	// (a producer, then a consumer whose first or second operand
	// reads the producer's register). The table just exercises both
	// positions.
	cases := []struct {
		name    string
		program []string
	}{
		{
			name:    "pending_Qj",
			program: []string{"addi x1, x0, 7", "add x2, x1, x0"},
		},
		{
			name:    "pending_Qk",
			program: []string{"addi x1, x0, 7", "add x2, x0, x1"},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			core := makeCPU(t, withALULatency(2))
			fetchProgram(t, core, tc.program...)
			core.RunCycle() // producer dispatches; consumer waits
			if busy := countBusyALUs(core); busy != 1 {
				t.Errorf("busy ALUs = %d, want 1 (only the producer should start)", busy)
			}
		})
	}
}

func TestExecuteFreezesOnNoFreeALU(t *testing.T) {
	core := makeCPU(t, withALURSCount(1), withALULatency(3))
	fetchProgram(t, core, "addi x1, x0, 1", "addi x2, x0, 2")
	core.RunCycle() // issue x1, dispatch x1
	if busy := countBusyALUs(core); busy != 1 {
		t.Errorf("busy ALUs = %d, want 1 (only one ALU exists)", busy)
	}
}

func TestExecuteOutOfOrderDispatch(t *testing.T) {
	core := makeCPU(t, withALURSCount(2), withALULatency(3))
	fetchProgram(t, core, "addi x1, x0, 1", "addi x2, x0, 2")
	core.RunCycle() // issue x1, dispatch x1
	core.RunCycle() // issue x2, dispatch x2
	if busy := countBusyALUs(core); busy != 2 {
		t.Errorf("busy ALUs = %d, want 2 (both entries should be running)", busy)
	}
}

func TestExecuteLatencyCountdownCompletes(t *testing.T) {
	core := makeCPU(t, withALULatency(2))
	fetchProgram(t, core, "addi x1, x0, 5")
	core.RunCycle() // dispatch + step: remain=2->1
	if busy := countBusyALUs(core); busy != 1 {
		t.Fatalf("busy ALUs after cycle 1 = %d, want 1", busy)
	}
	core.RunCycle() // step: remain=1->0, done, writeback fires
	if busy := countBusyALUs(core); busy != 0 {
		t.Errorf("busy ALUs after cycle 2 = %d, want 0 (latency=2 complete)", busy)
	}
}

func TestExecuteNoNewDispatchAfterAllDone(t *testing.T) {
	core := makeCPU(t, withALULatency(2))
	fetchProgram(t, core, "addi x1, x0, 1")
	for i := 0; i < 4; i++ {
		core.RunCycle()
	}
	if busy := countBusyALUs(core); busy != 0 {
		t.Errorf("busy ALUs = %d, want 0", busy)
	}
}
