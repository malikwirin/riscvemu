package tests

import (
	"path/filepath"
	"testing"

	"github.com/malikwirin/riscvemu/arch"
	"github.com/malikwirin/riscvemu/assembler"
	"github.com/stretchr/testify/assert"
)

type exampleCase struct {
	filename   string
	expect     map[int]uint32
	steps      int
	memoryInit map[uint32]uint32
}

var exampleTests = []exampleCase{
	{
		filename: "../examples/1.asm",
		expect:   map[int]uint32{1: 5, 2: 10, 3: 15},
		steps:    3,
	},
	{
		filename: "../examples/2.asm",
		expect:   map[int]uint32{1: 42, 2: 100, 3: 42},
		steps:    4,
	},
	{
		filename: "../examples/3.asm",
		expect:   map[int]uint32{1: 7, 2: 7, 3: 99},
		steps:    5,
	},
	{
		filename: "../examples/4.asm",
		expect:   map[int]uint32{2: 2},
		steps:    3,
	},
	{
		filename: "../examples/5.asm",
		expect:   map[int]uint32{1: 0},
		steps:    11,
	},
	{
		filename: "../examples/6.asm",
		expect:   map[int]uint32{1: 10, 2: 10, 3: 55},
		steps:    31,
	},
	{
		filename: "../examples/7.asm",
		expect:   map[int]uint32{4: 13},
		steps:    35,
	},
	{
		filename: "../examples/8.asm",
		expect:   map[int]uint32{3: 123},
		steps:    95,
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
		steps:    7,
	},
	{
        filename: "../examples/tribonacci.asm",
        expect:   map[int]uint32{2: 13}, // tribonacci(7) = 13 in x2
        steps:    30,
	},
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

			steps := tc.steps
			if steps == 0 {
				steps = len(prog) + 5
			}
			for i := 0; i < steps; i++ {
				_ = m.Step()
			}

			for reg, want := range tc.expect {
				got := m.CPU.Reg[reg]
				assert.Equalf(t, want, got, "Register x%d: expected %d, got %d", reg, want, got)
			}
		})
	}
}
