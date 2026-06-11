package cpu

import (
	"testing"

	"github.com/malikwirin/riscvemu/assembler"
)

// Test-only helpers shared by the arch/cpu test files. They live in the
// cpu package itself (not in cputest) so they can construct and
// manipulate a CPU directly without importing cpu back from cputest,
// which would be a cycle. The cputest package remains the home for
// pure-data helpers (MockWordHandler, Encode).

// makeCPU returns a CPU with the default config and the supplied
// functional-unit tweaks applied. Tests that need a non-default
// latency or RS count pass them as functional options; the default
// build keeps the helper one-liner for the common case.
func makeCPU(t *testing.T, opts ...func(*Config)) *CPU {
	t.Helper()
	cfg := DefaultConfig()
	for _, opt := range opts {
		opt(&cfg)
	}
	return NewCPU(cfg)
}

// withALULatency returns a config option that overrides the ALU latency.
func withALULatency(n int) func(*Config) { return func(c *Config) { c.ALULatency = n } }

// withLoadLatency returns a config option that overrides the load latency.
func withLoadLatency(n int) func(*Config) { return func(c *Config) { c.LoadLatency = n } }

// withStoreLatency returns a config option that overrides the store latency.
func withStoreLatency(n int) func(*Config) { return func(c *Config) { c.StoreLatency = n } }

// withALURSCount returns a config option that overrides the ALU RS count.
func withALURSCount(n int) func(*Config) { return func(c *Config) { c.ALURSCount = n } }

// withLSURSCount returns a config option that overrides the LSU RS count.
func withLSURSCount(n int) func(*Config) { return func(c *Config) { c.LSURSCount = n } }

// fetchProgram encodes each line of asm and enqueues it on the CPU at
// the matching PC, returning the last PC that was fetched so the caller
// can know where the program ends. It fails the test on parse errors
// or queue rejection.
func fetchProgram(t *testing.T, c *CPU, lines ...string) uint32 {
	t.Helper()
	var pc uint32
	for i, line := range lines {
		instr, err := assembler.ParseInstruction(line)
		if err != nil {
			t.Fatalf("ParseInstruction(%q): %v", line, err)
		}
		pc = uint32(i * 4)
		if !c.Fetch(uint32(instr), pc) {
			t.Fatalf("Fetch rejected %q at PC=0x%x", line, pc)
		}
	}
	return pc + 4
}

// runUntilBranch steps the CPU until LastBranch reports IsBranch=true
// or max cycles have been spent. It fails the test if no branch retires.
func runUntilBranch(t *testing.T, c *CPU, max int) {
	t.Helper()
	for i := 0; i < max; i++ {
		c.RunCycle()
		if c.LastBranch().IsBranch {
			return
		}
	}
	t.Fatalf("no branch retired within %d cycles", max)
}

// runUntilSettled steps the CPU until predicate returns true or max
// cycles have been spent. Use this for tests that wait on a register
// value or a memory write rather than a branch. If predicate is nil
// the helper just runs max cycles.
func runUntilSettled(t *testing.T, c *CPU, max int, predicate func() bool) {
	t.Helper()
	if predicate == nil {
		for i := 0; i < max; i++ {
			c.RunCycle()
		}
		return
	}
	for i := 0; i < max; i++ {
		c.RunCycle()
		if predicate() {
			return
		}
	}
	t.Fatalf("predicate stayed false for %d cycles", max)
}

// countBusyALUs returns the number of ALUs currently in the busy state.
func countBusyALUs(c *CPU) int {
	n := 0
	for _, a := range c.alus {
		if a.IsBusy() {
			n++
		}
	}
	return n
}

// countBusyLSUs returns the number of LSUs currently in the busy state.
func countBusyLSUs(c *CPU) int {
	n := 0
	for _, l := range c.lsus {
		if l.IsBusy() {
			n++
		}
	}
	return n
}
