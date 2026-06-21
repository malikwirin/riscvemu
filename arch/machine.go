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

// Step advances the machine by one clock cycle: try to fetch the next
// instruction into the CPU's instruction queue, run one Tomasulo cycle
// (issue, execute, writeback), and then update the PC based on whether
// the CPU retired a branch or jump this cycle. If the instruction queue
// is full or the issue stage is gated on a pending branch, the PC is not
// advanced and the next Step will fetch the same word again.
func (m *Machine) Step() error {
	// Do not fetch while the issue stage is gated on an unresolved
	// branch. The PC will stay put this cycle and the same word will be
	// re-fetched after the branch retires, which keeps the instruction
	// queue free of duplicates.
	if m.CPU.HasUnresolvedBranch() {
		m.CPU.RunCycle()
		if br := m.CPU.LastBranch(); br.IsBranch && br.Taken {
			m.PC = br.Target
		}
		return nil
	}
	word, err := m.Memory.ReadWord(m.PC)
	if err != nil {
		return fmt.Errorf("fetch at PC=0x%08x: %w", m.PC, err)
	}
	accepted := m.CPU.Fetch(word, m.PC)
	// After Fetch, the IQ may already have a stalled issue pending. We
	// detect this by checking whether there is an unresolved branch.
	m.CPU.RunCycle()
	if br := m.CPU.LastBranch(); br.IsBranch && br.Taken {
		m.PC = br.Target
	} else if accepted && !m.CPU.IQFull() && !m.CPU.HasUnresolvedBranch() {
		m.PC += 4
	}
	// If the IQ was full, the PC stays at the current value and the same
	// word is fetched again on the next Step.
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
