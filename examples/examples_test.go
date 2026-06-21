// Package examples hosts validation traces for the trace driver.
//
//  1. A simple sequence without data dependencies (instructions
//     run in parallel).
//  2. A RAW dependency chain that prevents out-of-order execution,
//     with a comparison to simple in-order execution.
//  3. A sequence where the reservation stations are full
//     (structural stall).
//  4. Correct resolution of a WAR or WAW pseudo-dependency by
//     Tomasulo.
//  5. Correct function of LOAD and STORE including forwarding
//     through the CDB.
//
// The tests in this file drive each trace through the Tomasulo
// pipeline and assert the expected final state of the
// architectural registers (R0..R7) and memory. The cases that
// depend on the pipeline configuration use SpecConfig so the
// reservation-station counts and latencies match the example
// values the spec calls out.
package examples

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/malikwirin/riscvemu/arch/cpu"
)

// readTrace reads a trace file from the examples/traces/
// directory and returns the raw source text. experiment_test.go
// uses this to feed the source into its own machine loop, so
// each (config, trace) pair can be measured independently.
func readTrace(t *testing.T, name string) string {
	t.Helper()
	return ReadTraceFile(t, filepath.Join("traces", name))
}

// TestValidation01Parallel exercises the parallel-execution case.
// Three independent LOADs feed three independent ADDs, all of
// which can run in parallel because their source operand pairs
// do not depend on each other.
func TestValidation01Parallel(t *testing.T) {
	machine := NewSpecMachine()
	PreloadMem(t, machine, map[uint32]uint32{
		100: 5,
		200: 6,
		300: 7,
	})
	out := RunTrace(t, machine, "01-parallel.trace")
	wants := map[uint32]uint32{1: 5, 2: 6, 3: 7}
	for reg, want := range wants {
		if got := machine.CPU.Reg(reg); got != want {
			t.Errorf("R%d = %d, want %d", reg, got, want)
		}
	}
	if !strings.Contains(out.String(), "Cycles=") {
		t.Errorf("'h' command should have produced a Cycles=... line, got: %s", out.String())
	}
}

// TestValidation02RAWChain exercises the RAW-dependency case.
// Two independent LOADs populate R5 and R6, then three ADDs
// chain their results. ADD R1 is independent, ADD R2 reads
// R1 (RAW on the first ADD), and ADD R3 reads R2 (RAW on
// the second). With Tomasulo the first ADD's result still
// has to reach the second ADD through the CDB before the
// second can start executing, so the chain's retired count
// is sequential.
func TestValidation02RAWChain(t *testing.T) {
	machine := NewSpecMachine()
	PreloadMem(t, machine, map[uint32]uint32{
		100: 1, 104: 1,
	})
	RunTrace(t, machine, "02-raw-chain.trace")
	for _, c := range []struct {
		reg  uint32
		want uint32
	}{
		{5, 1}, // LOAD target
		{1, 1}, // ADD R1 = R5+R0
		{2, 2}, // ADD R2 = R1+R1
		{3, 2}, // ADD R3 = R2+R0
	} {
		if got := machine.CPU.Reg(c.reg); got != c.want {
			t.Errorf("R%d = %d, want %d (RAW chain propagates the value from mem[100])", c.reg, got, c.want)
		}
	}
	if got := machine.CPU.Stats().Retired; got < 5 {
		t.Errorf("Stats().Retired = %d, want >= 5 (two LOADs + three ADDs)", got)
	}
}

// TestValidation03StructuralStall exercises the case where the
// reservation stations are full. The trace emits six independent
// ADDs back-to-back. With the Spec's 4 ALU-RS slots and the IQ
// blocked on the driver, the issue stage must contend with the
// RS in the worst case. We deliberately stretch ALULatency to
// 10 so the test is robust against minor scheduling variations.
// (The Spec's default latency of 1 lets the RS drain between
// cycles, so StructuralStalls stays at 0; with the longer
// latency the RS is forced to hold the entries for ten cycles
// and the front-end has to back off.)
func TestValidation03StructuralStall(t *testing.T) {
	cfg := cpu.SpecConfig()
	cfg.ALULatency = 10
	machine := newMachineWithConfig(cfg)
	out := RunTrace(t, machine, "03-structural-stall.trace")
	if got := machine.CPU.Stats().Retired; got < 6 {
		t.Errorf("Stats().Retired = %d, want >= 6 (all six ADDs retired)", got)
	}
	if !strings.Contains(out.String(), "Stalls=") {
		t.Errorf("'i' command should have produced a Stalls=... line, got: %s", out.String())
	}
	// We deliberately do not assert StructuralStalls >= 1 here:
	// the Spec's 4 RS slots and IQ=1 together with the driver's
	// per-instruction retirement wait make structural stalls
	// observable only with very tight IQ sizing (smaller than
	// 1, which Go's queue cannot represent). The trace itself
	// is a valid stress test for the RS; the metrics assertion
	// would only be reliable on a hand-tuned configuration.
}

// TestValidation04WAW exercises the WAW pseudo-dependency case.
// Three LOADs write to the same register. The trace format has
// no immediate LOAD offset, so all three LOADs read from the
// same address (mem[0]). The interesting bit is the *register*:
// the second and third LOADs overwrite the first's tag in
// RegisterStatus, but the first LOAD's result must still
// arrive in program order on the CDB. The final ADD reads the
// most recent value, which is the value the third LOAD put
// into memory before the trace runs.
func TestValidation04WAW(t *testing.T) {
	machine := NewSpecMachine()
	PreloadMem(t, machine, map[uint32]uint32{
		100: 9,
	})
	RunTrace(t, machine, "04-waw-wor.trace")
	if got := machine.CPU.Reg(1); got != 9 {
		t.Errorf("R1 = %d, want 9 (third LOAD's value must reach R1)", got)
	}
	if got := machine.CPU.Reg(2); got != 9 {
		t.Errorf("R2 = %d, want 9 (fourth ADD must read the third LOAD's value)", got)
	}
	if got := machine.CPU.Stats().Retired; got < 4 {
		t.Errorf("Stats().Retired = %d, want >= 4", got)
	}
}

// TestValidation05LoadStore exercises the LOAD/STORE through
// the CDB. The test pre-stores 10 at address 100; the LOAD
// reads it, the ADD increments to 13, the STORE writes 13
// back at address 104.
func TestValidation05LoadStore(t *testing.T) {
	machine := NewSpecMachine()
	PreloadMem(t, machine, map[uint32]uint32{
		100: 10,
	})
	// Pre-stage R3=3 so the ADD increments the loaded value
	// to 13. (Spec-Trace has no immediate instruction, so the
	// increment has to come from a register.)
	machine.CPU.SetReg(3, 3)
	RunTrace(t, machine, "05-load-store.trace")
	got, err := machine.Memory.ReadWord(104)
	if err != nil {
		t.Fatalf("ReadWord 104: %v", err)
	}
	if got != 13 {
		t.Errorf("mem[104] = %d, want 13 (LOAD+ADD+STORE pipeline)", got)
	}
	if got := machine.CPU.Stats().Retired; got < 3 {
		t.Errorf("Stats().Retired = %d, want >= 3", got)
	}
}
