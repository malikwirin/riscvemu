package arch

import (
    "testing"
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
    for i, reg := range m.CPU.Registers {
        if reg != 0 {
            t.Errorf("Expected register x%d to be 0, got %d", i, reg)
        }
    }
}

func TestMachineStepIncreasesPC(t *testing.T) {
    m := NewMachine(2048)
    oldPC := m.CPU.PC
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
    m.CPU.Registers[5] = 42
    m.Memory.Data[10] = 0xFF
    m.Reset()
    if m.CPU.PC != 0 {
        t.Errorf("After reset, expected PC to be 0, got %d", m.CPU.PC)
    }
    for i, reg := range m.CPU.Registers {
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
