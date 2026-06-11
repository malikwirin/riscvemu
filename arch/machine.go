package arch

import (
	"fmt"

	"github.com/malikwirin/riscvemu/arch/cpu"
	"github.com/malikwirin/riscvemu/assembler"
)

// Machine owns the program counter and feeds instructions to the Tomasulo CPU.
type Machine struct {
	CPU    *cpu.CPU
	Memory *Memory
	PC     uint32
	cfg    cpu.Config
}

func NewMachine(memSize int) *Machine {
	return NewMachineWithConfig(memSize, cpu.DefaultConfig())
}

func NewMachineWithConfig(memSize int, cfg cpu.Config) *Machine {
	mem := NewMemory(memSize)
	core := cpu.NewCPU(cfg)
	core.AttachMemory(mem)
	return &Machine{
		CPU:    core,
		Memory: mem,
		PC:     0,
		cfg:    cfg,
	}
}

// Step advances the machine by one clock cycle: fetch the next instruction,
// feed it to the CPU, run one Tomasulo cycle, and advance the PC.
func (m *Machine) Step() error {
	word, err := m.Memory.ReadWord(m.PC)
	if err != nil {
		return fmt.Errorf("fetch at PC=0x%08x: %w", m.PC, err)
	}
    if err := m.CPU.ReceiveInstruction(word, m.PC); err != nil {
		return err
	}
	m.CPU.RunCycle()
	m.PC += 4
	return nil
}

// Reset re-initializes the machine to a fresh CPU and memory.
func (m *Machine) Reset() error {
	m.Memory = NewMemory(len(m.Memory.Data))
	m.CPU = cpu.NewCPU(m.cfg)
	m.CPU.AttachMemory(m.Memory)
	m.PC = 0
	return nil
}

// WriteProgramWords writes a slice of instructions (uint32) into memory at startAddr.
func (m *Machine) WriteProgramWords(prog []assembler.Instruction, startAddr uint32) error {
	for i, instr := range prog {
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
	m.PC = startAddr
	return nil
}
