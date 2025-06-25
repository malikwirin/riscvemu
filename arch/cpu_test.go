package arch

import (
	"errors"
	"testing"

	"github.com/malikwirin/riscvemu/assembler"
	"github.com/stretchr/testify/assert"
)

type MockWordHandler struct {
	Instr uint32
	Err   error
	Mem   map[uint32]uint32 // address -> value
}

func (m *MockWordHandler) ReadWord(addr uint32) (uint32, error) {
	if m.Err != nil {
		return 0, m.Err
	}
	if m.Mem != nil {
		if val, ok := m.Mem[addr]; ok {
			return val, nil
		}
	}
	return m.Instr, nil
}

func (m *MockWordHandler) WriteWord(addr uint32, value uint32) error {
	if m.Err != nil {
		return m.Err
	}
	if m.Mem != nil {
		m.Mem[addr] = value
	}
	return nil
}

func TestCPURegisters(t *testing.T) {
	cpu := NewCPU()

	t.Run("CPU has 32 registers", func(t *testing.T) {
		assert.Equal(t, 32, len(cpu.Reg), "CPU should have 32 registers")
	})

	t.Run("Registers are zero initialized", func(t *testing.T) {
		for i := 0; i < 32; i++ {
			assert.Equal(t, uint32(0), cpu.Reg[i], "Register x%d should be initialized to 0", i)
		}
	})

	t.Run("Register x0 is always zero", func(t *testing.T) {
		cpu.SetReg(0, 1234)
		assert.Equal(t, uint32(0), cpu.Reg[0], "Register x0 must always be 0")
	})
}

func TestCPUStep(t *testing.T) {
	t.Run("Step executes NOP and advances PC", func(t *testing.T) {
		cpu := NewCPU()
		instr, _ := assembler.ParseInstruction("addi x0, x0, 0")
		mem := &MockWordHandler{Instr: uint32(instr)}

		err := cpu.Step(mem)
		assert.NoError(t, err)
		assert.Equal(t, uint32(INSTRUCTION_SIZE), cpu.PC, "PC should be 4 after step")
	})

	t.Run("Step returns error for unknown opcode", func(t *testing.T) {
		cpu := NewCPU()
		mem := &MockWordHandler{Instr: 0xFF}
		err := cpu.Step(mem)
		assert.Error(t, err, "Expected error for unknown opcode")
		assert.Equal(t, uint32(0), cpu.PC, "PC should not advance on error")
	})

	t.Run("Step returns error on memory read failure", func(t *testing.T) {
		cpu := NewCPU()
		mem := &MockWordHandler{Err: errors.New("out of bounds")}
		err := cpu.Step(mem)
		assert.Error(t, err, "Expected error from memory read")
		assert.Equal(t, uint32(0), cpu.PC, "PC should not advance on memory error")
	})
}

