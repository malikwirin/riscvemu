package tests

import (
	"path/filepath"
	"testing"

	"codeberg.org/malik/riscvemu/arch"
	"codeberg.org/malik/riscvemu/assembler"
	"github.com/stretchr/testify/assert"
)

type exampleCase struct {
	filename   string
	expect     map[int]uint32
	memoryInit map[uint32]uint32
}

var exampleTests = []exampleCase{
	{
		filename: "../examples/1.asm",
		expect:   map[int]uint32{1: 5, 2: 10, 3: 15},
	},
	{
		filename: "../examples/2.asm",
		expect:   map[int]uint32{1: 42, 2: 100, 3: 42},
	},
	{
		filename: "../examples/3.asm",
		expect:   map[int]uint32{1: 7, 2: 7, 3: 99},
	},
	{
		filename: "../examples/4.asm",
		expect:   map[int]uint32{2: 2},
	},
	{
		filename: "../examples/5.asm",
		expect:   map[int]uint32{1: 0},
	},
	{
		filename: "../examples/6.asm",
		expect:   map[int]uint32{1: 10, 2: 10, 3: 55},
	},
	{
		filename: "../examples/7.asm",
		expect:   map[int]uint32{4: 13},
	},
	{
		filename: "../examples/8.asm",
		expect:   map[int]uint32{3: 123},
		memoryInit: map[uint32]uint32{
			100: 1,
			104: 2,
			108: 3,
			112: 123,
			116: 4,
		},
	},
	{
		filename: "../examples/9.asm",
		expect:   map[int]uint32{6: 42},
	},
	{
		filename: "../examples/tribonacci.asm",
		expect:   map[int]uint32{2: 13}, // tribonacci(7) = 13 in x2
	},
}

// allSettled reports whether the CPU registers match the expected map.
func allSettled(m *arch.Machine, expect map[int]uint32) bool {
	for reg, want := range expect {
		if m.CPU.Reg(uint32(reg)) != want {
			return false
		}
	}
	return true
}

// runUntilSettled steps the machine until every entry in expect matches the
// CPU's architectural register file, or until the budget is exhausted. The
// budget is a generous upper bound (4 cycles per instruction plus 20 cycles
// of slack) so this works for every example regardless of the Tomasulo
// pipeline depth, ALU/LSU latencies, or branch stalls. Tests stay robust
// against future config changes.
func runUntilSettled(t *testing.T, m *arch.Machine, prog []assembler.Instruction, expect map[int]uint32) {
	t.Helper()
	budget := 20*len(prog) + 100
	for i := 0; i < budget; i++ {
		if err := m.Step(); err != nil {
			t.Fatalf("Step %d: %v", i, err)
		}
		if allSettled(m, expect) {
			return
		}
	}
}

func TestExamplesIntegration(t *testing.T) {
	for _, tc := range exampleTests {
		tc := tc
		t.Run(filepath.Base(tc.filename), func(t *testing.T) {
			prog, err := assembler.AssembleFile(tc.filename)
			assert.NoError(t, err)

			m := arch.NewMachine(1024)

			for addr, val := range tc.memoryInit {
				err := m.Memory.WriteWord(addr, val)
				assert.NoErrorf(t, err, "Memory init failed at 0x%X", addr)
			}

			err = m.LoadProgram(prog, 0)
			assert.NoError(t, err)

			runUntilSettled(t, m, prog, tc.expect)

			for reg, want := range tc.expect {
				got := m.CPU.Reg(uint32(reg))
				assert.Equalf(t, want, got, "Register x%d: expected %d, got %d", reg, want, got)
			}
		})
	}
}
