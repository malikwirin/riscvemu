package cpu

import "testing"

func TestWriteBackUpdatesRegister(t *testing.T) {
	core := makeCPU(t)
	fetchProgram(t, core, "addi x1, x0, 5")
	core.RunCycle() // dispatch x1 to ALU[0]
	core.RunCycle() // step: ALU[0] finishes; writeback broadcasts
	if got := core.Reg(1); got != 5 {
		t.Errorf("x1 = %d, want 5", got)
	}
}

func TestWriteBackClearsQi(t *testing.T) {
	core := makeCPU(t)
	fetchProgram(t, core, "addi x1, x0, 5")
	core.issueStage()
	if core.rf.Qi[1] == NoTag {
		t.Fatal("Qi[x1] should be set after issue")
	}
	core.RunCycle()
	if core.rf.Qi[1] != NoTag {
		t.Errorf("Qi[x1] = %d, want NoTag after writeback", core.rf.Qi[1])
	}
}

func TestWriteBackIncrementsRetired(t *testing.T) {
	core := makeCPU(t)
	fetchProgram(t, core, "addi x1, x0, 5")
	core.RunCycle()
	core.RunCycle()
	if got := core.Stats().Retired; got != 1 {
		t.Errorf("Retired = %d, want 1", got)
	}
}

// TestWriteBackCDBPath covers both wake-up (filling the consumer's
// operand) and the RAW counter that tracks how many such resolutions
// happened. The two flows are independent: the wake is observed on
// the consumer's RS entry, the RAW counter is read from stats.
func TestWriteBackCDBPathWake(t *testing.T) {
	core := makeCPU(t)
	fetchProgram(t, core, "addi x1, x0, 5", "add x2, x1, x0")
	core.issueStage() // addi x1
	core.issueStage() // add x2 with Qj = tag(addi)
	// Simulate addi's writeback by clearing Qi[1] and putting the
	// value into the architectural file, then run a cycle so the
	// dispatch loop sees the consumer's cleared Qj and dispatches it.
	core.rf.V[1] = 5
	core.rf.Qi[1] = NoTag
	core.RunCycle() // dispatch addi, complete, writeback; wake add x2
	entry := findRSEntry(t, core, 2)
	if entry.Qj != NoTag {
		t.Errorf("Qj = %d, want NoTag", entry.Qj)
	}
	if entry.Vj != 5 {
		t.Errorf("Vj = %d, want 5", entry.Vj)
	}
}

func TestWriteBackIncrementsRAWResolved(t *testing.T) {
	// Same setup as the wake test, but with a latency that lets the
	// addi actually retire through writeback so the CDB broadcast
	// drives the wake and the RAW counter.
	core := makeCPU(t, withALULatency(3))
	fetchProgram(t, core, "addi x1, x0, 5", "add x2, x1, x0")
	for i := 0; i < 8; i++ {
		core.RunCycle()
	}
	if got := core.Stats().RAWResolved; got < 1 {
		t.Errorf("RAWResolved = %d, want >= 1", got)
	}
}

func TestWriteBackArbitratesOnePerCycle(t *testing.T) {
	core := makeCPU(t, withALURSCount(2), withALULatency(2))
	fetchProgram(t, core, "addi x1, x0, 1", "addi x2, x0, 2")
	core.RunCycle() // issue x1, dispatch ALU[0]
	core.RunCycle() // issue x2, dispatch ALU[1]
	if got := countBusyALUs(core); got != 1 {
		t.Fatalf("busy ALUs = %d, want 1 (one just completed, one still running)", got)
	}
	for i := 0; i < 5; i++ {
		core.RunCycle()
	}
	if got := core.Stats().Retired; got != 2 {
		t.Errorf("Retired = %d, want 2", got)
	}
}

func TestWriteBackWithPendingOperandDoesNotStall(t *testing.T) {
	// Producer and consumer both reticulate in order: dispatch x1,
	// x1 retires, dispatch x2, x2 retires. After 4 cycles the chain
	// has fully committed.
	core := makeCPU(t, withALURSCount(2))
	fetchProgram(t, core, "addi x1, x0, 5", "add x2, x1, x0")
	core.RunCycle() // dispatch x1
	core.RunCycle() // x1 done; writeback wakes x2
	core.RunCycle() // dispatch x2
	core.RunCycle() // x2 done; writeback commits x2
	if got := core.Reg(2); got != 5 {
		t.Errorf("x2 = %d, want 5", got)
	}
}

func TestWriteBackEndToEndMultipleArithmetic(t *testing.T) {
	// Three addis, two of which feed the third. With latency 1
	// and a 2-ALU pool, the chain retires in roughly 4 cycles.
	core := makeCPU(t)
	fetchProgram(t, core,
		"addi x1, x0, 10",
		"addi x2, x0, 20",
		"add x3, x1, x2",
	)
	for i := 0; i < 5; i++ {
		core.RunCycle()
	}
	if got := core.Reg(3); got != 30 {
		t.Errorf("x3 = %d, want 30", got)
	}
	if got := core.Stats().Retired; got < 3 {
		t.Errorf("Retired = %d, want >= 3", got)
	}
}
