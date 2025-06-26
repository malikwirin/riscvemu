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
	Mem   map[uint32]uint32
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
	assert.Equal(t, 32, len(cpu.Reg), "CPU should have 32 registers")
	for i := range cpu.Reg {
		assert.Equal(t, uint32(0), cpu.Reg[i], "Register x%d should be initialized to 0", i)
	}
	cpu.SetReg(0, 1234)
	assert.Equal(t, uint32(0), cpu.Reg[0], "Register x0 must always be 0")
}

func TestCPUStepErrors(t *testing.T) {
	cases := []struct {
		name      string
		setup     func() *CPU
		mem       *MockWordHandler
		expectErr bool
		wantPC    uint32
	}{
		{
			name:  "NOP advances PC",
			setup: func() *CPU { return NewCPU() },
			mem: func() *MockWordHandler {
				instr, _ := assembler.ParseInstruction("addi x0, x0, 0")
				return &MockWordHandler{Instr: uint32(instr)}
			}(),
			expectErr: false,
			wantPC:    INSTRUCTION_SIZE,
		},
		{
			name:      "Unknown opcode returns error",
			setup:     func() *CPU { return NewCPU() },
			mem:       &MockWordHandler{Instr: 0xFF},
			expectErr: true,
			wantPC:    0,
		},
		{
			name:      "Memory read failure returns error",
			setup:     func() *CPU { return NewCPU() },
			mem:       &MockWordHandler{Err: errors.New("out of bounds")},
			expectErr: true,
			wantPC:    0,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cpu := tc.setup()
			err := cpu.Step(tc.mem)
			if tc.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tc.wantPC, cpu.PC)
		})
	}
}

func TestCPU_Opcode_ALU_Branch_Jump_Memory(t *testing.T) {
	type regSetup func(cpu *CPU)
	tests := []struct {
		name    string
		asm     string
		setup   regSetup
		pc      uint32
		mem     map[uint32]uint32
		wantReg map[int]uint32
		wantPC  uint32
	}{
		{"ADDI", "addi x2, x1, 5", func(c *CPU) { c.Reg[1] = 10 }, 0, nil, map[int]uint32{2: 15}, 4},
		{"ADD", "add x5, x3, x4", func(c *CPU) { c.Reg[3], c.Reg[4] = 7, 5 }, 0, nil, map[int]uint32{5: 12}, 4},
		{"SUB", "sub x8, x6, x7", func(c *CPU) { c.Reg[6], c.Reg[7] = 20, 8 }, 0, nil, map[int]uint32{8: 12}, 4},
		{"SLT", "slt x12, x10, x11", func(c *CPU) { c.Reg[10], c.Reg[11] = 3, 7 }, 0, nil, map[int]uint32{12: 1}, 4},
		{"SLLI", "slli x5, x2, 3", func(c *CPU) { c.Reg[2] = 10 }, 0, nil, map[int]uint32{5: 80}, 4},
		{"BEQ taken", "beq x1, x2, 8", func(c *CPU) { c.Reg[1], c.Reg[2] = 5, 5 }, 100, nil, nil, 108},
		{"BEQ not taken", "beq x1, x2, 8", func(c *CPU) { c.Reg[1], c.Reg[2] = 5, 7 }, 100, nil, nil, 104},
		{"BNE taken", "bne x1, x2, 12", func(c *CPU) { c.Reg[1], c.Reg[2] = 5, 9 }, 100, nil, nil, 112},
		{"BNE not taken", "bne x1, x2, 12", func(c *CPU) { c.Reg[1], c.Reg[2] = 9, 9 }, 100, nil, nil, 104},
		{"JAL", "jal x5, 12", nil, 200, nil, map[int]uint32{5: 204}, 212},
		{"JALR", "jalr x6, 4(x2)", func(c *CPU) { c.Reg[2] = 500 }, 100, nil, map[int]uint32{6: 104}, 504},
		{"LW", "lw x3, 0(x2)", func(c *CPU) { c.Reg[2] = 100 }, 0, map[uint32]uint32{100: 0xDEADBEEF}, map[int]uint32{3: 0xDEADBEEF}, 4},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cpu := NewCPU()
			if tc.setup != nil {
				tc.setup(cpu)
			}
			cpu.PC = tc.pc
			instr, _ := assembler.ParseInstruction(tc.asm)
			mem := &MockWordHandler{Instr: uint32(instr), Mem: tc.mem}
			_ = cpu.Step(mem)
			for reg, want := range tc.wantReg {
				assert.Equalf(t, want, cpu.Reg[reg], "Reg x%d", reg)
			}
			assert.Equal(t, tc.wantPC, cpu.PC)
		})
	}
}

func TestAssemblerEncodings(t *testing.T) {
	type encTest struct {
		asm      string
		notZero  bool
		notASCII bool
	}
	tests := []encTest{
		{"lw x3, 0(x2)", true, true},
		{"sw x5, 0(x1)", true, true},
	}
	for _, tc := range tests {
		instr, err := assembler.ParseInstruction(tc.asm)
		assert.NoError(t, err)
		if tc.notZero {
			assert.NotEqual(t, uint32(0), uint32(instr))
		}
		if tc.notASCII {
			assert.NotEqual(t, uint32(0x4C4F4144), uint32(instr)) // "LOAD"
			assert.NotEqual(t, uint32(0x53544F52), uint32(instr)) // "STOR"
		}
	}
}

func TestCPU_InvalidJump(t *testing.T) {
	cpu := NewCPU()
	memory := NewMemory(64)
	cpu.PC = 0x1000
	_, err := memory.ReadWord(cpu.PC)
	assert.Error(t, err)
}

func TestCPU_Step_ReadsCorrectInstruction(t *testing.T) {
	mem := NewMemory(32)
	cpu := NewCPU()
	instr := assembler.Instruction(0x00112023)
	assert.NoError(t, mem.WriteWord(0, uint32(instr)))
	cpu.PC = 0
	assert.NoError(t, cpu.Step(mem))
	assert.Equal(t, uint32(4), cpu.PC)
}
