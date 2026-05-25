package arch

import (
	"fmt"

	"github.com/malikwirin/riscvemu/arch/cpu"
	"github.com/malikwirin/riscvemu/assembler"
)

type Machine struct {
	CPU    *cpu.CPU
	Memory *Memory
}

func NewMachine(memSize int) *Machine {
	return &Machine{
		CPU:    cpu.NewCPU(),
		Memory: NewMemory(memSize),
	}
}

func (m *Machine) Step() error {
	return m.CPU.Step(m.Memory)
}

func (m *Machine) Reset() error {
	m.CPU = cpu.NewCPU()
	m.Memory = NewMemory(len(m.Memory.Data))
	return nil
}

// WriteProgramWords writes a slice of instructions (uint32) into memory at startAddr.
func (m *Machine) WriteProgramWords(prog []assembler.Instruction, startAddr uint32) error {
	for i, instr := range prog {
		fmt.Printf("WriteProgramWords: Instr %d @ 0x%08x: 0x%08x\n", i, startAddr+uint32(i*4), uint32(instr))
		if err := m.Memory.WriteWord(startAddr+uint32(i*4), uint32(instr)); err != nil {
			return err
		}
	}
	return nil
}

// LoadProgram writes the instructions and sets the PC to startAddr.
func (m *Machine) LoadProgram(prog []assembler.Instruction, startAddr uint32) error {
	if err := m.WriteProgramWords(prog, startAddr); err != nil {
		return err
	}
	m.CPU.PC = startAddr
	return nil
}
