// Package trace's driver consumes a stream of parsed Lines and
// drives the arch.Machine accordingly. The driver is the only
// place that knows about the Spec's control commands (v, s, h, i)
// and how each must be reflected in the machine's state.
//
// The driver is intentionally simple: one instruction per call to
// machine.CPU.Fetch, then a tight loop on machine.Step until the
// machine has retired the just-fetched instruction. The Tomasulo
// pipeline decides how many cycles that takes; the driver just
// waits. This is the canonical "trace-driven simulation" pattern
// from computer-architecture textbooks: the trace lists the program
// order, and the simulator models the parallel execution.
package trace

import (
	"fmt"
	"io"
	"strconv"
	"strings"

	"codeberg.org/malik/riscvemu/arch"
	"codeberg.org/malik/riscvemu/arch/cpu"
)

// Driver runs a parsed trace through a Machine. The driver writes
// its output (instruction explanations, register dumps, statistics)
// to the configured writer; this keeps the type easy to test
// against a *bytes.Buffer without coupling the trace package to
// os.Stdout.
type Driver struct {
	machine *arch.Machine
	out     io.Writer
	// pc is the trace-level program counter. It is decoupled from
	// machine.PC because the trace enumerates instructions in their
	// source order; jumps and branches in the trace are
	// instructions too (if they were supported), and the driver
	// increments pc only after a successful fetch.
	pc uint32
	// verbose toggles the 'v' control command. The Spec example
	// uses two 'v' lines to bracket a section of the trace, so the
	// command is a stateless toggle rather than a one-shot
	// command.
	verbose bool
	// explainInstr produces a one-line description of an
	// instruction's effect for the verbose mode. It is a field so
	// tests can override it to a deterministic string.
	explainInstr func(instr Instr) string
}

// NewDriver builds a Driver that drives the given machine and
// writes its output to out. out must not be nil; pass
// os.Stdout for the REPL or a *bytes.Buffer for tests.
func NewDriver(machine *arch.Machine, out io.Writer) *Driver {
	return &Driver{
		machine: machine,
		out:     out,
		explainInstr: func(instr Instr) string {
			return defaultExplain(instr)
		},
	}
}

// Run consumes a stream of lines and produces the corresponding
// machine activity. Lines must come from a single ParseTrace or
// ParseTraceFile call so line numbers are unique. The driver
// returns the first machine or parse error encountered.
//
// After the last line, Run returns nil. The caller is then free
// to inspect machine.CPU.Stats() for the final statistics.
func (d *Driver) Run(lines []Line) error {
	for _, line := range lines {
		if err := d.runLine(line); err != nil {
			return err
		}
	}
	return nil
}

// runLine dispatches a single parsed line to the appropriate
// action. The two line kinds (instruction and control) take
// disjoint paths so neither pays the other's overhead.
func (d *Driver) runLine(line Line) error {
	switch line.Kind {
	case LineInstr:
		return d.executeInstr(line)
	case LineControl:
		return d.executeControl(line)
	}
	return fmt.Errorf("unknown line kind %d at line %d", line.Kind, line.LineNum)
}

// executeInstr drives one instruction through the machine. The
// driver's responsibility ends at "encode, fetch, wait for
// retirement"; the actual work (issue, dispatch, writeback) is the
// CPU's job.
//
// We deliberately bypass Machine.Step: that helper re-reads the
// machine's own PC from memory, which fights the trace driver's
// source-order layout. Instead, we Fetch directly at the
// driver's PC and then loop on CPU.RunCycle (issue + execute +
// writeback) until the new instruction has retired.
func (d *Driver) executeInstr(line Line) error {
	word, err := EncodeInstr(line.Instr)
	if err != nil {
		return fmt.Errorf("line %d: encode %q: %w", line.LineNum, line.Instr.Mnemonic, err)
	}
	startRetired := d.machine.CPU.Stats().Retired
	if !d.machine.CPU.Fetch(word, d.pc) {
		return fmt.Errorf("line %d: fetch rejected %q at PC=0x%x (instruction queue full)",
			line.LineNum, line.Instr.Mnemonic, d.pc)
	}
	if d.verbose {
		fmt.Fprintf(d.out, "%s (PC=0x%x)\n", d.explainInstr(line.Instr), d.pc)
	}
	// The driver's PC advances unconditionally; the supported
	// subset (ADD/SUB/MUL/DIV/LOAD/STORE) has no branches, so
	// source order is the only order the trace cares about.
	d.pc += 4
	// Wait for this instruction to retire. Tomasulo issues and
	// retires out of order, so a single RunCycle is not
	// enough: we poll Stats().Retired until it has advanced
	// past the count we observed before the fetch. The
	// Tomasulo pipeline can retire the new instruction in as
	// few as ceil(latency) cycles, so this loop terminates
	// quickly.
	for {
		d.machine.CPU.RunCycle()
		if d.machine.CPU.Stats().Retired > startRetired {
			return nil
		}
	}
}

// executeControl handles the four single-character control
// commands from the Spec (v, s, h, i). Each is a one-shot effect;
// the driver does not retain any state across control commands
// except the 'v' toggle.
func (d *Driver) executeControl(line Line) error {
	switch line.Control.Op {
	case 'v':
		d.verbose = !d.verbose
		return nil
	case 's':
		return d.dumpState()
	case 'h':
		return d.dumpIPC()
	case 'i':
		return d.dumpStalls()
	}
	return fmt.Errorf("line %d: unknown control command %q", line.LineNum, string(line.Control.Op))
}