func TestCPUOpcodes(t *testing.T) {
	t.Run("ADDI: addi x2, x1, 5 -> x2 == 15", func(t *testing.T) {
		cpu := NewCPU()
		cpu.Reg[1] = 10
		instr, _ := assembler.ParseInstruction("addi x2, x1, 5")
		mem := &MockWordHandler{Instr: uint32(instr)}
		_ = cpu.Step(mem)
		assert.Equal(t, uint32(15), cpu.Reg[2], "Expected x2 = 15")
	})

	t.Run("ADD: add x5, x3, x4 -> x5 == 12", func(t *testing.T) {
		cpu := NewCPU()
		cpu.Reg[3] = 7
		cpu.Reg[4] = 5
		instr, _ := assembler.ParseInstruction("add x5, x3, x4")
		mem := &MockWordHandler{Instr: uint32(instr)}
		_ = cpu.Step(mem)
		assert.Equal(t, uint32(12), cpu.Reg[5], "Expected x5 = 12")
	})

	t.Run("SUB: sub x8, x6, x7 -> x8 == 12", func(t *testing.T) {
		cpu := NewCPU()
		cpu.Reg[6] = 20
		cpu.Reg[7] = 8
		instr, _ := assembler.ParseInstruction("sub x8, x6, x7")
		mem := &MockWordHandler{Instr: uint32(instr)}
		_ = cpu.Step(mem)
		assert.Equal(t, uint32(12), cpu.Reg[8], "Expected x8 = 12")
	})

	t.Run("SLT: slt x12, x10, x11 -> x12 == 1", func(t *testing.T) {
		cpu := NewCPU()
		cpu.Reg[10] = 3
		cpu.Reg[11] = 7
		instr, _ := assembler.ParseInstruction("slt x12, x10, x11")
		mem := &MockWordHandler{Instr: uint32(instr)}
		_ = cpu.Step(mem)
		assert.Equal(t, uint32(1), cpu.Reg[12], "Expected x12 = 1")
	})

	t.Run("SLLI: slli x5, x2, 3 -> x5 == 80", func(t *testing.T) {
		cpu := NewCPU()
		cpu.Reg[2] = 10
		instr, _ := assembler.ParseInstruction("slli x5, x2, 3")
		mem := &MockWordHandler{Instr: uint32(instr)}
		_ = cpu.Step(mem)
		assert.Equal(t, uint32(80), cpu.Reg[5], "Expected x5 = 80")
	})

	t.Run("BEQ: beq x1, x2, 8 (taken)", func(t *testing.T) {
		cpu := NewCPU()
		cpu.Reg[1] = 5
		cpu.Reg[2] = 5
		cpu.PC = 100
		instr, _ := assembler.ParseInstruction("beq x1, x2, 8")
		mem := &MockWordHandler{Instr: uint32(instr)}
		_ = cpu.Step(mem)
		assert.Equal(t, uint32(108), cpu.PC, "Expected PC = 108")
	})

	t.Run("BEQ: beq x1, x2, 8 (not taken)", func(t *testing.T) {
		cpu := NewCPU()
		cpu.Reg[1] = 5
		cpu.Reg[2] = 7
		cpu.PC = 100
		instr, _ := assembler.ParseInstruction("beq x1, x2, 8")
		mem := &MockWordHandler{Instr: uint32(instr)}
		_ = cpu.Step(mem)
		assert.Equal(t, uint32(104), cpu.PC, "Expected PC = 104")
	})

	t.Run("BNE: bne x1, x2, 12 (taken)", func(t *testing.T) {
		cpu := NewCPU()
		cpu.Reg[1] = 5
		cpu.Reg[2] = 9
		cpu.PC = 100
		instr, _ := assembler.ParseInstruction("bne x1, x2, 12")
		mem := &MockWordHandler{Instr: uint32(instr)}
		_ = cpu.Step(mem)
		assert.Equal(t, uint32(112), cpu.PC, "Expected PC = 112")
	})

	t.Run("BNE: bne x1, x2, 12 (not taken)", func(t *testing.T) {
		cpu := NewCPU()
		cpu.Reg[1] = 9
		cpu.Reg[2] = 9
		cpu.PC = 100
		instr, _ := assembler.ParseInstruction("bne x1, x2, 12")
		mem := &MockWordHandler{Instr: uint32(instr)}
		_ = cpu.Step(mem)
		assert.Equal(t, uint32(104), cpu.PC, "Expected PC = 104")
	})

	t.Run("JAL: jal x5, 12", func(t *testing.T) {
		cpu := NewCPU()
		cpu.PC = 200
		instr, _ := assembler.ParseInstruction("jal x5, 12")
		mem := &MockWordHandler{Instr: uint32(instr)}
		_ = cpu.Step(mem)
		assert.Equal(t, uint32(204), cpu.Reg[5], "Expected x5 = 204")
		assert.Equal(t, uint32(212), cpu.PC, "Expected PC = 212")
	})

	t.Run("JALR: jalr x6, 4(x2)", func(t *testing.T) {
		cpu := NewCPU()
		cpu.PC = 100
		cpu.Reg[2] = 500
		instr, _ := assembler.ParseInstruction("jalr x6, 4(x2)")
		mem := &MockWordHandler{Instr: uint32(instr)}
		_ = cpu.Step(mem)
		assert.Equal(t, uint32(104), cpu.Reg[6], "Expected x6 = 104")
		assert.Equal(t, uint32(504), cpu.PC, "Expected PC = 504")
	})

	t.Run("Assembler produces valid encoding for lw", func(t *testing.T) {
		instr, err := assembler.ParseInstruction("lw x3, 0(x2)")
		assert.NoError(t, err, "ParseInstruction failed for lw")
		t.Logf("lw encoded as: 0x%08x", instr)
		assert.NotEqual(t, uint32(0), uint32(instr), "Assembler returned suspicious instruction for lw (0)")
		assert.NotEqual(t, uint32(0x4C4F4144), uint32(instr), "Assembler returned suspicious instruction for lw (ASCII 'LOAD')")
	})

	t.Run("Assembler produces valid encoding for sw", func(t *testing.T) {
		instr, err := assembler.ParseInstruction("sw x5, 0(x1)")
		assert.NoError(t, err, "ParseInstruction failed for sw")
		t.Logf("sw encoded as: 0x%08x", instr)
		assert.NotEqual(t, uint32(0), uint32(instr), "Assembler returned suspicious instruction for sw (0)")
		assert.NotEqual(t, uint32(0x53544F52), uint32(instr), "Assembler returned suspicious instruction for sw (ASCII 'STOR')")
	})

	t.Run("LW: lw x3, 0(x2) loads value from memory", func(t *testing.T) {
		cpu := NewCPU()
		cpu.Reg[2] = 100

		instr, _ := assembler.ParseInstruction("lw x3, 0(x2)")

		mem := &MockWordHandler{
			Instr: uint32(instr),
			Mem: map[uint32]uint32{
				100: 0xDEADBEEF,
			},
		}

		err := cpu.Step(mem)
		assert.NoError(t, err, "Step failed")
		assert.Equal(t, uint32(0xDEADBEEF), cpu.Reg[3], "Expected x3 = 0xDEADBEEF")
	})
}

