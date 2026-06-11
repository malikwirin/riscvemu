package cpu

import "testing"

func TestALUBEQEqualTaken(t *testing.T) {
	core := NewCPU(DefaultConfig())
	if !core.Fetch(encode(t, "addi x1, x0, 5"), 0) {
		t.Fatal("Fetch failed")
	}
	if !core.Fetch(encode(t, "addi x2, x0, 5"), 4) {
		t.Fatal("Fetch failed")
	}
	if !core.Fetch(encode(t, "beq x1, x2, 12"), 8) {
		t.Fatal("Fetch failed")
	}
	for i := 0; i < 20; i++ {
		core.RunCycle()
		if core.LastBranch().IsBranch {
			break
		}
	}
	if !core.LastBranch().IsBranch {
		t.Fatalf("LastBranch.IsBranch = false after 20 cycles")
	}
	if !core.LastBranch().Taken {
		t.Errorf("LastBranch.Taken = false, want true (5 == 5)")
	}
}

func TestALUBEQNotEqualNotTaken(t *testing.T) {
	core := NewCPU(DefaultConfig())
	if !core.Fetch(encode(t, "addi x1, x0, 5"), 0) {
		t.Fatal("Fetch failed")
	}
	if !core.Fetch(encode(t, "addi x2, x0, 7"), 4) {
		t.Fatal("Fetch failed")
	}
	if !core.Fetch(encode(t, "beq x1, x2, 12"), 8) {
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
	if core.LastBranch().Taken {
		t.Errorf("LastBranch.Taken = true, want false (5 != 7)")
	}
}

func TestALUBNE(t *testing.T) {
	core := NewCPU(DefaultConfig())
	if !core.Fetch(encode(t, "addi x1, x0, 5"), 0) {
		t.Fatal("Fetch failed")
	}
	if !core.Fetch(encode(t, "addi x2, x0, 7"), 4) {
		t.Fatal("Fetch failed")
	}
	if !core.Fetch(encode(t, "bne x1, x2, 12"), 8) {
		t.Fatal("Fetch failed")
	}
	for i := 0; i < 10; i++ {
		core.RunCycle()
		if core.LastBranch().IsBranch {
			break
		}
	}
	if !core.LastBranch().Taken {
		t.Errorf("LastBranch.Taken = false, want true (5 != 7)")
	}
}

func TestALUBLT(t *testing.T) {
	core := NewCPU(DefaultConfig())
	if !core.Fetch(encode(t, "addi x1, x0, -1"), 0) { // 0xFFFFFFFF
		t.Fatal("Fetch failed")
	}
	if !core.Fetch(encode(t, "addi x2, x0, 1"), 4) {
		t.Fatal("Fetch failed")
	}
	if !core.Fetch(encode(t, "blt x1, x2, 12"), 8) {
		t.Fatal("Fetch failed")
	}
	for i := 0; i < 10; i++ {
		core.RunCycle()
		if core.LastBranch().IsBranch {
			break
		}
	}
	if !core.LastBranch().Taken {
		t.Errorf("LastBranch.Taken = false, want true (-1 < 1 signed)")
	}
}

func TestALUBLTUnsignedComparison(t *testing.T) {
	// x1 = 0xFFFFFFFF (unsigned: huge), x2 = 1 (unsigned: small)
	// blt is signed: -1 < 1 -> taken
	core := NewCPU(DefaultConfig())
	if !core.Fetch(encode(t, "addi x1, x0, -1"), 0) {
		t.Fatal("Fetch failed")
	}
	if !core.Fetch(encode(t, "addi x2, x0, 1"), 4) {
		t.Fatal("Fetch failed")
	}
	if !core.Fetch(encode(t, "blt x1, x2, 12"), 8) {
		t.Fatal("Fetch failed")
	}
	for i := 0; i < 10; i++ {
		core.RunCycle()
		if core.LastBranch().IsBranch {
			break
		}
	}
	if !core.LastBranch().Taken {
		t.Errorf("LastBranch.Taken = false, want true (signed -1 < 1)")
	}
}

func TestLastBranchResetForNonBranch(t *testing.T) {
	// After a non-branch instruction completes, LastBranch should be reset
	// (so the Machine can detect "no branch happened this cycle").
	core := NewCPU(DefaultConfig())
	if !core.Fetch(encode(t, "addi x1, x0, 5"), 0) {
		t.Fatal("Fetch failed")
	}
	core.RunCycle()
	core.RunCycle() // addi completes
	if core.LastBranch().IsBranch {
		t.Errorf("LastBranch.IsBranch = true after non-branch, want false")
	}
}