// dumpState writes the current reservation-station contents and the
// architectural register file to the output. The 's' command
// produces a state snapshot at the moment it is parsed; the next
// instruction in the trace starts from that snapshot.
func (d *Driver) dumpState() error {
	snap := d.machine.CPU.SnapshotRS()
	fmt.Fprintln(d.out, "Reservation stations (ALU/MUL/DIV pool):")
	d.printPool("ALU", snap.ALU)
	fmt.Fprintln(d.out, "Reservation stations (LSU pool):")
	d.printPool("LSU", snap.LSU)
	fmt.Fprintln(d.out, "Registers (R0..R7):")
	regs := d.machine.CPU.SnapshotRegisters()
	for i := uint32(0); i < 8; i++ {
		fmt.Fprintf(d.out, "  R%d = %d  (Qi=R%d)\n", i, regs.V[i], regs.Qi[i])
	}
	return nil
}

// printPool renders one reservation-station pool in a tabular form
// suitable for the spec example. Empty rows are omitted.
func (d *Driver) printPool(label string, entries []cpu.RSEntry) {
	header := fmt.Sprintf("  %s: idx Busy Op            Vj         Vk         Qj  Qk  Rd", label)
	fmt.Fprintln(d.out, header)
	any := false
	for i, e := range entries {
		if !e.Busy {
			continue
		}
		any = true
		opName := opKindName(cpu.OpKind(e.Kind))
		fmt.Fprintf(d.out, "  %s: %2d   %-4v %-12s %-10d %-10d %-3d %-3d %-3d\n",
			label, i, true, opName, e.Vj, e.Vk, e.Qj, e.Qk, e.Rd)
	}
	if !any {
		fmt.Fprintf(d.out, "  %s: (empty)\n", label)
	}
}

// dumpIPC writes the cycles and IPC summary. The Spec example
// uses the 'h' command to show progress, so the output is a
// single line per call.
func (d *Driver) dumpIPC() error {
	s := d.machine.CPU.Stats()
	fmt.Fprintf(d.out, "Cycles=%d  Retired=%d  IPC=%.4f\n",
		s.Cycles, s.Retired, s.IPC())
	return nil
}

// dumpStalls writes the structural / branch stall counts and the
// number of RAW dependencies the Tomasulo CDB has resolved. The
// 'i' command reports stalls-and-resolutions, not instructions.
func (d *Driver) dumpStalls() error {
	s := d.machine.CPU.Stats()
	fmt.Fprintf(d.out, "Stalls=%d (structural=%d, branch=%d)  RAWResolved=%d\n",
		s.StructuralStalls+s.BranchStalls, s.StructuralStalls, s.BranchStalls, s.RAWResolved)
	return nil
}

// defaultExplain produces a one-line description of an instruction's
// effect, used by the 'v' verbose mode. The wording is intentionally
// short: the spec example uses the format "ADD R1, R2, R3 (R1 = R2 + R3)".
func defaultExplain(instr Instr) string {
	switch instr.Mnemonic {
	case "ADD":
		return fmt.Sprintf("%s %s, %s, %s (R%s = R%s + R%s)",
			instr.Mnemonic, instr.Args[0], instr.Args[1], instr.Args[2],
			stripR(instr.Args[0]), stripR(instr.Args[1]), stripR(instr.Args[2]))
	case "SUB":
		return fmt.Sprintf("%s %s, %s, %s (R%s = R%s - R%s)",
			instr.Mnemonic, instr.Args[0], instr.Args[1], instr.Args[2],
			stripR(instr.Args[0]), stripR(instr.Args[1]), stripR(instr.Args[2]))
	case "MUL":
		return fmt.Sprintf("%s %s, %s, %s (R%s = R%s * R%s)",
			instr.Mnemonic, instr.Args[0], instr.Args[1], instr.Args[2],
			stripR(instr.Args[0]), stripR(instr.Args[1]), stripR(instr.Args[2]))
	case "DIV":
		return fmt.Sprintf("%s %s, %s, %s (R%s = R%s / R%s)",
			instr.Mnemonic, instr.Args[0], instr.Args[1], instr.Args[2],
			stripR(instr.Args[0]), stripR(instr.Args[1]), stripR(instr.Args[2]))
	case "LOAD":
		return fmt.Sprintf("LOAD %s, %s (R%s = mem[%s])",
			instr.Args[0], instr.Args[1],
			stripR(instr.Args[0]), instr.Args[1])
	case "STORE":
		return fmt.Sprintf("STORE %s, %s (mem[%s] = R%s)",
			instr.Args[0], instr.Args[1],
			instr.Args[1], stripR(instr.Args[0]))
	}
	return instr.Mnemonic + " " + strings.Join(instr.Args[:], ",")
}

// stripR removes the leading 'R' from a register name like "R3" so
// the verbose explainer can interpolate it into a sentence without
// adding redundant "R"s.
func stripR(reg string) string {
	if len(reg) > 0 && reg[0] == 'R' {
		return reg[1:]
	}
	return reg
}

// opKindName returns a short human label for an OpKind, used by
// the 's' state dump. We avoid the OpKind.String method on the
// hot path because the supported trace subset is small and the
// labels line up with the Spec's vocabulary (ADD, MUL, ...).
func opKindName(k cpu.OpKind) string {
	switch k {
	case cpu.OpADD:
		return "ADD"
	case cpu.OpADDI:
		return "ADDI"
	case cpu.OpSUB:
		return "SUB"
	case cpu.OpMUL:
		return "MUL"
	case cpu.OpDIV:
		return "DIV"
	case cpu.OpLOAD:
		return "LOAD"
	case cpu.OpSTORE:
		return "STORE"
	}
	return "k=" + strconv.Itoa(int(k))
}
