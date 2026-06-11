package cpu

import "testing"

func TestALUBEQEqualTaken(t *testing.T) {
    core := NewCPU(DefaultConfig())
    if err := core.ReceiveInstruction(encode(t, "addi x1, x0, 5"), 0); err != nil {
        t.Fatal(err)
    }
    if err := core.ReceiveInstruction(encode(t, "addi x2, x0, 5"), 4); err != nil {
        t.Fatal(err)
    }
    if err := core.ReceiveInstruction(encode(t, "beq x1, x2, 12"), 8); err != nil {
        t.Fatal(err)
    }
    for i := 0; i < 10; i++ {
        core.RunCycle()
        if got := core.LastBranch().Taken; core.LastBranch().IsBranch && got {
            return
        }
    }
    if !core.LastBranch().IsBranch {
        t.Fatalf("LastBranch.IsBranch = false, want true")
    }
    if !core.LastBranch().Taken {
        t.Errorf("LastBranch.Taken = false, want true (5 == 5)")
    }
}

func TestALUBEQNotEqualNotTaken(t *testing.T) {
    core := NewCPU(DefaultConfig())
    if err := core.ReceiveInstruction(encode(t, "addi x1, x0, 5"), 0); err != nil {
        t.Fatal(err)
    }
    if err := core.ReceiveInstruction(encode(t, "addi x2, x0, 7"), 4); err != nil {
        t.Fatal(err)
    }
    if err := core.ReceiveInstruction(encode(t, "beq x1, x2, 12"), 8); err != nil {
        t.Fatal(err)
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
    if err := core.ReceiveInstruction(encode(t, "addi x1, x0, 5"), 0); err != nil {
        t.Fatal(err)
    }
    if err := core.ReceiveInstruction(encode(t, "addi x2, x0, 7"), 4); err != nil {
        t.Fatal(err)
    }
    if err := core.ReceiveInstruction(encode(t, "bne x1, x2, 12"), 8); err != nil {
        t.Fatal(err)
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
    if err := core.ReceiveInstruction(encode(t, "addi x1, x0, -1"), 0); err != nil { // 0xFFFFFFFF
        t.Fatal(err)
    }
    if err := core.ReceiveInstruction(encode(t, "addi x2, x0, 1"), 4); err != nil {
        t.Fatal(err)
    }
    if err := core.ReceiveInstruction(encode(t, "blt x1, x2, 12"), 8); err != nil {
        t.Fatal(err)
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
    if err := core.ReceiveInstruction(encode(t, "addi x1, x0, -1"), 0); err != nil {
        t.Fatal(err)
    }
    if err := core.ReceiveInstruction(encode(t, "addi x2, x0, 1"), 4); err != nil {
        t.Fatal(err)
    }
    if err := core.ReceiveInstruction(encode(t, "blt x1, x2, 12"), 8); err != nil {
        t.Fatal(err)
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
    if err := core.ReceiveInstruction(encode(t, "addi x1, x0, 5"), 0); err != nil {
        t.Fatal(err)
    }
    core.RunCycle()
    core.RunCycle() // addi completes
    if core.LastBranch().IsBranch {
        t.Errorf("LastBranch.IsBranch = true after non-branch, want false")
    }
}
