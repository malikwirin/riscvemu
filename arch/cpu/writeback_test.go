package cpu

import "testing"

func TestWriteBackUpdatesRegister(t *testing.T) {
	core := NewCPU(DefaultConfig())
	if !core.Fetch(encode(t, "addi x1, x0, 5"), 0) {
		t.Fatal("Fetch failed")
	}
	core.RunCycle() // dispatch x1 to ALU[0]
	core.RunCycle() // step: ALU[0] finishes; writeback broadcasts
	if got := core.Reg(1); got != 5 {
		t.Errorf("x1 = %d, want 5", got)
	}
}

func TestWriteBackWakesRSEntryOnCDB(t *testing.T) {
	core := NewCPU(DefaultConfig())
	if !core.Fetch(encode(t, "addi x1, x0, 5"), 0) {
		t.Fatal("Fetch failed")
	}
	if !core.Fetch(encode(t, "add x2, x1, x0"), 0) {
		t.Fatal("Fetch failed")
	}
	core.issueStage() // issue addi x1 (sets Qi[1])
	core.issueStage() // issue add x2 (Qj = Qi[1])
	// Simulate addi's writeback by clearing Qi[1] and setting V[1] directly.
	core.rf.V[1] = 5
	core.rf.Qi[1] = NoTag
	core.RunCycle() // dispatch addi, complete, writeback; wakeup add x2
	var found *RSEntry
	for i := range core.rs.alu {
		if core.rs.alu[i].Busy && core.rs.alu[i].Rd == 2 {
			found = &core.rs.alu[i]
			break
		}
	}
	if found == nil {
		t.Fatal("RS entry for x2 not found")
	}
	if found.Qj != NoTag {
		t.Errorf("RS entry Qj = %d, want NoTag", found.Qj)
	}
	if found.Vj != 5 {
		t.Errorf("RS entry Vj = %d, want 5", found.Vj)
	}
}

func TestWriteBackIncrementsRetired(t *testing.T) {
	core := NewCPU(DefaultConfig())
	if !core.Fetch(encode(t, "addi x1, x0, 5"), 0) {
		t.Fatal("Fetch failed")
	}
	core.RunCycle()
	core.RunCycle()
	if got := core.Stats().Retired; got != 1 {
		t.Errorf("Retired = %d, want 1", got)
	}
}

func TestWriteBackIncrementsRAWResolved(t *testing.T) {
	cfg := DefaultConfig()
	cfg.ALULatency = 3
	core := NewCPU(cfg)
	if !core.Fetch(encode(t, "addi x1, x0, 5"), 0) {
		t.Fatal("Fetch failed")
	}
	if !core.Fetch(encode(t, "add x2, x1, x0"), 0) {
		t.Fatal("Fetch failed")
	}
	core.issueStage() // issue addi x1 (sets Qi[1])
	core.issueStage() // issue add x2 (Qj = Qi[1])
	// Simulate addi's writeback by clearing Qi[1] and setting V[1] directly.
	core.rf.V[1] = 5
	core.rf.Qi[1] = NoTag
	for i := 0; i < 5; i++ {
		core.RunCycle()
	}
	if got := core.Stats().RAWResolved; got < 1 {
		t.Errorf("RAWResolved = %d, want >= 1", got)
	}
}

func TestWriteBackArbitratesOnePerCycle(t *testing.T) {
	cfg := DefaultConfig()
	cfg.ALURSCount = 2
	cfg.ALULatency = 2
	core := NewCPU(cfg)
	if !core.Fetch(encode(t, "addi x1, x0, 1"), 0) {
		t.Fatal("Fetch failed")
	}
	if !core.Fetch(encode(t, "addi x2, x0, 2"), 0) {
		t.Fatal("Fetch failed")
	}
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

func TestWriteBackClearsQi(t *testing.T) {
	core := NewCPU(DefaultConfig())
	if !core.Fetch(encode(t, "addi x1, x0, 5"), 0) {
		t.Fatal("Fetch failed")
	}
	core.issueStage()
	if core.rf.Qi[1] == NoTag {
		t.Fatal("Qi[x1] should be set after issue")
	}
	core.RunCycle()
	if core.rf.Qi[1] != NoTag {
		t.Errorf("Qi[x1] = %d, want NoTag after writeback", core.rf.Qi[1])
	}
}

func TestWriteBackWithPendingOperandDoesNotStall(t *testing.T) {
	cfg := DefaultConfig()
	cfg.ALURSCount = 2
	core := NewCPU(cfg)
	// x1 <- 5
	if !core.Fetch(encode(t, "addi x1, x0, 5"), 0) {
		t.Fatal("Fetch failed")
	}
	// x2 <- x1 + 0
	if !core.Fetch(encode(t, "add x2, x1, x0"), 0) {
		t.Fatal("Fetch failed")
	}
	core.RunCycle() // dispatch x1
	core.RunCycle() // x1 done; writeback wakes x2
	core.RunCycle() // dispatch x2
	core.RunCycle() // x2 done; writeback commits x2
	if got := core.Reg(2); got != 5 {
		t.Errorf("x2 = %d, want 5", got)
	}
}

func TestWriteBackEndToEndMultipleArithmetic(t *testing.T) {
	core := NewCPU(DefaultConfig())
	// x1 <- 10
	if !core.Fetch(encode(t, "addi x1, x0, 10"), 0) {
		t.Fatal("Fetch failed")
	}
	// x2 <- 20
	if !core.Fetch(encode(t, "addi x2, x0, 20"), 0) {
		t.Fatal("Fetch failed")
	}
	// x3 <- x1 + x2
	if !core.Fetch(encode(t, "add x3, x1, x2"), 0) {
		t.Fatal("Fetch failed")
	}
	core.RunCycle() // dispatch x1, x2; x3 has pending operands
	core.RunCycle() // one of x1/x2 broadcasts, the other waits
	core.RunCycle() // the other one broadcasts; x3 now has both operands
	core.RunCycle() // dispatch x3
	core.RunCycle() // x3 completes
	if got := core.Reg(3); got != 30 {
		t.Errorf("x3 = %d, want 30", got)
	}
	if got := core.Stats().Retired; got < 3 {
		t.Errorf("Retired = %d, want >= 3", got)
	}
}
