package arch

import (
	"errors"
	"testing"

	"github.com/malikwirin/riscvemu/assembler"
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
		if len(cpu.Reg) != 32 {
			t.Errorf("CPU should have 32 registers, got %d", len(cpu.Reg))
		}
	})

	t.Run("Registers are zero initialized", func(t *testing.T) {
		for i := 0; i < 32; i++ {
			if cpu.Reg[i] != 0 {
				t.Errorf("Register x%d should be initialized to 0, got %d", i, cpu.Reg[i])
			}
		}
	})

	t.Run("Register x0 is always zero", func(t *testing.T) {
		cpu.SetReg(0, 1234)
		if cpu.Reg[0] != 0 {
			t.Errorf("Register x0 must always be 0, got %d", cpu.Reg[0])
		}
	})
}

func TestCPUStep(t *testing.T) {
	t.Run("Step executes NOP and advances PC", func(t *testing.T) {
		cpu := NewCPU()
		instr, _ := assembler.ParseInstruction("addi x0, x0, 0") // echtes NOP!
		mem := &MockWordReader{Instr: uint32(instr)}

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

func TestCPUOpcodes(t *testing.T) {
	t.Run("ADDI: addi x2, x1, 5 -> x2 == 15", func(t *testing.T) {
		cpu := NewCPU()
		cpu.Reg[1] = 10
		instr, _ := assembler.ParseInstruction("addi x2, x1, 5")
		mem := &MockWordReader{Instr: uint32(instr)}
		_ = cpu.Step(mem)
		if cpu.Reg[2] != 15 {
			t.Errorf("Expected x2 = 15, got %d", cpu.Reg[2])
		}
	})

	t.Run("ADD: add x5, x3, x4 -> x5 == 12", func(t *testing.T) {
		cpu := NewCPU()
		cpu.Reg[3] = 7
		cpu.Reg[4] = 5
		instr, _ := assembler.ParseInstruction("add x5, x3, x4")
		mem := &MockWordReader{Instr: uint32(instr)}
		_ = cpu.Step(mem)
		if cpu.Reg[5] != 12 {
			t.Errorf("Expected x5 = 12, got %d", cpu.Reg[5])
		}
	})

	t.Run("SUB: sub x8, x6, x7 -> x8 == 12", func(t *testing.T) {
		cpu := NewCPU()
		cpu.Reg[6] = 20
		cpu.Reg[7] = 8
		instr, _ := assembler.ParseInstruction("sub x8, x6, x7")
		mem := &MockWordReader{Instr: uint32(instr)}
		_ = cpu.Step(mem)
		if cpu.Reg[8] != 12 {
			t.Errorf("Expected x8 = 12, got %d", cpu.Reg[8])
		}
	})

	t.Run("SLT: slt x12, x10, x11 -> x12 == 1", func(t *testing.T) {
		cpu := NewCPU()
		cpu.Reg[10] = 3
		cpu.Reg[11] = 7
		instr, _ := assembler.ParseInstruction("slt x12, x10, x11")
		mem := &MockWordReader{Instr: uint32(instr)}
		_ = cpu.Step(mem)
		if cpu.Reg[12] != 1 {
			t.Errorf("Expected x12 = 1, got %d", cpu.Reg[12])
		}
	})

	t.Run("SLLI: slli x5, x2, 3 -> x5 == 80", func(t *testing.T) {
		cpu := NewCPU()
		cpu.Reg[2] = 10
		instr, _ := assembler.ParseInstruction("slli x5, x2, 3") // 10 << 3 == 80
		mem := &MockWordReader{Instr: uint32(instr)}
		_ = cpu.Step(mem)
		if cpu.Reg[5] != 80 {
			t.Errorf("Expected x5 = 80, got %d", cpu.Reg[5])
		}
	})

	t.Run("BEQ: beq x1, x2, 8 (taken)", func(t *testing.T) {
		cpu := NewCPU()
		cpu.Reg[1] = 5
		cpu.Reg[2] = 5
		cpu.PC = 100
		instr, _ := assembler.ParseInstruction("beq x1, x2, 8")
		mem := &MockWordReader{Instr: uint32(instr)}
		_ = cpu.Step(mem)
		if cpu.PC != 108 {
			t.Errorf("Expected PC = 108, got %d", cpu.PC)
		}
	})

	t.Run("BEQ: beq x1, x2, 8 (not taken)", func(t *testing.T) {
		cpu := NewCPU()
		cpu.Reg[1] = 5
		cpu.Reg[2] = 7
		cpu.PC = 100
		instr, _ := assembler.ParseInstruction("beq x1, x2, 8")
		mem := &MockWordReader{Instr: uint32(instr)}
		_ = cpu.Step(mem)
		if cpu.PC != 104 {
			t.Errorf("Expected PC = 104, got %d", cpu.PC)
		}
	})

	t.Run("BNE: bne x1, x2, 12 (taken)", func(t *testing.T) {
		cpu := NewCPU()
		cpu.Reg[1] = 5
		cpu.Reg[2] = 9
		cpu.PC = 100
		instr, _ := assembler.ParseInstruction("bne x1, x2, 12")
		mem := &MockWordReader{Instr: uint32(instr)}
		_ = cpu.Step(mem)
		if cpu.PC != 112 {
			t.Errorf("Expected PC = 112, got %d", cpu.PC)
		}
	})

	t.Run("BNE: bne x1, x2, 12 (not taken)", func(t *testing.T) {
		cpu := NewCPU()
		cpu.Reg[1] = 9
		cpu.Reg[2] = 9
		cpu.PC = 100
		instr, _ := assembler.ParseInstruction("bne x1, x2, 12")
		mem := &MockWordReader{Instr: uint32(instr)}
		_ = cpu.Step(mem)
		if cpu.PC != 104 {
			t.Errorf("Expected PC = 104, got %d", cpu.PC)
		}
	})

	t.Run("JAL: jal x5, 12", func(t *testing.T) {
		cpu := NewCPU()
		cpu.PC = 200
		instr, _ := assembler.ParseInstruction("jal x5, 12")
		mem := &MockWordReader{Instr: uint32(instr)}
		_ = cpu.Step(mem)
		if cpu.Reg[5] != 204 {
			t.Errorf("Expected x5 = 204, got %d", cpu.Reg[5])
		}
		if cpu.PC != 212 {
			t.Errorf("Expected PC = 212, got %d", cpu.PC)
		}
	})

	t.Run("JALR: jalr x6, 4(x2)", func(t *testing.T) {
		cpu := NewCPU()
		cpu.PC = 100
		cpu.Reg[2] = 500
		instr, _ := assembler.ParseInstruction("jalr x6, 4(x2)")
		mem := &MockWordReader{Instr: uint32(instr)}
		_ = cpu.Step(mem)
		if cpu.Reg[6] != 104 {
			t.Errorf("Expected x6 = 104, got %d", cpu.Reg[6])
		}
		if cpu.PC != 504 {
			t.Errorf("Expected PC = 504, got %d", cpu.PC)
		}
	})

	t.Run("LW: lw x3, 0(x2) (not implemented yet, skipped)", func(t *testing.T) {
		t.Skip("Memory access for lw/sw not implemented in CPU")
	})

	t.Run("SW: sw x5, 0(x1) (not implemented yet, skipped)", func(t *testing.T) {
		t.Skip("Memory access for lw/sw not implemented in CPU")
	})
}
