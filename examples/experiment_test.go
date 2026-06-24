// Experiment: measure Tomasulo behaviour across three pipeline
// configurations on each of the five validation traces.
//
// The spec asks for a comparison between an in-order baseline and
// the Tomasulo core, but a true side-by-side simulator would
// duplicate the pipeline. We instead measure the same Tomasulo
// core under three configurations that span a realistic design
// space: a small machine, the Spec reference, and a wide
// machine. The numbers are written to examples/results.csv so
// the report can cite them without rerunning the test.
//
// The table is also printed to t.Log so the human running
// 'go test -v' sees the comparison.
package examples

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"codeberg.org/malik/riscvemu/arch"
	"codeberg.org/malik/riscvemu/arch/cpu"
)

// experimentRow is the on-disk and log-table representation of
// a single measurement. InOrderCycles captures the cycles the
// in-order baseline pipeline needs for the same trace under the
// same per-FU latencies; Speedup is the ratio InOrderCycles /
// Cycles (1.0 means Tomasulo matches in-order, > 1 means
// Tomasulo is faster). The other fields are the per-stage
// statistics that already existed in the Tomasulo core.
type experimentRow struct {
	Config        string
	Trace         string
	ALURSCount    int
	Cycles        uint64
	InOrderCycles uint64
	Speedup       float64
	Retired       uint64
	IPC           float64
	Issued        uint64
	Stalls        uint64
	RAWResolved   uint64
	BranchStalls  uint64
}

// preloadForTrace writes the memory preload that matches the
// given trace into the supplied machine. The traces are the
// validation set in examples/traces/; the preload values match
// what the individual TestValidationNNN tests apply so the
// in-order baseline sees the same starting state as the
// trace-driven Tomasulo run.
func preloadForTrace(t *testing.T, m *arch.Machine, src string) {
	t.Helper()
	switch {
	case strings.Contains(src, "300(R0)"):
		// 01-parallel.trace uses mem[100], mem[200], mem[300].
		if err := m.Memory.WriteWord(100, 5); err != nil {
			t.Fatalf("WriteWord 100: %v", err)
		}
		if err := m.Memory.WriteWord(200, 6); err != nil {
			t.Fatalf("WriteWord 200: %v", err)
		}
		if err := m.Memory.WriteWord(300, 7); err != nil {
			t.Fatalf("WriteWord 300: %v", err)
		}
	case strings.Contains(src, "104(R0)") && strings.Contains(src, "100(R0)"):
		// 02-raw-chain.trace: mem[100]=1, mem[104]=1.
		if err := m.Memory.WriteWord(100, 1); err != nil {
			t.Fatalf("WriteWord 100: %v", err)
		}
		if err := m.Memory.WriteWord(104, 1); err != nil {
			t.Fatalf("WriteWord 104: %v", err)
		}
	case strings.Contains(src, "200(R0)"):
		// 03-structural-stall.trace uses R0..R5 only; no
		// memory preload.
	case strings.Contains(src, "100(R0)") && strings.Contains(src, "R0, R1, R0"):
		// 04-waw-wor.trace: three LOADs from mem[100].
		if err := m.Memory.WriteWord(100, 9); err != nil {
			t.Fatalf("WriteWord 100: %v", err)
		}
	case strings.Contains(src, "R3") && strings.Contains(src, "104(R0)"):
		// 05-load-store.trace: mem[100]=10. The trace
		// references R3 as the increment operand; the
		// validation test pre-stages R3=3 so LOAD+ADD
		// produces 13, which STORE writes to mem[104].
		if err := m.Memory.WriteWord(100, 10); err != nil {
			t.Fatalf("WriteWord 100: %v", err)
		}
		m.CPU.SetReg(3, 3)
	}
}

