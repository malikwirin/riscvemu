// Package cputest provides test-only helpers shared by the cpu package
// tests: a small in-memory WordHandler and a one-liner that turns an
// assembly mnemonic into the encoded instruction word.
//
// The package deliberately does not import the cpu package: the test
// files in arch/cpu import cputest, so a back-import would be a cycle
// that Go's test build rejects. MockWordHandler therefore mirrors the
// cpu.WordHandler interface by signature; if the interface ever drifts
// the structural assignment in cpu.AttachMemory will fail to compile.
package cputest

import (
	"testing"

	"codeberg.org/malik/riscvemu/assembler"
)

// MockWordHandler is a simple in-memory backend for CPU tests. It
// satisfies cpu.WordHandler by signature.
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

// Encode converts an assembly line to a uint32 instruction word. It
// fails the test if parsing fails.
func Encode(t *testing.T, line string) uint32 {
	t.Helper()
	instr, err := assembler.ParseInstruction(line)
	if err != nil {
		t.Fatalf("ParseInstruction(%q): %v", line, err)
	}
	return uint32(instr)
}
