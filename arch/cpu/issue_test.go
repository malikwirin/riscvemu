package cpu

import "testing"

func TestIssueADDIPlacesInALURS(t *testing.T) {
	cfg := DefaultConfig()
	cfg.ALULatency = 2
	core := NewCPU(cfg)
	word := encode(t, "addi x1, x0, 5")
	if !core.Fetch(word, 0) {
		t.Fatal("Fetch failed")
	}
	if got := core.rs.ALUCountBusy(); got != 0 {
		t.Fatalf("ALU RS busy before issue = %d, want 0", got)
	}
	core.RunCycle() // dispatch + step (remain=2->1)
	if got := countBusyALUs(core); got != 1 {
		t.Errorf("ALU busy = %d, want 1", got)
	}
	if got := core.rs.LSUCountBusy(); got != 0 {
		t.Errorf("LSU RS busy = %d, want 0", got)
	}
}

func TestIssueLWPPlacesInLSURS(t *testing.T) {
	cfg := DefaultConfig()
	cfg.LoadLatency = 3
	core := NewCPU(cfg)
	word := encode(t, "lw x3, 0(x2)")
	if !core.Fetch(word, 0) {
		t.Fatal("Fetch failed")
	}
	if got := core.rs.LSUCountBusy(); got != 0 {
		t.Fatalf("LSU RS busy before issue = %d, want 0", got)
	}
	core.RunCycle() // dispatch + step (remain=3->2)
	if got := countBusyLSUs(core); got != 1 {
		t.Errorf("LSU busy = %d, want 1", got)
	}
	if got := core.rs.ALUCountBusy(); got != 0 {
		t.Errorf("ALU RS busy = %d, want 0", got)
	}
}

func TestIssueSetsQiForDestination(t *testing.T) {
	core := NewCPU(DefaultConfig())
	word := encode(t, "addi x1, x0, 5")
	if !core.Fetch(word, 0) {
		t.Fatal("Fetch failed")
	}
	core.issueStage() // issue only, no dispatch/writeback
	if core.rf.Qi[1] == NoTag {
		t.Errorf("Qi[x1] = NoTag after issuing addi to x1")
	}
}

func TestIssueDoesNotSetQiForX0(t *testing.T) {
	core := NewCPU(DefaultConfig())
	word := encode(t, "addi x0, x0, 0")
	if !core.Fetch(word, 0) {
		t.Fatal("Fetch failed")
	}
	core.issueStage()
	if core.rf.Qi[0] != NoTag {
		t.Errorf("Qi[x0] must stay NoTag, got %d", core.rf.Qi[0])
	}
}

func TestStructuralStallWhenALUFull(t *testing.T) {
	cfg := DefaultConfig()
	cfg.ALURSCount = 1
	cfg.ALULatency = 5
	core := NewCPU(cfg)
	if !core.Fetch(encode(t, "addi x1, x0, 1"), 0) {
		t.Fatal("Fetch 1 failed")
	}
	if !core.Fetch(encode(t, "addi x2, x0, 2"), 0) {
		t.Fatal("Fetch 2 failed")
	}
	if !core.Fetch(encode(t, "addi x3, x0, 3"), 0) {
		t.Fatal("Fetch 3 failed")
	}
	for i := 0; i < 5; i++ {
		core.RunCycle()
	}
	if got := core.Stats().StructuralStalls; got < 1 {
		t.Errorf("StructuralStalls = %d, want >= 1", got)
	}
}

func TestIssueCapturesImmediateOperands(t *testing.T) {
	core := NewCPU(DefaultConfig())
	word := encode(t, "addi x1, x0, 42")
	if !core.Fetch(word, 0) {
		t.Fatal("Fetch failed")
	}
	core.issueStage()
	var found *RSEntry
	for i := range core.rs.alu {
		if core.rs.alu[i].Busy && core.rs.alu[i].Rd == 1 {
			found = &core.rs.alu[i]
			break
		}
	}
	if found == nil {
		t.Fatal("RS entry for x1 not found")
	}
	if found.Imm != 42 {
		t.Errorf("RS entry Imm = %d, want 42", found.Imm)
	}
}

func TestIssueTakesOperandValueWhenReady(t *testing.T) {
	core := NewCPU(DefaultConfig())
	if !core.Fetch(encode(t, "addi x1, x0, 7"), 0) {
		t.Fatal("Fetch 1 failed")
	}
	core.issueStage() // issue addi x1 first (this sets Qi[1])
	// Simulate the addi having written back by clearing Qi[1] and
	// putting the value into the architectural file.
	core.rf.V[1] = 7
	core.rf.Qi[1] = NoTag
	if !core.Fetch(encode(t, "add x2, x1, x0"), 0) {
		t.Fatal("Fetch 2 failed")
	}
	core.issueStage() // issue add x2; operand is now ready
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
	if found.Vj != 7 {
		t.Errorf("RS entry Vj = %d, want 7", found.Vj)
	}
	if found.Qj != NoTag {
		t.Errorf("RS entry Qj = %d, want NoTag (operand was ready)", found.Qj)
	}
}

func TestIssueSetsTagWhenOperandPending(t *testing.T) {
	core := NewCPU(DefaultConfig())
	if !core.Fetch(encode(t, "addi x1, x0, 7"), 0) {
		t.Fatal("Fetch 1 failed")
	}
	if !core.Fetch(encode(t, "add x2, x1, x0"), 0) {
		t.Fatal("Fetch 2 failed")
	}
	core.issueStage() // issue addi x1
	core.issueStage() // issue add x2
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
	if found.Qj == NoTag {
		t.Errorf("RS entry Qj = NoTag, want a tag (operand was pending)")
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

// countBusyLSUs returns the number of LSUs currently in the busy state.
func countBusyLSUs(c *CPU) int {
	n := 0
	for _, l := range c.lsus {
		if l.IsBusy() {
			n++
		}
	}
	return n
}