func TestCPU_Integration_Example2(t *testing.T) {
	asm := []string{
		"addi x1, x0, 42",
		"addi x2, x0, 100",
		"sw x1, 0(x2)",
		"lw x3, 0(x2)",
	}
	var program []assembler.Instruction
	for i, line := range asm {
		instr, err := assembler.ParseInstruction(line)
		assert.NoErrorf(t, err, "ParseInstruction failed at line %d: %q", i+1, line)
		program = append(program, instr)
	}

	m := NewMachine(256)
	err := m.LoadProgram(program, 0)
	assert.NoError(t, err, "LoadProgram failed")

	assert.NoError(t, m.Step(), "Step 1 failed")
	assert.Equal(t, uint32(42), m.CPU.Reg[1], "After Step 1: x1")

	assert.NoError(t, m.Step(), "Step 2 failed")
	assert.Equal(t, uint32(100), m.CPU.Reg[2], "After Step 2: x2")

	assert.NoError(t, m.Step(), "Step 3 failed")
	w, err := m.Memory.ReadWord(100)
	assert.NoError(t, err, "Memory.ReadWord(100) failed")
	assert.Equal(t, uint32(42), w, "After Step 3: Memory[100]")

	assert.NoError(t, m.Step(), "Step 4 failed")
	assert.Equal(t, uint32(42), m.CPU.Reg[3], "After Step 4: x3")
}

func TestCPU_InvalidJump(t *testing.T) {
	cpu := NewCPU()
	memory := NewMemory(64)
	cpu.PC = 0x1000
	_, err := memory.ReadWord(cpu.PC)
	assert.Error(t, err, "Invalid jump to out-of-bounds address was not detected!")
}

func TestCPU_Step_ReadsCorrectInstruction(t *testing.T) {
	mem := NewMemory(32)
	cpu := NewCPU()
	instr := assembler.Instruction(0x00112023)
	err := mem.WriteWord(0, uint32(instr))
	assert.NoError(t, err, "WriteWord failed")

	cpu.PC = 0

	err = cpu.Step(mem)
	assert.NoError(t, err, "CPU step failed unexpectedly")
	assert.Equal(t, uint32(4), cpu.PC, "PC not incremented correctly")
}
