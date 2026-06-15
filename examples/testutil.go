package examples

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/malikwirin/riscvemu/arch"
	"github.com/malikwirin/riscvemu/arch/cpu"
	"github.com/malikwirin/riscvemu/trace"
)

// Test helpers shared across the examples package's *_test.go
// files. They live in production code (not _test.go suffix) so
// that downstream test suites can reuse them when running
// Spec-validation traces against their own machine configuration.
// The helpers stay small on purpose: each captures one setup
// step that appeared in three or more test functions before the
// consolidation.

// NewSpecMachine builds a 1024-byte machine configured with the
// Spec-mandated 8-register layout and pipeline widths. Tests
// that need a non-default config should call
// arch.NewMachineWithConfig directly and then attach the memory
// themselves.
func NewSpecMachine() *arch.Machine {
	return newMachineWithConfig(cpu.SpecConfig())
}

// newMachineWithConfig is the under-the-hood builder. Tests
// that only need SpecConfig use NewSpecMachine; tests that
// override one or two fields use newMachineWithConfig directly
// to keep the boilerplate in one place.
func newMachineWithConfig(cfg cpu.Config) *arch.Machine {
	m := arch.NewMachineWithConfig(1024, cfg)
	m.CPU.AttachMemory(m.Memory)
	return m
}

// PreloadMem writes the given addresses and values into the
// machine's memory. The architectural register R0 stays zero
// (it cannot be written) and the other registers remain at
// their reset value; tests that need a non-zero starting value
// for a register should call CPU.SetReg directly.
func PreloadMem(t *testing.T, machine *arch.Machine, mem map[uint32]uint32) {
	t.Helper()
	machine.CPU.AttachMemory(machine.Memory)
	for addr, val := range mem {
		if err := machine.Memory.WriteWord(addr, val); err != nil {
			t.Fatalf("WriteWord %d: %v", addr, err)
		}
	}
}

// RunTrace reads a trace file from the examples/traces/ directory,
// drives it through the given machine, and returns the captured
// output. t.Fatal is called on read, parse, or run errors so
// the test fails fast at the point of misuse rather than later
// in an assert.
func RunTrace(t *testing.T, machine *arch.Machine, name string) *bytes.Buffer {
	t.Helper()
	src := ReadTraceFile(t, filepath.Join("traces", name))
	lines, err := trace.ParseTrace(src)
	if err != nil {
		t.Fatalf("ParseTrace: %v", err)
	}
	var buf bytes.Buffer
	d := trace.NewDriver(machine, &buf)
	if err := d.Run(lines); err != nil {
		t.Fatalf("Driver.Run: %v", err)
	}
	return &buf
}

// ReadTraceFile is the raw-text reader used by RunTrace and
// exposed for callers (notably experiment_test.go) that want the
// trace source string without driving a machine.
func ReadTraceFile(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile %q: %v", path, err)
	}
	return string(data)
}

// RunTraceDiscard drives a trace through the given machine and
// discards the driver's output. It exists for benchmark-style
// tests (such as TestExperiment) that only care about the
// resulting Statistics, not the captured text.
func RunTraceDiscard(t *testing.T, machine *arch.Machine, name string) {
	t.Helper()
	src := ReadTraceFile(t, filepath.Join("traces", name))
	lines, err := trace.ParseTrace(src)
	if err != nil {
		t.Fatalf("ParseTrace %q: %v", name, err)
	}
	d := trace.NewDriver(machine, io.Discard)
	if err := d.Run(lines); err != nil {
		t.Fatalf("Driver.Run %q: %v", name, err)
	}
}
