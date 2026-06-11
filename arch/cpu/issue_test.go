package cpu

import (
	"testing"

	"github.com/malikwirin/riscvemu/arch/cpu/cputest"
)

func TestIssuePlacesInMatchingPool(t *testing.T) {
	cases := []struct {
		name     string
		opts     []func(*Config)
		asm      string
		rsBefore int
		rsAfter  int
		otherRS  int
	}{
		{
			name:    "addi_goes_to_alu_pool",
			opts:    []func(*Config){withALULatency(2)},
			asm:     "addi x1, x0, 5",
			rsAfter: 1,
		},
		{
			name:    "lw_goes_to_lsu_pool",
			opts:    []func(*Config){withLoadLatency(3)},
			asm:     "lw x3, 0(x2)",
			rsAfter: 1,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			core := makeCPU(t, tc.opts...)
			if !core.Fetch(cputest.Encode(t, tc.asm), 0) {
				t.Fatal("Fetch failed")
			}
			core.RunCycle() // dispatch + step
			// We just check the right pool is busy after issue; the
			// exact FU count is asserted per-case below.
			if tc.rsAfter == 1 {
				if tc.name == "addi_goes_to_alu_pool" {
					if got := countBusyALUs(core); got != 1 {
						t.Errorf("ALU busy = %d, want 1", got)
					}
					if got := core.rs.LSUCountBusy(); got != 0 {
						t.Errorf("LSU busy = %d, want 0", got)
					}
				} else {
					if got := countBusyLSUs(core); got != 1 {
						t.Errorf("LSU busy = %d, want 1", got)
					}
					if got := core.rs.ALUCountBusy(); got != 0 {
						t.Errorf("ALU busy = %d, want 0", got)
					}
				}
			}
		})
	}
}

func TestIssueSetsQiForDestination(t *testing.T) {
	core := makeCPU(t)
	if !core.Fetch(cputest.Encode(t, "addi x1, x0, 5"), 0) {
		t.Fatal("Fetch failed")
	}
	core.issueStage()
	if core.rf.Qi[1] == NoTag {
		t.Errorf("Qi[x1] = NoTag after issuing addi to x1")
	}
}

func TestIssueDoesNotSetQiForX0(t *testing.T) {
	core := makeCPU(t)
	if !core.Fetch(cputest.Encode(t, "addi x0, x0, 0"), 0) {
		t.Fatal("Fetch failed")
	}
	core.issueStage()
	if core.rf.Qi[0] != NoTag {
		t.Errorf("Qi[x0] must stay NoTag, got %d", core.rf.Qi[0])
	}
}

func TestIssueCapturesImmediateOperand(t *testing.T) {
	core := makeCPU(t)
	if !core.Fetch(cputest.Encode(t, "addi x1, x0, 42"), 0) {
		t.Fatal("Fetch failed")
	}
	core.issueStage()
	entry := findRSEntry(t, core, 1)
	if entry.Imm != 42 {
		t.Errorf("RS entry Imm = %d, want 42", entry.Imm)
	}
}

func TestIssueOperandCapture(t *testing.T) {
	// Both operand-capture paths (value-when-ready, tag-when-pending)
	// share the same setup; the difference is whether the producer
	// has retired before the consumer is issued.
	cases := []struct {
		name      string
		setup     func(t *testing.T, c *CPU)
		wantVj    uint32
		wantQjTag bool // true -> Qj must be a real tag, not NoTag
	}{
		{
			name: "value_captured_when_producer_retired",
			setup: func(t *testing.T, c *CPU) {
				if !c.Fetch(cputest.Encode(t, "addi x1, x0, 7"), 0) {
					t.Fatal("Fetch 1 failed")
				}
				c.issueStage() // issues addi x1, sets Qi[1]
				c.rf.V[1] = 7
				c.rf.Qi[1] = NoTag
				if !c.Fetch(cputest.Encode(t, "add x2, x1, x0"), 0) {
					t.Fatal("Fetch 2 failed")
				}
				c.issueStage() // operand is now ready
			},
			wantVj:    7,
			wantQjTag: false,
		},
		{
			name: "tag_captured_when_producer_pending",
			setup: func(t *testing.T, c *CPU) {
				if !c.Fetch(cputest.Encode(t, "addi x1, x0, 7"), 0) {
					t.Fatal("Fetch 1 failed")
				}
				if !c.Fetch(cputest.Encode(t, "add x2, x1, x0"), 0) {
					t.Fatal("Fetch 2 failed")
				}
				c.issueStage() // addi x1
				c.issueStage() // add x2 with pending Qj
			},
			wantVj:    0,
			wantQjTag: true,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			core := makeCPU(t)
			tc.setup(t, core)
			entry := findRSEntry(t, core, 2)
			if entry.Vj != tc.wantVj {
				t.Errorf("Vj = %d, want %d", entry.Vj, tc.wantVj)
			}
			if tc.wantQjTag {
				if entry.Qj == NoTag {
					t.Errorf("Qj = NoTag, want a tag (operand was pending)")
				}
			} else {
				if entry.Qj != NoTag {
					t.Errorf("Qj = %d, want NoTag (operand was ready)", entry.Qj)
				}
			}
		})
	}
}

func TestStructuralStallWhenALUFull(t *testing.T) {
	core := makeCPU(t, withALURSCount(1), withALULatency(5))
	fetchProgram(t, core,
		"addi x1, x0, 1",
		"addi x2, x0, 2",
		"addi x3, x0, 3",
	)
	for i := 0; i < 5; i++ {
		core.RunCycle()
	}
	if got := core.Stats().StructuralStalls; got < 1 {
		t.Errorf("StructuralStalls = %d, want >= 1", got)
	}
}

// findRSEntry returns the ALU RS entry whose Rd matches rd, failing
// the test if no such entry exists.
func findRSEntry(t *testing.T, c *CPU, rd uint32) *RSEntry {
	t.Helper()
	for i := range c.rs.alu {
		if c.rs.alu[i].Busy && c.rs.alu[i].Rd == rd {
			return &c.rs.alu[i]
		}
	}
	t.Fatalf("RS entry for x%d not found", rd)
	return nil
}
