package cpu

import (
	"testing"

	"github.com/malikwirin/riscvemu/arch/cpu/cputest"
)

func TestBranchOutcomes(t *testing.T) {
	// Each subtest loads two addi values, then a branch, then runs
	// the CPU until the branch retires. The check is the taken/not
	// outcome of that branch.
	cases := []struct {
		name      string
		x1, x2    int32
		branchAsm string
		wantTaken bool
	}{
		{"beq_equal_taken", 5, 5, "beq x1, x2, 12", true},
		{"beq_unequal_not_taken", 5, 7, "beq x1, x2, 12", false},
		{"bne_unequal_taken", 5, 7, "bne x1, x2, 12", true},
		{"blt_signed_taken", -1, 1, "blt x1, x2, 12", true},
		{"blt_signed_not_taken", 1, -1, "blt x1, x2, 12", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			core := makeCPU(t)
			fetchProgram(t, core,
				"addi x1, x0, "+itoa(tc.x1),
				"addi x2, x0, "+itoa(tc.x2),
				tc.branchAsm,
			)
			runUntilBranch(t, core, 20)
			if got := core.LastBranch().Taken; got != tc.wantTaken {
				t.Errorf("Taken = %v, want %v", got, tc.wantTaken)
			}
		})
	}
}

func TestLastBranchResetForNonBranch(t *testing.T) {
	// After a non-branch instruction completes, LastBranch should be reset
	// (so the Machine can detect "no branch happened this cycle").
	core := makeCPU(t)
	fetchProgram(t, core, "addi x1, x0, 5")
	core.RunCycle()
	core.RunCycle() // addi completes
	if core.LastBranch().IsBranch {
		t.Errorf("LastBranch.IsBranch = true after non-branch, want false")
	}
}

// itoa is a tiny helper so the table-driven test can pass signed
// decimal immediates to the assembler. The assembler's I-format
// instruction parser accepts "-1" directly, so a string conversion
// is all we need.
func itoa(n int32) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var buf [16]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}

// _ keeps cputest referenced even when individual subtests above do
// not need Encode directly (e.g. when fetchProgram handles the
// encoding inline). cputest.Encode is the public entry point used
// by tests that prefer the one-call form.
var _ = cputest.Encode
