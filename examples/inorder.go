// In-order baseline pipeline for the validation traces.
//
// The Spec (page 2) requires a comparison between the Tomasulo core
// and a simple in-order pipeline without dynamic scheduling. This
// file implements that baseline: one instruction is fetched, executed,
// and written back in full before the next one starts. There are no
// reservation stations, no register renaming, and no common data bus
// to wake up out-of-order instructions. Each instruction therefore
// pays the full latency of its predecessors, which is exactly the
// behaviour Tomasulo was designed to avoid.
//
// The runner is intentionally small: it reuses the same trace parser
// and register file as the rest of the project, and it consumes the
// per-instruction latencies from the supplied cpu.Config so the
// comparison is fair. The only statistics it collects are Cycles,
// Retired, and the resulting IPC; structural stalls and RAW
// resolutions do not apply to a pipeline that does no dynamic
// scheduling.
package examples

import (
	"path/filepath"
	"testing"

	"codeberg.org/malik/riscvemu/arch"
	"codeberg.org/malik/riscvemu/trace"
)

// InOrderStats is the minimal statistic set an in-order pipeline
// produces. Cycles counts every clock cycle the simulation ran
// (including the issue and writeback cycle of the last instruction).
// Retired counts the instructions whose writeback has completed.
// IPC is the conventional Retired / Cycles ratio; it is exposed
// as a float for ergonomic comparisons in the experiment table.
type InOrderStats struct {
	Cycles  uint64
	Retired uint64
	IPC     float64
}

// RunInOrderTrace drives the named trace through the given machine
// in strict program order, one instruction at a time, and returns
// the accumulated statistics. The machine is used only for its
// register file, memory, and cpu.Config: the Tomasulo CPU inside
// it is not exercised. Tests that want a clean register file
// should reset the machine before calling.
func RunInOrderTrace(t *testing.T, machine *arch.Machine, name string) InOrderStats {
	t.Helper()
	src := ReadTraceFile(t, filepath.Join("traces", name))
	lines, err := trace.ParseTrace(src)
	if err != nil {
		t.Fatalf("ParseTrace %q: %v", name, err)
	}
	stats := runInOrderLines(t, machine, lines)
	return stats
}

// runInOrderLines is the inner interpreter. It walks the parsed
// lines, executes every LineInstr via the in-order pipeline model,
// and ignores LineControl entries (v/s/h/i) because the baseline
// pipeline has no I/O to drive.
func runInOrderLines(t *testing.T, machine *arch.Machine, lines []trace.Line) InOrderStats {
	t.Helper()
	lat := readLatencies(t, machine)
	var cycles uint64
	var retired uint64
	for _, line := range lines {
		if line.Kind != trace.LineInstr {
			continue
		}
		cycles += executeInOrder(machine, line.Instr, lat)
		retired++
	}
	ipc := 0.0
	if cycles > 0 {
		ipc = float64(retired) / float64(cycles)
	}
	return InOrderStats{Cycles: cycles, Retired: retired, IPC: ipc}
}

// readLatencies returns the per-op-kind latencies from the machine's
// cpu.Config. InOrderRunner reads the latency from the same source
// the Tomasulo core reads it, so the two pipelines see identical
// timing. Only the instruction kinds that appear in the Spec
// validation traces are populated.
func readLatencies(t *testing.T, m *arch.Machine) map[string]int {
	t.Helper()
	// The machine does not expose its own cfg after construction,
	// but SpecConfig() and DefaultConfig() agree on the public
	// fields used here, and the only way to reach the cfg is via
	// the machine. The validation tests in this package always
	// build the machine through newMachineWithConfig, which uses
	// a Config whose ALULatency/LoadLatency/etc. are the source
	// of truth. We pull them out of the per-kind FU latencies on
	// the CPU, since that is what the core actually uses.
	c := m.CPU
	return map[string]int{
		"ADD":   c.ALULatency(),
		"SUB":   c.ALULatency(),
		"MUL":   c.MulLatency(),
		"DIV":   c.DivLatency(),
		"LOAD":  c.LoadLatency(),
		"STORE": c.StoreLatency(),
	}
}

