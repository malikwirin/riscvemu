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

	t.Run("Assembler produces valid encoding for lw", func(t *testing.T) {
		instr, err := assembler.ParseInstruction("lw x3, 0(x2)")
		if err != nil {
			t.Fatalf("ParseInstruction failed for lw: %v", err)
		}
		// Print the encoded instruction for debugging
		t.Logf("lw encoded as: 0x%08x", instr)
		// Check for suspicious values (e.g., ASCII "LOAD" or 0)
		if instr == 0 || instr == 0x4C4F4144 { // "LOAD"
			t.Errorf("Assembler returned suspicious instruction for lw: 0x%08x", instr)
		}
	})

	t.Run("Assembler produces valid encoding for sw", func(t *testing.T) {
		instr, err := assembler.ParseInstruction("sw x5, 0(x1)")
		if err != nil {
			t.Fatalf("ParseInstruction failed for sw: %v", err)
		}
		// Print the encoded instruction for debugging
		t.Logf("sw encoded as: 0x%08x", instr)
		// Check for suspicious values (e.g., ASCII "STOR" or 0)
		if instr == 0 || instr == 0x53544F52 { // "STOR"
			t.Errorf("Assembler returned suspicious instruction for sw: 0x%08x", instr)
		}
	})

	t.Run("CPU returns error for unimplemented lw", func(t *testing.T) {
		cpu := NewCPU()
		instr, _ := assembler.ParseInstruction("lw x3, 0(x2)")
		mem := &MockWordReader{Instr: uint32(instr)}
		err := cpu.Step(mem)
		if err == nil {
			t.Error("Expected error for unimplemented lw, got nil")
		} else {
			t.Logf("CPU error for lw: %v", err)
		}
	})

	t.Run("CPU returns error for unimplemented sw", func(t *testing.T) {
		cpu := NewCPU()
		instr, _ := assembler.ParseInstruction("sw x5, 0(x1)")
		mem := &MockWordReader{Instr: uint32(instr)}
		err := cpu.Step(mem)
		if err == nil {
			t.Error("Expected error for unimplemented sw, got nil")
		} else {
			t.Logf("CPU error for sw: %v", err)
		}
	})
}

func TestCPU_Integration_Example2(t *testing.T) {
	// assemble the instructions from examples/2.asm
	asm := []string{
		"addi x1, x0, 42",    // x1 = 42
		"addi x2, x0, 100",   // x2 = 100 (Adresse)
		"sw x1, 0(x2)",       // Speicher[100] = 42
		"lw x3, 0(x2)",       // x3 = Speicher[100] -> 42
	}
	var program []assembler.Instruction
	for i, line := range asm {
		instr, err := assembler.ParseInstruction(line)
		if err != nil {
			t.Fatalf("ParseInstruction failed at line %d: %q: %v", i+1, line, err)
		}
		program = append(program, instr)
	}

	// Set up a Machine, load program at address 0
	m := NewMachine(256) // 256 Bytes Memory
	err := m.LoadProgram(program, 0)
	if err != nil {
		t.Fatalf("LoadProgram failed: %v", err)
	}

	// Step 1: addi x1, x0, 42
	if err := m.Step(); err != nil {
		t.Fatalf("Step 1 failed: %v", err)
	}
	if got := m.CPU.Reg[1]; got != 42 {
		t.Errorf("After Step 1: x1 = %d, want 42", got)
	}

	// Step 2: addi x2, x0, 100
	if err := m.Step(); err != nil {
		t.Fatalf("Step 2 failed: %v", err)
	}
	if got := m.CPU.Reg[2]; got != 100 {
		t.Errorf("After Step 2: x2 = %d, want 100", got)
	}

	// Step 3: sw x1, 0(x2)
	if err := m.Step(); err != nil {
		t.Fatalf("Step 3 failed: %v", err)
	}
	w, err := m.Memory.ReadWord(100)
	if err != nil {
		t.Fatalf("Memory.ReadWord(100) failed: %v", err)
	}
	if w != 42 {
		t.Errorf("After Step 3: Memory[100] = %d, want 42", w)
	}

	// Step 4: lw x3, 0(x2)
	if err := m.Step(); err != nil {
		t.Fatalf("Step 4 failed: %v", err)
	}
	if got := m.CPU.Reg[3]; got != 42 {
		t.Errorf("After Step 4: x3 = %d, want 42", got)
	}
}
