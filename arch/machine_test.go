package arch

import (
	"testing"

	"github.com/malikwirin/riscvemu/assembler"
	"github.com/stretchr/testify/assert"
)

func TestMachineInitialization(t *testing.T) {
	m := NewMachine(1024)
	assert.NotNil(t, m.CPU, "CPU should not be nil after initialization")
	assert.NotNil(t, m.Memory, "Memory should not be nil after initialization")
	assert.Equal(t, 1024, len(m.Memory.Data), "Expected memory size 1024")
	assert.Equal(t, uint32(0), m.CPU.PC, "Expected PC to be 0")
	for i, reg := range m.CPU.Reg {
		assert.Equalf(t, uint32(0), reg, "Expected register x%d to be 0", i)
	}
}

func TestMachineStepIncreasesPC(t *testing.T) {
	m := NewMachine(2048)
	oldPC := m.CPU.PC
	instr, _ := assembler.ParseInstruction("addi x0, x0, 0")
	m.Memory.Data[0] = byte(instr)
	m.Memory.Data[1] = byte(instr >> 8)
	m.Memory.Data[2] = byte(instr >> 16)
	m.Memory.Data[3] = byte(instr >> 24)
	err := m.Step()
	assert.NoError(t, err, "Step returned error")
	assert.Equal(t, oldPC+4, m.CPU.PC, "Expected PC to increase by 4")
}

func TestMachineReset(t *testing.T) {
	m := NewMachine(128)
	m.CPU.PC = 100
	m.CPU.Reg[5] = 42
	m.Memory.Data[10] = 0xFF
	m.Reset()
	assert.Equal(t, uint32(0), m.CPU.PC, "After reset, expected PC to be 0")
	for i, reg := range m.CPU.Reg {
		assert.Equalf(t, uint32(0), reg, "After reset, expected register x%d to be 0", i)
	}
	for i, b := range m.Memory.Data {
		assert.Equalf(t, byte(0), b, "After reset, expected memory at %d to be 0", i)
	}
}

func TestMachineLoadProgram(t *testing.T) {
	m := NewMachine(64)
	program := []assembler.Instruction{0xDEADBEEF, 0x12345678, 0xCAFEBABE}
	startAddr := uint32(8)

	err := m.LoadProgram(program, startAddr)
	assert.NoError(t, err, "LoadProgram returned error")

	// Check if instructions are loaded at the correct addresses
	for i, want := range program {
		addr := startAddr + uint32(i*4)
		got, err := m.Memory.ReadWord(addr)
		assert.NoErrorf(t, err, "ReadWord failed at addr %d", addr)
		assert.Equalf(t, uint32(want), got, "Instruction at %d", addr)
	}

	assert.Equal(t, startAddr, m.CPU.PC, "Expected PC to be set to startAddr after LoadProgram")
}
