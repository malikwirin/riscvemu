package cli

import (
	"bytes"
	"os"

	"github.com/malikwirin/riscvemu/arch"
)

// testOwner is a test double for machineOwner
type testOwner struct {
	m *arch.Machine
}

func (t *testOwner) Machine() *arch.Machine { return t.m }

// captureOutput runs f and returns what is printed to os.Stdout as a string.
func captureOutput(f func()) string {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	f()

	w.Close()
	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	os.Stdout = old
	return buf.String()
}

// withMachine runs f with a new Machine of given size and a testOwner.
// Usage: withMachine(128, func(m *arch.Machine, owner *testOwner) { ... })
func withMachine(memSize int, f func(m *arch.Machine, owner *testOwner)) {
	m := arch.NewMachine(memSize)
	owner := &testOwner{m}
	f(m, owner)
}

// isRandomized returns true if the values slice is plausibly random (not all zero, not all the same).
func isRandomized(values []uint32) bool {
	if len(values) == 0 {
		return false
	}
	allZero := true
	first := values[0]
	allSame := true
	for _, v := range values {
		if v != 0 {
			allZero = false
		}
		if v != first {
			allSame = false
		}
	}
	return !allZero && !allSame
}
