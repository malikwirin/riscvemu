package cpu

import (
	"testing"

	"github.com/malikwirin/riscvemu/assembler"
)

// MockWordHandler is a simple in-memory backend for CPU tests.
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

// encode converts an assembly line to a uint32 instruction word.
func encode(t *testing.T, line string) uint32 {
	t.Helper()
	instr, err := assembler.ParseInstruction(line)
	if err != nil {
		t.Fatalf("ParseInstruction(%q): %v", line, err)
	}
	return uint32(instr)
}
