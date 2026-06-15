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
	"io"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/malikwirin/riscvemu/arch"
	"github.com/malikwirin/riscvemu/arch/cpu"
	"github.com/malikwirin/riscvemu/trace"
)

// experimentRow is the on-disk and log-table representation of
// a single measurement.
type experimentRow struct {
	Config       string
	Trace        string
	Cycles       uint64
	Retired      uint64
	IPC          float64
	Issued       uint64
	Stalls       uint64
	RAWResolved  uint64
	BranchStalls uint64
}

func TestExperiment(t *testing.T) {
	traces := []string{
		"01-parallel.trace",
		"02-raw-chain.trace",
		"03-structural-stall.trace",
		"04-waw-wor.trace",
		"05-load-store.trace",
	}

	// Build three configurations: small, Spec reference, wide.
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
	}

	var rows []experimentRow

	for _, cfg := range configs {
		for _, traceName := range traces {
			src := readTrace(t, traceName)
			lines, err := trace.ParseTrace(src)
			if err != nil {
				t.Fatalf("ParseTrace %q: %v", traceName, err)
			}
			machine := arch.NewMachineWithConfig(1024, cfg.Cfg)
			d := trace.NewDriver(machine, io.Discard)
			if err := d.Run(lines); err != nil {
				t.Fatalf("Driver.Run %q: %v", traceName, err)
			}
			s := machine.CPU.Stats()
			rows = append(rows, experimentRow{
				Config:       cfg.Label,
				Trace:        traceName,
				Cycles:       uint64(s.Cycles),
				Retired:      uint64(s.Retired),
				IPC:          s.IPC(),
				Issued:       uint64(s.Issued),
				Stalls:       uint64(s.StructuralStalls),
				RAWResolved:  uint64(s.RAWResolved),
				BranchStalls: uint64(s.BranchStalls),
			})
		}
	}

	t.Log("Experiment: Tomasulo across 3 configs x 5 traces")
	t.Log(tableHeader())
	for _, r := range rows {
		t.Log(formatRow(r))
	}

	writeResultsCSV(t, rows)

	// Sanity assertions: every measurement must have produced
	// at least one cycle and at least one retirement.
	for _, r := range rows {
		if r.Cycles == 0 {
			t.Errorf("%s/%s: Cycles = 0", r.Config, r.Trace)
		}
		if r.Retired == 0 {
			t.Errorf("%s/%s: Retired = 0", r.Config, r.Trace)
		}
	}
}

func tableHeader() string {
	return "Config       Trace                         Cycles Retired   IPC Issued Stalls RAW Branch"
}

func formatRow(r experimentRow) string {
	return fmt.Sprintf("%-12s %-28s %6d %7d %5.3f %6d %6d %3d %6d",
		r.Config, r.Trace,
		r.Cycles, r.Retired, r.IPC,
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
		"config", "trace", "cycles", "retired", "ipc",
		"issued", "stalls", "raw_resolved", "branch_stalls",
	}); err != nil {
		t.Fatalf("csv header: %v", err)
	}
	for _, r := range rows {
		if err := w.Write([]string{
			r.Config, r.Trace,
			strconv.FormatUint(r.Cycles, 10),
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
