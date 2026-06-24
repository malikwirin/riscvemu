// Package core is the shared application layer for the three
// project frontends (REPL, TUI, future). It owns an arch.Machine
// and exposes actions plus a value-type Snapshot. Frontends call
// into App; the emulator's internal state stays in core and the
// arch package.
package core_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"codeberg.org/malik/riscvemu/arch/cpu"
	"codeberg.org/malik/riscvemu/internal/testutil"
)

func TestNewBuildsApp(t *testing.T) {
	app := testutil.NewSpecApp()
	if app == nil {
		t.Fatal("New returned nil")
	}
}

func TestCfgReturnsConfig(t *testing.T) {
	want := cpu.SpecConfig()
	app := testutil.NewAppWith(1024, want)
	if got := app.Cfg(); got != want {
		t.Errorf("Cfg() = %+v, want %+v", got, want)
	}
}

// TestStep pins the Step contract: positive n advances
// cycles by n, zero is a no-op, and negative n is rejected
// before any side effect.
func TestStep(t *testing.T) {
	cases := []struct {
		name      string
		n         int
		wantErr   bool
		wantCycle uint64
	}{
		{"n=1 advances by 1", 1, false, 1},
		{"n=3 advances by 3", 3, false, 3},
		{"n=0 is no-op", 0, false, 0},
		{"n=-1 rejected", -1, true, 0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			app := testutil.NewSpecApp()
			err := app.Step(tc.n)
			if tc.wantErr && err == nil {
				t.Errorf("Step(%d): want error, got nil", tc.n)
			}
			if !tc.wantErr && err != nil {
				t.Errorf("Step(%d): %v", tc.n, err)
			}
			if got := app.Snapshot().Stats.Cycles; got != tc.wantCycle {
				t.Errorf("Cycles = %d, want %d", got, tc.wantCycle)
			}
		})
	}
}

func TestResetClearsCyclesAndPC(t *testing.T) {
	app := testutil.NewSpecApp()
	if err := app.Step(5); err != nil {
		t.Fatalf("Step(5): %v", err)
	}
	if err := app.Reset(); err != nil {
		t.Fatalf("Reset: %v", err)
	}
	s := app.Snapshot()
	if s.Stats.Cycles != 0 {
		t.Errorf("Cycles = %d after reset, want 0", s.Stats.Cycles)
	}
	if s.PC != 0 {
		t.Errorf("PC = %d after reset, want 0", s.PC)
	}
}

func TestLoadProgramParsesAsm(t *testing.T) {
	app := testutil.NewSpecApp()
	testutil.LoadProgramOrFail(t, app, "addi x1, x0, 5")
	testutil.StepOrFail(t, app, 2)
	if got := app.Snapshot().Registers[1]; got != 5 {
		t.Errorf("R1 = %d, want 5", got)
	}
}

// TestLoadProgramRejects pins the parser-error class:
// empty input, garbage line, and the requirement that the
// error mention the offending text so the user can locate
// it in their input.
func TestLoadProgramRejects(t *testing.T) {
	cases := []struct {
		input   string
		wantSub string
	}{
		{"", ""},
		{"garbage line", "garbage"},
	}
	for _, tc := range cases {
		t.Run(tc.input, func(t *testing.T) {
			app := testutil.NewSpecApp()
			err := app.LoadProgram(tc.input)
			if err == nil {
				t.Fatalf("LoadProgram(%q): want error, got nil", tc.input)
			}
			if tc.wantSub != "" && !strings.Contains(err.Error(), tc.wantSub) {
				t.Errorf("error %q should mention %q", err, tc.wantSub)
			}
		})
	}
}

// TestLoadTrace pins the trace parser: a Spec-format ADD
// line parses and the instruction eventually retires, while
// an unsupported mnemonic (BEQ) is rejected.
func TestLoadTrace(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		app := testutil.NewSpecApp()
		if err := app.LoadTrace("ADD R1, R2, R3"); err != nil {
			t.Fatalf("LoadTrace: %v", err)
		}
		if err := app.Step(2); err != nil {
			t.Fatalf("Step(2): %v", err)
		}
		if got := app.Snapshot().Stats.Retired; got < 1 {
			t.Errorf("Retired = %d, want >= 1", got)
		}
	})
	t.Run("rejects unknown mnemonic", func(t *testing.T) {
		app := testutil.NewSpecApp()
		if err := app.LoadTrace("BEQ R1, R2, 4"); err == nil {
			t.Errorf("LoadTrace(BEQ): want error, got nil")
		}
	})
}

func TestSnapshotIsValueType(t *testing.T) {
	app := testutil.NewSpecApp()
	s1 := app.Snapshot()
	s2 := app.Snapshot()
	s1.PC = 999
	if s2.PC == 999 {
		t.Errorf("Snapshot is not a value type: mutating s1 changed s2.PC")
	}
}

func TestSnapshotReflectsStep(t *testing.T) {
	app := testutil.NewSpecApp()
	if got := app.Snapshot().Stats.Cycles; got != 0 {
		t.Fatalf("initial Cycles = %d, want 0", got)
	}
	if err := app.Step(2); err != nil {
		t.Fatalf("Step(2): %v", err)
	}
	if got := app.Snapshot().Stats.Cycles; got != 2 {
		t.Errorf("Cycles = %d, want 2", got)
	}
}

