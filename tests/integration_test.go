package tests

import (
	"path/filepath"
	"testing"

	"github.com/malikwirin/riscvemu/arch"
	"github.com/malikwirin/riscvemu/assembler"
	"github.com/stretchr/testify/assert"
)

// exampleCase defines a test case for an assembly example.
type exampleCase struct {
	filename     string
	expect       map[int]uint32    // expected register values after execution
	steps        int               // number of instructions to execute
	memoryInit   map[uint32]uint32 // address -> value, initial memory state
}

var exampleTests = []exampleCase{
	{
		filename:   "../examples/1.asm",
		expect:     map[int]uint32{1: 5, 2: 10, 3: 15},
		steps:      3,
	},
	{
		filename:   "../examples/2.asm",
		expect:     map[int]uint32{1: 42, 2: 100, 3: 42},
		steps:      4,
	},
	{
		filename:   "../examples/8.asm",
		expect:     map[int]uint32{3: 123}, // expected register values
		steps:      100,
		memoryInit: map[uint32]uint32{
			100: 1,
			104: 2,
			108: 3,
			112: 123,
			116: 4,
		},
	},
}

func TestExamplesIntegration(t *testing.T) {
	for _, tc := range exampleTests {
		tc := tc // capture range variable
		t.Run(filepath.Base(tc.filename), func(t *testing.T) {
			prog, err := assembler.AssembleFile(tc.filename)
			assert.NoError(t, err, "Failed to assemble %s", tc.filename)

			// Use a sufficiently large memory for all examples
			m := arch.NewMachine(1024)

			// Optionally initialize memory as required by the example
			for addr, val := range tc.memoryInit {
				err := m.Memory.WriteWord(addr, val)
				assert.NoErrorf(t, err, "Failed to initialize memory at 0x%X", addr)
			}

			err = m.LoadProgram(prog, 0)
			assert.NoError(t, err, "Failed to load program")

			// Run the specified number of steps
			steps := tc.steps
			if steps == 0 {
				steps = len(prog) + 5 // fallback: run a bit more than the program size
			}
			for i := 0; i < steps; i++ {
				_ = m.Step()
			}

			// Check all expected register values
			for reg, want := range tc.expect {
				got := m.CPU.Reg[reg]
				assert.Equalf(t, want, got, "Register x%d: expected %d, got %d", reg, want, got)
			}
		})
	}
}
