package cpu

import "testing"

func TestJALWritesLinkRegister(t *testing.T) {
	core := NewCPU(DefaultConfig())
	if !core.Fetch(encode(t, "jal x5, 12"), 0) {
		t.Fatal("Fetch failed")
	}
	for i := 0; i < 10; i++ {
		core.RunCycle()
		if core.LastBranch().IsBranch {
			break
		}
	}
	if !core.LastBranch().IsBranch {
		t.Fatalf("LastBranch.IsBranch = false, want true")
	}
	if !core.LastBranch().Taken {
		t.Errorf("LastBranch.Taken = false, want true (JAL is unconditional)")
	}
	// Link register should be PC+4 = 0+4 = 4
	if got := core.Reg(5); got != 4 {
		t.Errorf("x5 = %d, want 4 (PC+4)", got)
	}
	// Target = 0 + 12 = 12
	if got := core.LastBranch().Target; got != 12 {
		t.Errorf("LastBranch.Target = %d, want 12", got)
	}
}

func TestJALRTargetFromRegister(t *testing.T) {
	core := NewCPU(DefaultConfig())
	if !core.Fetch(encode(t, "addi x2, x0, 100"), 0) {
		t.Fatal("Fetch failed")
	}
	if !core.Fetch(encode(t, "jalr x6, 16(x2)"), 4) {
		t.Fatal("Fetch failed")
	}
	for i := 0; i < 20; i++ {
		core.RunCycle()
		if core.LastBranch().IsBranch {
			break
		}
	}
	if !core.LastBranch().IsBranch {
		t.Fatalf("LastBranch.IsBranch = false, want true")
	}
	if !core.LastBranch().Taken {
		t.Errorf("LastBranch.Taken = false, want true (JALR is unconditional)")
	}
	// Link register should be PC+4 = 4+4 = 8
	if got := core.Reg(6); got != 8 {
		t.Errorf("x6 = %d, want 8 (PC+4)", got)
	}
	// Target = (100 + 16) & ~1 = 116
	if got := core.LastBranch().Target; got != 116 {
		t.Errorf("LastBranch.Target = %d, want 116 (low bit cleared)", got)
	}
}

func TestJALRAlignmentClearsLowBit(t *testing.T) {
	// When Rs1 + imm has the low bit set, the JALR target clears it.
	// We use a small base and a 1-bit offset that, when sign-extended,
	// lands us on an odd address. Use addi x2, x0, 5 (base 5, low bit set)
	// and jalr x6, 0(x2) (imm 0, target = 5 & ~1 = 4).
	core := NewCPU(DefaultConfig())
	if !core.Fetch(encode(t, "addi x2, x0, 5"), 0) {
		t.Fatal("Fetch failed")
	}
	if !core.Fetch(encode(t, "jalr x6, 0(x2)"), 4) {
		t.Fatal("Fetch failed")
	}
	for i := 0; i < 20; i++ {
		core.RunCycle()
		if core.LastBranch().IsBranch {
			break
		}
	}
	// Target should be 5 & ~1 = 4 (low bit cleared)
	if got := core.LastBranch().Target; got != 4 {
		t.Errorf("LastBranch.Target = %d, want 4 (low bit cleared from 5)", got)
	}
}

func TestJALDoesNotStallOnBranch(t *testing.T) {
	// JAL is unconditional; it should NOT block the issue stage for the
	// next instruction (unlike conditional branches, which the user asked
	// to stall).
	core := NewCPU(DefaultConfig())
	if !core.Fetch(encode(t, "addi x1, x0, 5"), 0) {
		t.Fatal("Fetch 1 failed")
	}
	if !core.Fetch(encode(t, "jal x5, 12"), 4) {
		t.Fatal("Fetch 2 failed")
	}
	if !core.Fetch(encode(t, "addi x3, x0, 1"), 8) {
		t.Fatal("Fetch 3 failed")
	}
	for i := 0; i < 20; i++ {
		core.RunCycle()
		if core.LastBranch().IsBranch {
			break
		}
	}
	// After the JAL resolves, the addi x3 should be able to issue in the
	// very next cycle, so the BranchStalls counter should not grow beyond 0.
	if got := core.Stats().BranchStalls; got != 0 {
		t.Errorf("BranchStalls = %d, want 0 (JAL is unconditional)", got)
	}
}
