package cpu

import (
	"testing"

	"codeberg.org/malik/riscvemu/arch/cpu/cputest"
)

func TestJALWritesLinkAndTarget(t *testing.T) {
	core := makeCPU(t)
	fetchProgram(t, core, "jal x5, 12")
	runUntilBranch(t, core, 10)
	lb := core.LastBranch()
	if !lb.IsBranch {
		t.Fatalf("LastBranch.IsBranch = false")
	}
	if !lb.Taken {
		t.Errorf("Taken = false, want true (JAL is unconditional)")
	}
	if got := core.Reg(5); got != 4 {
		t.Errorf("x5 = %d, want 4 (PC+4 of JAL at PC=0)", got)
	}
	if got := lb.Target; got != 12 {
		t.Errorf("Target = %d, want 12", got)
	}
}

func TestJALRComputesTargetAndLink(t *testing.T) {
	cases := []struct {
		name    string
		baseAsm string
		jalrAsm string
		basePC  uint32
		jalrPC  uint32
		wantReg uint32
		wantLnk uint32
		wantTgt uint32
	}{
		{
			name:    "target_from_register_with_offset",
			baseAsm: "addi x2, x0, 100",
			jalrAsm: "jalr x6, 16(x2)",
			basePC:  0,
			jalrPC:  4,
			wantReg: 6,
			wantLnk: 8,   // PC+4 of JALR at PC=4
			wantTgt: 116, // (100 + 16) & ~1
		},
		{
			name:    "alignment_clears_low_bit",
			baseAsm: "addi x2, x0, 5",
			jalrAsm: "jalr x6, 0(x2)",
			basePC:  0,
			jalrPC:  4,
			wantReg: 6,
			wantLnk: 8,
			wantTgt: 4, // 5 & ~1
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			core := makeCPU(t)
			if !core.Fetch(cputest.Encode(t, tc.baseAsm), tc.basePC) {
				t.Fatalf("Fetch %q failed", tc.baseAsm)
			}
			if !core.Fetch(cputest.Encode(t, tc.jalrAsm), tc.jalrPC) {
				t.Fatalf("Fetch %q failed", tc.jalrAsm)
			}
			runUntilBranch(t, core, 20)
			lb := core.LastBranch()
			if !lb.IsBranch {
				t.Fatalf("LastBranch.IsBranch = false")
			}
			if !lb.Taken {
				t.Errorf("Taken = false, want true (JALR is unconditional)")
			}
			if got := core.Reg(tc.wantReg); got != tc.wantLnk {
				t.Errorf("x%d = %d, want %d (link PC+4)", tc.wantReg, got, tc.wantLnk)
			}
			if got := lb.Target; got != tc.wantTgt {
				t.Errorf("Target = %d, want %d", got, tc.wantTgt)
			}
		})
	}
}

func TestJALDoesNotStallOnBranch(t *testing.T) {
	// JAL is unconditional; it should NOT block the issue stage for the
	// next instruction. After the JAL retires, the following addi must
	// be able to issue immediately, so BranchStalls stays at 0.
	core := makeCPU(t)
	fetchProgram(t, core,
		"addi x1, x0, 5",
		"jal x5, 12",
		"addi x3, x0, 1",
	)
	runUntilBranch(t, core, 20)
	if got := core.Stats().BranchStalls; got != 0 {
		t.Errorf("BranchStalls = %d, want 0 (JAL is unconditional)", got)
	}
}