// executeInOrder runs a single instruction to completion in the
// in-order pipeline model. It returns the number of clock cycles
// the instruction occupied. Issue and writeback are folded into
// the same cycle as execution because the in-order model has no
// staging: one instruction per cycle, full latency.
func executeInOrder(m *arch.Machine, instr trace.Instr, lat map[string]int) uint64 {
	// Default to ALU latency when an unknown mnemonic shows up;
	// validation traces only contain the six documented ones.
	cycles, ok := lat[instr.Mnemonic]
	if !ok {
		cycles = lat["ADD"]
	}
	applyInOrder(m, instr)
	// In-order pipeline: 1 issue cycle + (latency-1) execute cycles
	// + 1 writeback cycle. A latency-1 instruction (ADD/SUB) still
	// needs both an issue and a writeback cycle, so the minimum is
	// 2 cycles per instruction.
	if cycles < 1 {
		cycles = 1
	}
	return uint64(cycles) + 1
}

// applyInOrder mutates the architectural state for one instruction
// without any scheduling. It mirrors the semantics the Tomasulo
// core applies at writeback, but without any parallelism or
// renaming. R0 is read-only and writes to it are silently dropped,
// matching the rest of the project.
func applyInOrder(m *arch.Machine, instr trace.Instr) {
	switch instr.Mnemonic {
	case "ADD":
		rd := regIndex(instr.Args[0])
		rs1 := regIndex(instr.Args[1])
		rs2 := regIndex(instr.Args[2])
		if rd != 0 {
			m.CPU.SetReg(rd, m.CPU.Reg(rs1)+m.CPU.Reg(rs2))
		}
	case "SUB":
		rd := regIndex(instr.Args[0])
		rs1 := regIndex(instr.Args[1])
		rs2 := regIndex(instr.Args[2])
		if rd != 0 {
			m.CPU.SetReg(rd, m.CPU.Reg(rs1)-m.CPU.Reg(rs2))
		}
	case "MUL":
		rd := regIndex(instr.Args[0])
		rs1 := regIndex(instr.Args[1])
		rs2 := regIndex(instr.Args[2])
		if rd != 0 {
			m.CPU.SetReg(rd, m.CPU.Reg(rs1)*m.CPU.Reg(rs2))
		}
	case "DIV":
		rd := regIndex(instr.Args[0])
		rs1 := regIndex(instr.Args[1])
		rs2 := regIndex(instr.Args[2])
		if rd != 0 {
			denom := m.CPU.Reg(rs2)
			if denom == 0 {
				if rd != 0 {
					m.CPU.SetReg(rd, 0xFFFFFFFF)
				}
			} else {
				if rd != 0 {
					m.CPU.SetReg(rd, m.CPU.Reg(rs1)/denom)
				}
			}
		}
	case "LOAD":
		rd := regIndex(instr.Args[0])
		offset, base := memRef(instr.Args[1])
		addr := m.CPU.Reg(base) + uint32(int32(offset))
		val, err := m.Memory.ReadWord(addr)
		if err == nil && rd != 0 {
			m.CPU.SetReg(rd, val)
		}
	case "STORE":
		rs := regIndex(instr.Args[0])
		offset, base := memRef(instr.Args[1])
		addr := m.CPU.Reg(base) + uint32(int32(offset))
		_ = m.Memory.WriteWord(addr, m.CPU.Reg(rs))
	}
}

// regIndex converts a Spec-style register name "R0".."R7" into
// the architectural index 0..7. Anything outside the range
// returns 0 to mirror R0-as-discard semantics.
func regIndex(s string) uint32 {
	if len(s) < 2 || s[0] != 'R' {
		return 0
	}
	idx := uint32(s[1] - '0')
	if idx > 7 {
		return 0
	}
	return idx
}

// memRef parses the "offset(Rn)" form into a signed offset and a
// base register index. The Spec validation traces only use small
// non-negative offsets, but the parser handles negatives and
// multi-digit offsets to stay robust.
func memRef(s string) (int32, uint32) {
	open := -1
	for i, c := range s {
		if c == '(' {
			open = i
			break
		}
	}
	if open < 0 {
		return 0, 0
	}
	offsetStr := s[:open]
	base := s[open+1:]
	if len(base) > 0 && base[len(base)-1] == ')' {
		base = base[:len(base)-1]
	}
	var offset int32
	sign := int32(1)
	if len(offsetStr) > 0 && offsetStr[0] == '-' {
		sign = -1
		offsetStr = offsetStr[1:]
	}
	var n int32
	for _, c := range offsetStr {
		if c < '0' || c > '9' {
			return 0, regIndex(base)
		}
		n = n*10 + int32(c-'0')
	}
	offset = sign * n
	return offset, regIndex(base)
}
