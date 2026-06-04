package cpu

import "testing"

func TestWriteBackUpdatesRegister(t *testing.T) {
	core := NewCPU(DefaultConfig())
	if err := core.ReceiveInstruction(encode(t, "addi x1, x0, 5")); err != nil {
		t.Fatal(err)
	}
	core.RunCycle() // dispatch x1 to ALU[0]
	core.RunCycle() // step: ALU[0] finishes; writeback broadcasts
	if got := core.Reg(1); got != 5 {
		t.Errorf("x1 = %d, want 5", got)
	}
}

func TestWriteBackWakesRSEntryOnCDB(t *testing.T) {
	core := NewCPU(DefaultConfig())
	// x1 <- 5
	if err := core.ReceiveInstruction(encode(t, "addi x1, x0, 5")); err != nil {
		t.Fatal(err)
	}
	// x2 <- x1 + 0  (Qj captured from Qi[1])
	if err := core.ReceiveInstruction(encode(t, "add x2, x1, x0")); err != nil {
		t.Fatal(err)
	}
	core.RunCycle() // dispatch x1
	core.RunCycle() // x1 done; writeback wakes x2's operand
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
	if err := core.ReceiveInstruction(encode(t, "addi x1, x0, 5")); err != nil {
		t.Fatal(err)
	}
	core.RunCycle()
	core.RunCycle()
	if got := core.Stats().Retired; got != 1 {
		t.Errorf("Retired = %d, want 1", got)
	}
}

func TestWriteBackIncrementsRAWResolved(t *testing.T) {
	core := NewCPU(DefaultConfig())
	if err := core.ReceiveInstruction(encode(t, "addi x1, x0, 5")); err != nil {
		t.Fatal(err)
	}
	if err := core.ReceiveInstruction(encode(t, "add x2, x1, x0")); err != nil {
		t.Fatal(err)
	}
	core.RunCycle() // dispatch x1
	core.RunCycle() // x1 result wakes x2 operand
	if got := core.Stats().RAWResolved; got < 1 {
		t.Errorf("RAWResolved = %d, want >= 1", got)
	}
}

func TestWriteBackArbitratesOnePerCycle(t *testing.T) {
	cfg := DefaultConfig()
	cfg.ALURSCount = 2
	cfg.ALULatency = 1
	core := NewCPU(cfg)
	if err := core.ReceiveInstruction(encode(t, "addi x1, x0, 1")); err != nil {
		t.Fatal(err)
	}
	if err := core.ReceiveInstruction(encode(t, "addi x2, x0, 2")); err != nil {
		t.Fatal(err)
	}
	core.RunCycle() // dispatch both
	core.RunCycle() // both finish, but only one broadcast
	if got := core.Stats().Retired; got != 1 {
		t.Errorf("Retired after 2 cycles = %d, want 1", got)
	}
	core.RunCycle() // the other one broadcasts
	if got := core.Stats().Retired; got != 2 {
		t.Errorf("Retired after 3 cycles = %d, want 2", got)
	}
}

func TestWriteBackClearsQi(t *testing.T) {
	core := NewCPU(DefaultConfig())
	if err := core.ReceiveInstruction(encode(t, "addi x1, x0, 5")); err != nil {
		t.Fatal(err)
	}
	if core.rf.Qi[1] == NoTag {
		t.Fatal("Qi[x1] should be set after issue")
	}
	core.RunCycle()
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
	if err := core.ReceiveInstruction(encode(t, "addi x1, x0, 5")); err != nil {
		t.Fatal(err)
	}
	// x2 <- x1 + 0
	if err := core.ReceiveInstruction(encode(t, "add x2, x1, x0")); err != nil {
		t.Fatal(err)
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
	if err := core.ReceiveInstruction(encode(t, "addi x1, x0, 10")); err != nil {
		t.Fatal(err)
	}
	// x2 <- 20
	if err := core.ReceiveInstruction(encode(t, "addi x2, x0, 20")); err != nil {
		t.Fatal(err)
	}
	// x3 <- x1 + x2
	if err := core.ReceiveInstruction(encode(t, "add x3, x1, x2")); err != nil {
		t.Fatal(err)
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
