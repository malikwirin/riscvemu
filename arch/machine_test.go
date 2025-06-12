package arch

import (
	"testing"

	"github.com/malikwirin/riscvemu/assembler"
)

func TestMachineInitialization(t *testing.T) {
	m := NewMachine(1024)
	if m.CPU == nil {
		t.Fatal("CPU should not be nil after initialization")
	}
	if m.Memory == nil {
		t.Fatal("Memory should not be nil after initialization")
	}
	if len(m.Memory.Data) != 1024 {
		t.Errorf("Expected memory size 1024, got %d", len(m.Memory.Data))
	}
	if m.CPU.PC != 0 {
		t.Errorf("Expected PC to be 0, got %d", m.CPU.PC)
	}
	for i, reg := range m.CPU.Reg {
		if reg != 0 {
			t.Errorf("Expected register x%d to be 0, got %d", i, reg)
		}
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
	if err != nil {
		t.Fatalf("Step returned error: %v", err)
	}
	if m.CPU.PC != oldPC+4 {
		t.Errorf("Expected PC to increase by 4, got %d -> %d", oldPC, m.CPU.PC)
	}
}

func TestMachineReset(t *testing.T) {
	m := NewMachine(128)
	m.CPU.PC = 100
	m.CPU.Reg[5] = 42
	m.Memory.Data[10] = 0xFF
	m.Reset()
	if m.CPU.PC != 0 {
		t.Errorf("After reset, expected PC to be 0, got %d", m.CPU.PC)
	}
	for i, reg := range m.CPU.Reg {
		if reg != 0 {
			t.Errorf("After reset, expected register x%d to be 0, got %d", i, reg)
		}
	}
	for i, b := range m.Memory.Data {
		if b != 0 {
			t.Errorf("After reset, expected memory at %d to be 0, got %d", i, b)
		}
	}
}

func TestMachineLoadProgram(t *testing.T) {
	m := NewMachine(64)
	program := []assembler.Instruction{0xDEADBEEF, 0x12345678, 0xCAFEBABE}
	startAddr := uint32(8)

	err := m.LoadProgram(program, startAddr)
	if err != nil {
		t.Fatalf("LoadProgram returned error: %v", err)
	}

	// Check if instructions are loaded at the correct addresses
	for i, want := range program {
		addr := startAddr + uint32(i*4)
		got, err := m.Memory.ReadWord(addr)
		if err != nil {
			t.Fatalf("ReadWord failed at addr %d: %v", addr, err)
		}
		if got != uint32(want) {
			t.Errorf("Instruction at %d: got 0x%X, want 0x%X", addr, got, want)
		}
	}

	if m.CPU.PC != startAddr {
		t.Errorf("Expected PC to be set to %d after LoadProgram, got %d", startAddr, m.CPU.PC)
	}
}
