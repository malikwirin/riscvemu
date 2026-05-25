package arch

import (
	"testing"

	"github.com/malikwirin/riscvemu/arch/cpu"
	"github.com/malikwirin/riscvemu/assembler"
	"github.com/stretchr/testify/assert"
)

func TestCPU_InvalidJump(t *testing.T) {
	core := cpu.NewCPU()
	memory := NewMemory(64)
	core.PC = 0x1000
	_, err := memory.ReadWord(core.PC)
	assert.Error(t, err)
}

func TestCPU_Step_ReadsCorrectInstruction(t *testing.T) {
	core := cpu.NewCPU()
	memory := NewMemory(32)
	instr := assembler.Instruction(0x00112023)
	assert.NoError(t, memory.WriteWord(0, uint32(instr)))
	core.PC = 0
	assert.NoError(t, core.Step(memory))
	assert.Equal(t, uint32(4), core.PC)
}
