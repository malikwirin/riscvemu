package trace

import (
	"strings"
	"testing"
)

// TestDriverExecutesSpecExample drives the trace example from the
// Spec page 4 through the Driver. The example uses ADD, MUL, LOAD,
// SUB, DIV in that order with a 'v' toggle on either side and
// 's', 'h', 'i' at the end. We assert the output contains the
// expected substrings rather than a byte-exact match: the spec
// only requires the simulator to support the commands, not the
// exact wording.
//
// The Spec-ADD/SUB/MUL/DIV take three register operands; R0..R7
// are the only valid register names. We preload R2 and R5 with
// non-zero values so LOAD and MUL produce observable effects, and
// we pre-store 7 at memory address 104 (= R2+4) so LOAD R6, 4(R2)
// yields 7.
//
// The two 'v' commands in the spec example bracket a section of
// the trace: the first flips verbose on (so ADD/MUL/LOAD/SUB are
// explained), the second flips it off (so DIV is silent). The test
// asserts the four verbose lines are present and the silent DIV
// line is absent, mirroring the spec's two-toggle pattern.
func TestDriverExecutesSpecExample(t *testing.T) {
	machine := NewTestMachine()
	machine.CPU.SetReg(2, 100)
	machine.CPU.SetReg(5, 10)
	if err := machine.Memory.WriteWord(104, 7); err != nil {
		t.Fatalf("WriteWord: %v", err)
	}
	out := RunDriver(t, machine, `v
ADD R1, R2, R3
MUL R4, R1, R5
LOAD R6, 4(R2)
SUB R7, R6, R4
v
DIV R0, R7, R1
s
h
i
`).String()
	// Verbose explanations for the four instructions that fall
	// between the two 'v' toggles.
	AssertContainsAll(t, out,
		"ADD R1, R2, R3",
		"MUL R4, R1, R5",
		"LOAD R6, 4(R2)",
		"SUB R7, R6, R4",
		"Reservation stations",
		"Registers (R0..R7)",
		"Cycles=", "IPC=",
		"Stalls=", "RAWResolved=",
	)
	// DIV falls outside the verbose window, so its verbose line
	// must be absent. The state-dump ('s') still runs, so DIV
	// is observable in the final register file.
	if strings.Contains(out, "DIV R0, R7, R1") {
		t.Errorf("DIV verbose output should be absent, but appeared:\n%s", out)
	}
	// All five instructions must have retired by the time Run
	// returns; otherwise the driver is leaving the machine in a
	// half-finished state.
	if got := machine.CPU.Stats().Retired; got < 5 {
		t.Errorf("Stats().Retired = %d, want >= 5 (all five instructions retired)", got)
	}
}

// TestDriverVerboseToggle checks that the 'v' command flips the
// verbose mode on and off. We feed three ADD instructions
// sandwiched by toggles and assert that the verbose explanations
// appear only when the mode is on.
func TestDriverVerboseToggle(t *testing.T) {
	machine := NewTestMachine()
	out := RunDriver(t, machine, `v
ADD R1, R2, R3
v
ADD R2, R1, R4
v
ADD R3, R5, R6
`).String()
	if !strings.Contains(out, "ADD R1, R2, R3") {
		t.Errorf("expected verbose output for first ADD, got:\n%s", out)
	}
	if !strings.Contains(out, "ADD R3, R5, R6") {
		t.Errorf("expected verbose output for third ADD, got:\n%s", out)
	}
	// The second ADD (between toggles) is silent: the first 'v'
	// flipped verbose on, the second flipped it off, the third
	// flipped it on again. So the second ADD's verbose line must
	// be absent.
	if got := strings.Count(out, "ADD R2"); got != 0 {
		t.Errorf("expected 0 occurrences of 'ADD R2' (verbose off), got %d\nfull output:\n%s", got, out)
	}
}

// TestDriverHDumpsIPCAndCycles exercises the 'h' command directly
// with a known starting state. The 'h' output must include the
// cycle count and the IPC value.
func TestDriverHDumpsIPCAndCycles(t *testing.T) {
	out := RunDriver(t, NewTestMachine(), `ADD R1, R2, R3
ADD R4, R5, R6
h
`).String()
	AssertContainsAll(t, out, "Cycles=", "IPC=")
}

// TestDumpIPCIncludesFUUtilisation ensures the 'h' command reports
// per-functional-unit utilisation. The MUL and ADDI instructions
// below exercise two distinct FU pools, so the dump must contain a
// 'FU utilisation' header and a 'util=' line for at least one kind.
func TestDumpIPCIncludesFUUtilisation(t *testing.T) {
	out := RunDriver(t, NewTestMachine(), `ADD R1, R2, R3
MUL R4, R1, R5
h
`).String()
	AssertContainsAll(t, out, "FU utilisation:", "util=")
}

// TestDriverIDumpsStallsAndRAW checks the 'i' command. The exact
// stall counts depend on the pipeline configuration, so the test
// only checks that the substring is present.
func TestDriverIDumpsStallsAndRAW(t *testing.T) {
	out := RunDriver(t, NewTestMachine(), `ADD R1, R2, R3
ADD R4, R1, R5
i
`).String()
	AssertContainsAll(t, out, "Stalls=", "RAWResolved=")
}

// TestDriverSDumpsRSAndRegisters checks the 's' command. The
// state dump is tabular: it must show both the reservation-station
// pool headers and the R0..R7 register file.
func TestDriverSDumpsRSAndRegisters(t *testing.T) {
	out := RunDriver(t, NewTestMachine(), `ADD R1, R2, R3
s
`).String()
	AssertContainsAll(t, out,
		"ALU/MUL/DIV pool", "LSU pool",
		"R0 =", "R7 =",
	)
}

// TestDriverPropagatesEncodeError ensures that an encode failure
// short-circuits the run with a meaningful error message. This
// happens in practice only when the parser and encoder disagree
// (e.g. an unsupported mnemonic); we force the error by feeding
// a Line with an invalid mnemonic directly to runLine.
func TestDriverPropagatesEncodeError(t *testing.T) {
	machine := NewTestMachine()
	d := NewDriver(machine, nil)
	bad := Line{Kind: LineInstr, LineNum: 1, Instr: Instr{Mnemonic: "BEQ", Args: [3]string{"R1", "R2", "4"}}}
	err := d.runLine(bad)
	if err == nil {
		t.Fatal("expected error for unsupported mnemonic, got nil")
	}
	if !strings.Contains(err.Error(), "line 1") {
		t.Errorf("error %q should mention line 1", err)
	}
}

// TestDriverUnknownControlCommand checks that an unknown control
// command character surfaces an error.
func TestDriverUnknownControlCommand(t *testing.T) {
	d := NewDriver(NewTestMachine(), nil)
	bad := Line{Kind: LineControl, LineNum: 1, Control: Control{Op: 'X'}}
	if err := d.runLine(bad); err == nil {
		t.Fatal("expected error for unknown control, got nil")
	}
}