func TestSnapshotCapturesRegisters(t *testing.T) {
	app := testutil.NewSpecApp()
	testutil.LoadProgramOrFail(t, app, "addi x1, x0, 5")
	testutil.StepOrFail(t, app, 2)
	if got := app.Snapshot().Registers[1]; got != 5 {
		t.Errorf("R1 = %d, want 5", got)
	}
}

func TestSnapshotCapturesRenameTags(t *testing.T) {
	// Use a long-latency addi so the rename tag survives past
	// the first cycle. With latency 1 the addi retires in the
	// same cycle it issues, which clears the Qi before the
	// test can observe it.
	app := testutil.NewAppWith(1024, cpu.Config{
		ALURSCount: 1, ALULatency: 5, LoadLatency: 2, StoreLatency: 2,
		InstructionQueueSize: 4,
	})
	testutil.LoadProgramOrFail(t, app, "addi x1, x0, 5")
	testutil.StepOrFail(t, app, 1)
	if got := app.Snapshot().Qi[1]; got == cpu.NoTag {
		t.Errorf("Qi[x1] = NoTag after issue, want a rename tag (long-latency addi should still be in flight)")
	}
}

func TestSnapshotCapturesRS(t *testing.T) {
	// The Snapshot exposes an RSSnapshot with two slices whose
	// lengths match the configured RS counts. The slices may
	// be nil for a zero-size pool (NewCPU does not allocate
	// make([]RSEntry, 0) for n=0; the resulting append on a
	// nil slice is nil). What we pin here is the public field
	// shape: the Snapshot carries the RSSnapshot, and the
	// ALU slice has at least the number of slots the
	// configuration asked for.
	cfg := cpu.SpecConfig() // 4 ALU-RS, 3 LSU-RS
	app := testutil.NewAppWith(1024, cfg)
	testutil.LoadProgramOrFail(t, app, "addi x1, x0, 5")
	testutil.StepOrFail(t, app, 1)
	snap := app.Snapshot()
	if got := len(snap.RS.ALU); got != cfg.ALURSCount {
		t.Errorf("len(RS.ALU) = %d, want %d", got, cfg.ALURSCount)
	}
}

// TestLoadProgramFromFile pins the file-based loader: a
// valid .s file is parsed and its writes are observed, an
// empty path is rejected, and a non-existent path produces
// an error that names the path so the user can find it.
func TestLoadProgramFromFile(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "add.s")
		if err := os.WriteFile(path, []byte("addi x1, x0, 42\n"), 0o644); err != nil {
			t.Fatalf("WriteFile: %v", err)
		}
		app := testutil.NewSpecApp()
		if err := app.LoadProgramFromFile(path); err != nil {
			t.Fatalf("LoadProgramFromFile: %v", err)
		}
		if err := app.Step(3); err != nil {
			t.Fatalf("Step(3): %v", err)
		}
		if got := app.Snapshot().Registers[1]; got != 42 {
			t.Errorf("x1 = %d, want 42", got)
		}
	})
	t.Run("rejects empty path", func(t *testing.T) {
		app := testutil.NewSpecApp()
		if err := app.LoadProgramFromFile(""); err == nil {
			t.Error("LoadProgramFromFile(\"\"): want error, got nil")
		}
	})
	t.Run("rejects missing file with path in error", func(t *testing.T) {
		app := testutil.NewSpecApp()
		err := app.LoadProgramFromFile("/no/such/file.s")
		if err == nil {
			t.Fatal("want error, got nil")
		}
		if !strings.Contains(err.Error(), "/no/such/file.s") {
			t.Errorf("error %q should mention the path", err)
		}
	})
}

func TestRebuildEmptyProgram(t *testing.T) {
	app := testutil.NewSpecApp()
	newCfg := cpu.SpecConfig()
	newCfg.ALURSCount = 8
	if err := app.Rebuild(newCfg); err != nil {
		t.Fatalf("Rebuild: %v", err)
	}
	if got := app.Cfg().ALURSCount; got != 8 {
		t.Errorf("ALURSCount = %d, want 8", got)
	}
}

func TestRebuildChangesConfig(t *testing.T) {
	app := testutil.NewSpecApp()
	newCfg := cpu.DefaultConfig()
	newCfg.ALURSCount = 6
	newCfg.RegisterCount = 16
	if err := app.Rebuild(newCfg); err != nil {
		t.Fatalf("Rebuild: %v", err)
	}
	if got := app.Cfg().ALURSCount; got != 6 {
		t.Errorf("ALURSCount = %d, want 6", got)
	}
	if got := app.Cfg().RegisterCount; got != 16 {
		t.Errorf("RegisterCount = %d, want 16", got)
	}
}

func TestRebuildPreservesProgram(t *testing.T) {
	app := testutil.NewSpecApp()
	testutil.LoadProgramOrFail(t, app, "addi x1, x0, 7")
	if err := app.Rebuild(cpu.DefaultConfig()); err != nil {
		t.Fatalf("Rebuild: %v", err)
	}
	testutil.StepOrFail(t, app, 3)
	if got := app.Snapshot().Registers[1]; got != 7 {
		t.Errorf("x1 after rebuild = %d, want 7", got)
	}
}