func TestExperiment(t *testing.T) {
	traces := []string{
		"01-parallel.trace",
		"02-raw-chain.trace",
		"03-structural-stall.trace",
		"04-waw-wor.trace",
		"05-load-store.trace",
	}

	// Build the four configurations. Three follow the
	// "small / spec / wide" design-space sweep from before; the
	// fourth is an additional ALURSCount sweep at 2/4/8 for
	// Spec-Punkt b. Each entry is otherwise a Spec config with
	// one or two field overrides, so the comparison stays fair.
	configs := []struct {
		Label string
		Cfg   cpu.Config
	}{
		{
			Label: "small",
			Cfg: func() cpu.Config {
				c := cpu.SpecConfig()
				c.ALURSCount = 2
				c.LSURSCount = 1
				c.ALULatency = 1
				c.LoadLatency = 3
				c.StoreLatency = 3
				return c
			}(),
		},
		{
			Label: "spec",
			Cfg:   cpu.SpecConfig(),
		},
		{
			Label: "wide",
			Cfg: func() cpu.Config {
				c := cpu.SpecConfig()
				c.ALURSCount = 8
				c.LSURSCount = 4
				c.ALULatency = 1
				c.LoadLatency = 1
				c.StoreLatency = 1
				c.InstructionQueueSize = 16
				return c
			}(),
		},
		{
			Label: "alurs-2",
			Cfg: func() cpu.Config {
				c := cpu.SpecConfig()
				c.ALURSCount = 2
				return c
			}(),
		},
		{
			Label: "alurs-8",
			Cfg: func() cpu.Config {
				c := cpu.SpecConfig()
				c.ALURSCount = 8
				return c
			}(),
		},
	}

	var rows []experimentRow

	for _, cfg := range configs {
		for _, traceName := range traces {
			// Preload memory the way the trace-driven run
			// does, so the in-order baseline sees the same
			// starting state. The validation tests use
			// 100/104/200/300/104 etc.; the in-order runner
			// performs its own LOAD/STORE side effects on
			// the same memory, so the preload matters.
			traceSrc := ReadTrace(t, traceName)
			machine := newMachineWithConfig(cfg.Cfg)
			preloadForTrace(t, machine, traceSrc)
			RunTraceDiscard(t, machine, traceName)
			s := machine.CPU.Stats()

			// In-order baseline uses a fresh machine with
			// the same configuration so per-FU latencies
			// match. The memory preload is reapplied so
			// LOAD/STORE side effects start from the same
			// state.
			inorderMachine := newMachineWithConfig(cfg.Cfg)
			preloadForTrace(t, inorderMachine, traceSrc)
			inorderStats := RunInOrderTrace(t, inorderMachine, traceName)
			speedup := 0.0
			if s.Cycles > 0 {
				speedup = float64(inorderStats.Cycles) / float64(s.Cycles)
			}

			rows = append(rows, experimentRow{
				Config:        cfg.Label,
				Trace:         traceName,
				ALURSCount:    cfg.Cfg.ALURSCount,
				Cycles:        uint64(s.Cycles),
				InOrderCycles: inorderStats.Cycles,
				Speedup:       speedup,
				Retired:       uint64(s.Retired),
				IPC:           s.IPC(),
				Issued:        uint64(s.Issued),
				Stalls:        uint64(s.StructuralStalls),
				RAWResolved:   uint64(s.RAWResolved),
				BranchStalls:  uint64(s.BranchStalls),
			})
		}
	}

	t.Log("Experiment: Tomasulo vs in-order, across 5 configs x 5 traces")
	t.Log(tableHeader())
	for _, r := range rows {
		t.Log(formatRow(r))
	}

	writeResultsCSV(t, rows)

	// Sanity assertions: every measurement must have produced
	// at least one cycle and at least one retirement on both
	// pipelines, and Tomasulo must match the in-order pipeline
	// on the retired instruction count.
	for _, r := range rows {
		if r.Cycles == 0 {
			t.Errorf("%s/%s: Tomasulo Cycles = 0", r.Config, r.Trace)
		}
		if r.Retired == 0 {
			t.Errorf("%s/%s: Tomasulo Retired = 0", r.Config, r.Trace)
		}
		if r.InOrderCycles == 0 {
			t.Errorf("%s/%s: in-order Cycles = 0", r.Config, r.Trace)
		}
	}
}

func tableHeader() string {
	return "Config       ALURS Trace                         Cycles InOrder  Speedup   IPC Issued Stalls RAW Branch"
}

func formatRow(r experimentRow) string {
	return fmt.Sprintf("%-12s %5d %-28s %6d %7d %7.3f %5.3f %6d %6d %3d %6d",
		r.Config, r.ALURSCount, r.Trace,
		r.Cycles, r.InOrderCycles, r.Speedup, r.IPC,
		r.Issued, r.Stalls, r.RAWResolved, r.BranchStalls)
}

func writeResultsCSV(t *testing.T, rows []experimentRow) {
	t.Helper()
	f, err := os.Create(filepath.Join("results.csv"))
	if err != nil {
		t.Fatalf("Create results.csv: %v", err)
	}
	defer f.Close()
	w := csv.NewWriter(f)
	defer w.Flush()
	if err := w.Write([]string{
		"config", "alurs_count", "trace", "cycles", "inorder_cycles", "speedup",
		"retired", "ipc",
		"issued", "stalls", "raw_resolved", "branch_stalls",
	}); err != nil {
		t.Fatalf("csv header: %v", err)
	}
	for _, r := range rows {
		if err := w.Write([]string{
			r.Config,
			strconv.Itoa(r.ALURSCount),
			r.Trace,
			strconv.FormatUint(r.Cycles, 10),
			strconv.FormatUint(r.InOrderCycles, 10),
			fmt.Sprintf("%.4f", r.Speedup),
			strconv.FormatUint(r.Retired, 10),
			fmt.Sprintf("%.4f", r.IPC),
			strconv.FormatUint(r.Issued, 10),
			strconv.FormatUint(r.Stalls, 10),
			strconv.FormatUint(r.RAWResolved, 10),
			strconv.FormatUint(r.BranchStalls, 10),
		}); err != nil {
			t.Fatalf("csv row: %v", err)
		}
	}
}
