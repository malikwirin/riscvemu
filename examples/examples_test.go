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

	"codeberg.org/malik/riscvemu/arch/cpu"
	"codeberg.org/malik/riscvemu/assembler"
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
// do not depend on each other. The cycle bound below is loose
// enough to absorb driver-level fetch overhead but tight enough
// to fail if the pipeline degenerates into serial execution
// (which would be 12+ cycles for the six instructions on the
// Spec's LOAD/STORE=2 / ADDI=1 latencies).
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
	// Sanity bound on parallel execution: six instructions
	// (3 LOADs + 3 ADDs) with LOAD/STORE latency 2 and ADDI
	// latency 1 must finish well below the serial lower bound
	// of 12 cycles. The trace driver fetches one instruction at
	// a time and waits for retirement before fetching the next,
	// which adds an instruction-granularity of overhead on top
	// of the Tomasulo core. The bound of 10 is loose enough to
	// absorb that overhead but tight enough to fail if the
	// pipeline degenerates into serial execution.
	if got := machine.CPU.Stats().Cycles; got > 10 {
		t.Errorf("Cycles = %d, want <= 10 (independent LOAD+ADD pairs must run in parallel)", got)
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

// TestValidation02RAWChainComparesInOrder is the spec-mandated
// side-by-side comparison: the same RAW chain must run faster on
// the Tomasulo core than on a simple in-order pipeline. Both runs
// use the same instruction latencies (SpecConfig), so the only
// difference is dynamic scheduling. For an in-order pipeline the
// chain serialises completely (each instruction waits for its
// producer), so the cycle count is the sum of the per-instruction
// latencies. The Tomasulo core still has to honour the RAW edges,
// but the two independent LOADs in front can overlap with the
// first ADD; the cycles count must therefore be strictly lower
// than the in-order count.
func TestValidation02RAWChainComparesInOrder(t *testing.T) {
	cfg := cpu.SpecConfig()
	// Tomasulo run.
	tomasuloMachine := newMachineWithConfig(cfg)
	PreloadMem(t, tomasuloMachine, map[uint32]uint32{100: 1, 104: 1})
	RunTrace(t, tomasuloMachine, "02-raw-chain.trace")
	tomasuloCycles := tomasuloMachine.CPU.Stats().Cycles
	// InOrder baseline run.
	inorderMachine := newMachineWithConfig(cfg)
	PreloadMem(t, inorderMachine, map[uint32]uint32{100: 1, 104: 1})
	inorderStats := RunInOrderTrace(t, inorderMachine, "02-raw-chain.trace")
	if inorderStats.Cycles == 0 {
		t.Fatalf("in-order pipeline produced 0 cycles (run failed silently)")
	}
	if tomasuloCycles >= inorderStats.Cycles {
		t.Errorf("Tomasulo did not beat in-order: tomasulo=%d, inorder=%d (Tomasulo must be strictly faster on this RAW chain)",
			tomasuloCycles, inorderStats.Cycles)
	}
	// Sanity: both pipelines retired the same number of instructions.
	if got := tomasuloMachine.CPU.Stats().Retired; got != inorderStats.Retired {
		t.Errorf("retired count mismatch: tomasulo=%d, inorder=%d", got, inorderStats.Retired)
	}
}

// TestValidation03StructuralStall exercises the case where the
// reservation stations are full. Six independent ADDs are
// pre-loaded into the instruction queue; with a single ALU-RS
// slot and a long ALU latency, the second ADD cannot be issued
// until the first one has been dispatched to the functional
// unit and the RS slot has been freed. That wait is the
// structural stall the StructuralStalls statistic counts.
//
// The trace driver (trace.Run) only fetches one instruction at
// a time, so a stress test that fills the IQ must bypass the
// driver and push instructions directly via the CPU API. The
// 'i' trace command in the driver still produces a Stalls=
// line for the smoke check, but the stall count itself comes
// from the manual RunCycle loop.
func TestValidation03StructuralStall(t *testing.T) {
	cfg := cpu.SpecConfig()
	cfg.ALURSCount = 1
	cfg.ALULatency = 5
	machine := newMachineWithConfig(cfg)
	// Encode the same six ADDs as in 03-structural-stall.trace
	// and push them all into the instruction queue.
	asmLines := []string{
		"add x0, x0, x1",
		"add x1, x0, x0",
		"add x2, x0, x0",
		"add x3, x0, x0",
		"add x4, x0, x0",
		"add x5, x0, x0",
	}
	for i, line := range asmLines {
		pc := uint32(i * 4)
		word, err := assembler.ParseInstruction(line)
		if err != nil {
			t.Fatalf("ParseInstruction %q: %v", line, err)
		}
		if !machine.CPU.Fetch(uint32(word), pc) {
			t.Fatalf("Fetch rejected %q at PC=0x%x (IQ full?)", line, pc)
		}
	}
	// Run until all six ADDs have retired, with a generous
	// safety bound.
	for i := 0; i < 200; i++ {
		machine.CPU.RunCycle()
		if machine.CPU.Stats().Retired >= 6 {
			break
		}
	}
	if got := machine.CPU.Stats().Retired; got != 6 {
		t.Errorf("Retired = %d, want 6 (all six ADDs retired)", got)
	}
	if got := machine.CPU.Stats().StructuralStalls; got < 1 {
		t.Errorf("StructuralStalls = %d, want >= 1 (single-slot RS must stall the front-end)", got)
	}
}

// TestValidation03StructuralStallTraceOutput keeps the smoke check
// that the trace 'i' command produces a Stalls= line. The full
// stall assertion lives in TestValidation03StructuralStall above
// (which drives the CPU directly so the IQ is actually full); this
// test pins the trace output format.
func TestValidation03StructuralStallTraceOutput(t *testing.T) {
	machine := NewSpecMachine()
	out := RunTrace(t, machine, "03-structural-stall.trace")
	if !strings.Contains(out.String(), "Stalls=") {
		t.Errorf("'i' command should have produced a Stalls=... line, got: %s", out.String())
	}
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
