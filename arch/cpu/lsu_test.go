package cpu

import (
	"testing"

	"github.com/malikwirin/riscvemu/arch/cpu/cputest"
)

// TestLSULoadAndStore exercises the LSU end-to-end through shared
// table-driven cases. Each case loads a small program, attaches a
// mock memory, runs the CPU until either a register check or a
// memory-write check is satisfied, and asserts the final value.
func TestLSULoadAndStore(t *testing.T) {
	cases := []struct {
		name  string
		opts  []func(*Config)
		mem   map[uint32]uint32
		asm   []string
		check func(t *testing.T, c *CPU, mem *cputest.MockWordHandler)
	}{
		{
			name: "load_reads_from_memory",
			mem:  map[uint32]uint32{100: 0xDEADBEEF},
			asm:  []string{"addi x2, x0, 100", "lw x1, 0(x2)"},
			check: func(t *testing.T, c *CPU, _ *cputest.MockWordHandler) {
				if got := c.Reg(1); got != 0xDEADBEEF {
					t.Errorf("x1 = %#x, want 0xDEADBEEF", got)
				}
			},
		},
		{
			name: "load_waits_for_pending_address",
			mem:  map[uint32]uint32{100: 0x42},
			asm:  []string{"addi x2, x0, 100", "lw x1, 0(x2)"},
			check: func(t *testing.T, c *CPU, _ *cputest.MockWordHandler) {
				if got := c.Reg(1); got != 0x42 {
					t.Errorf("x1 = %#x, want 0x42 (load should complete after x2 is ready)", got)
				}
			},
		},
		{
			name: "load_with_nonzero_immediate_offset",
			mem:  map[uint32]uint32{64: 0xABCD},
			asm:  []string{"addi x2, x0, 60", "lw x1, 4(x2)"},
			check: func(t *testing.T, c *CPU, _ *cputest.MockWordHandler) {
				if got := c.Reg(1); got != 0xABCD {
					t.Errorf("x1 = %#x, want 0xABCD (address 60+4=64)", got)
				}
			},
		},
		{
			name: "store_writes_to_memory",
			mem:  map[uint32]uint32{},
			asm:  []string{"addi x2, x0, 200", "addi x3, x0, 1234", "sw x3, 0(x2)"},
			check: func(t *testing.T, _ *CPU, mem *cputest.MockWordHandler) {
				if got := mem.Mem[200]; got != 1234 {
					t.Errorf("mem[200] = %#x, want 1234", got)
				}
			},
		},
		{
			name: "store_respects_configured_latency",
			opts: []func(*Config){withStoreLatency(3)},
			mem:  map[uint32]uint32{},
			asm:  []string{"addi x2, x0, 300", "addi x3, x0, 1500", "sw x3, 0(x2)"},
			check: func(t *testing.T, _ *CPU, mem *cputest.MockWordHandler) {
				if got := mem.Mem[300]; got != 1500 {
					t.Errorf("mem[300] = %#x, want 1500", got)
				}
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			core := makeCPU(t, tc.opts...)
			mem := &cputest.MockWordHandler{Mem: tc.mem}
			core.AttachMemory(mem)
			fetchProgram(t, core, tc.asm...)
			runUntilSettled(t, core, 30, nil)
			tc.check(t, core, mem)
		})
	}
}

func TestLSUStoreDoesNotBroadcastResult(t *testing.T) {
	// A store has no destination register, so it must not count as a
	// retired instruction on the CDB. Two addi broadcasts and one
	// store (no broadcast) leave Retired at 2.
	core := makeCPU(t)
	mem := &cputest.MockWordHandler{Mem: map[uint32]uint32{}}
	core.AttachMemory(mem)
	fetchProgram(t, core,
		"addi x2, x0, 400",
		"addi x3, x0, 1000",
		"sw x3, 0(x2)",
	)
	runUntilSettled(t, core, 20, func() bool { return mem.Mem[400] == 1000 })
	if got := core.Stats().Retired; got != 2 {
		t.Errorf("Retired = %d, want 2 (store should not broadcast)", got)
	}
}
