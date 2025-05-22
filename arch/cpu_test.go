package arch

import (
	"testing"
	"errors"
)

type MockWordReader struct {
    Instr uint32
    Err   error
}

func (m *MockWordReader) ReadWord(addr uint32) (uint32, error) {
    return m.Instr, m.Err
}

func TestCPURegisters(t *testing.T) {
    cpu := NewCPU()

    t.Run("CPU has 32 registers", func(t *testing.T) {
        if len(cpu.Registers) != 32 {
            t.Errorf("CPU should have 32 registers, got %d", len(cpu.Registers))
        }
    })

    t.Run("Registers are zero initialized", func(t *testing.T) {
        for i := 0; i < 32; i++ {
            if cpu.Registers[i] != 0 {
                t.Errorf("Register x%d should be initialized to 0, got %d", i, cpu.Registers[i])
            }
        }
    })

    t.Run("Register x0 is always zero", func(t *testing.T) {
        cpu.SetRegister(0, 1234)
        if cpu.Registers[0] != 0 {
            t.Errorf("Register x0 must always be 0, got %d", cpu.Registers[0])
        }
    })
}

func TestCPUStep(t *testing.T) {
    t.Run("Step executes NOP and advances PC", func(t *testing.T) {
        cpu := NewCPU()
        mem := &MockWordReader{Instr: 0x00} // 0x00 as NOP

        err := cpu.Step(mem)
        if err != nil {
            t.Fatalf("Step returned unexpected error: %v", err)
        }
        if cpu.PC != INSTRUCTION_SIZE {
            t.Errorf("PC should be 4 after step, got %d", cpu.PC)
        }
    })

    t.Run("Step returns error for unknown opcode", func(t *testing.T) {
        cpu := NewCPU()
        mem := &MockWordReader{Instr: 0xFF}

        err := cpu.Step(mem)
        if err == nil {
            t.Fatal("Expected error for unknown opcode, got nil")
        }
        if cpu.PC != 0 {
            t.Errorf("PC should not advance on error, got %d", cpu.PC)
        }
    })

    t.Run("Step returns error on memory read failure", func(t *testing.T) {
        cpu := NewCPU()
        mem := &MockWordReader{Err: errors.New("out of bounds")}

        err := cpu.Step(mem)
        if err == nil {
            t.Fatal("Expected error from memory read, got nil")
        }
        if cpu.PC != 0 {
            t.Errorf("PC should not advance on memory error, got %d", cpu.PC)
        }
    })
}
