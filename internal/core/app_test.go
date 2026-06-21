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
	"codeberg.org/malik/riscvemu/internal/core"
)

func TestNewBuildsApp(t *testing.T) {
	app := core.New(1024, cpu.SpecConfig())
	if app == nil {
		t.Fatal("New returned nil")
	}
}

func TestCfgReturnsConfig(t *testing.T) {
	want := cpu.SpecConfig()
	app := core.New(1024, want)
	if got := app.Cfg(); got != want {
		t.Errorf("Cfg() = %+v, want %+v", got, want)
	}
}

func TestStepOneAdvancesOneCycle(t *testing.T) {
	app := core.New(1024, cpu.SpecConfig())
	if err := app.Step(1); err != nil {
		t.Fatalf("Step(1): %v", err)
	}
	if got := app.Snapshot().Stats.Cycles; got != 1 {
		t.Errorf("Cycles = %d, want 1", got)
	}
}

func TestStepAdvancesMultipleCycles(t *testing.T) {
	app := core.New(1024, cpu.SpecConfig())
	if err := app.Step(3); err != nil {
		t.Fatalf("Step(3): %v", err)
	}
	if got := app.Snapshot().Stats.Cycles; got != 3 {
		t.Errorf("Cycles = %d, want 3", got)
	}
}

func TestStepZeroIsNoOp(t *testing.T) {
	app := core.New(1024, cpu.SpecConfig())
	if err := app.Step(0); err != nil {
		t.Errorf("Step(0): unexpected error: %v", err)
	}
	if got := app.Snapshot().Stats.Cycles; got != 0 {
		t.Errorf("Cycles = %d, want 0 (Step(0) should be a no-op)", got)
	}
}

func TestStepNegativeRejected(t *testing.T) {
	app := core.New(1024, cpu.SpecConfig())
	if err := app.Step(-1); err == nil {
		t.Errorf("Step(-1): expected error, got nil")
	}
}

func TestResetClearsCyclesAndPC(t *testing.T) {
	app := core.New(1024, cpu.SpecConfig())
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
	app := core.New(1024, cpu.SpecConfig())
	if err := app.LoadProgram("addi x1, x0, 5"); err != nil {
		t.Fatalf("LoadProgram: %v", err)
	}
	if err := app.Step(2); err != nil {
		t.Fatalf("Step(2): %v", err)
	}
	if got := app.Snapshot().Registers[1]; got != 5 {
		t.Errorf("R1 = %d, want 5", got)
	}
}

func TestLoadProgramRejectsEmpty(t *testing.T) {
	app := core.New(1024, cpu.SpecConfig())
	if err := app.LoadProgram(""); err == nil {
		t.Errorf("LoadProgram(\"\"): expected error, got nil")
	}
}

func TestLoadProgramRejectsSyntaxError(t *testing.T) {
	app := core.New(1024, cpu.SpecConfig())
	if err := app.LoadProgram("garbage line"); err == nil {
		t.Errorf("LoadProgram: expected error for garbage input, got nil")
	}
}

func TestLoadTraceParsesSpecFormat(t *testing.T) {
	app := core.New(1024, cpu.SpecConfig())
	if err := app.LoadTrace("ADD R1, R2, R3"); err != nil {
		t.Fatalf("LoadTrace: %v", err)
	}
	if err := app.Step(2); err != nil {
		t.Fatalf("Step(2): %v", err)
	}
	if got := app.Snapshot().Stats.Retired; got < 1 {
		t.Errorf("Retired = %d, want >= 1 (the ADD should have retired)", got)
	}
}

func TestLoadTraceRejectsUnknownMnemonic(t *testing.T) {
	app := core.New(1024, cpu.SpecConfig())
	if err := app.LoadTrace("BEQ R1, R2, 4"); err == nil {
		t.Errorf("LoadTrace: expected error for unsupported mnemonic, got nil")
	}
}

func TestSnapshotIsValueType(t *testing.T) {
	app := core.New(1024, cpu.SpecConfig())
	s1 := app.Snapshot()
	s2 := app.Snapshot()
	s1.PC = 999
	if s2.PC == 999 {
		t.Errorf("Snapshot is not a value type: mutating s1 changed s2.PC")
	}
}

func TestSnapshotReflectsStep(t *testing.T) {
	app := core.New(1024, cpu.SpecConfig())
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
	app := core.New(1024, cpu.SpecConfig())
	if err := app.LoadProgram("addi x1, x0, 5"); err != nil {
		t.Fatalf("LoadProgram: %v", err)
	}
	if err := app.Step(2); err != nil {
		t.Fatalf("Step(2): %v", err)
	}
	if got := app.Snapshot().Registers[1]; got != 5 {
		t.Errorf("R1 = %d, want 5", got)
	}
}

func TestSnapshotCapturesRenameTags(t *testing.T) {
	// Use a long-latency addi so the rename tag survives past
	// the first cycle. With latency 1 the addi retires in the
	// same cycle it issues, which clears the Qi before the
	// test can observe it.
	app := core.New(1024, cpu.Config{
		ALURSCount: 1, ALULatency: 5, LoadLatency: 2, StoreLatency: 2,
		InstructionQueueSize: 4,
	})
	if err := app.LoadProgram("addi x1, x0, 5"); err != nil {
		t.Fatalf("LoadProgram: %v", err)
	}
	if err := app.Step(1); err != nil {
		t.Fatalf("Step(1): %v", err)
	}
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
	app := core.New(1024, cfg)
	if err := app.LoadProgram("addi x1, x0, 5"); err != nil {
		t.Fatalf("LoadProgram: %v", err)
	}
	if err := app.Step(1); err != nil {
		t.Fatalf("Step(1): %v", err)
	}
	snap := app.Snapshot()
	if got := len(snap.RS.ALU); got != cfg.ALURSCount {
		t.Errorf("len(RS.ALU) = %d, want %d", got, cfg.ALURSCount)
	}
}

// TestLoadProgramReportsSyntaxErrorMentionsLine guards a small
// usability detail: the error from LoadProgram should mention
// the offending text so the user can find it in their input.
func TestLoadProgramReportsSyntaxErrorMentionsInput(t *testing.T) {
	app := core.New(1024, cpu.SpecConfig())
	err := app.LoadProgram("garbage line")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "garbage") {
		t.Errorf("error %q should mention the offending input", err)
	}
}

// TestLoadProgramFromFileLoadsAssembly pins that
// LoadProgramFromFile reads a .s file and writes its
// instructions to memory. After a few Step cycles the target
// register must hold the value the assembly program wrote.
func TestLoadProgramFromFileLoadsAssembly(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "add.s")
	if err := os.WriteFile(path, []byte("addi x1, x0, 42\n"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	app := core.New(1024, cpu.SpecConfig())
	if err := app.LoadProgramFromFile(path); err != nil {
		t.Fatalf("LoadProgramFromFile: %v", err)
	}
	if err := app.Step(3); err != nil {
		t.Fatalf("Step(3): %v", err)
	}
	if got := app.Snapshot().Registers[1]; got != 42 {
		t.Errorf("x1 = %d, want 42", got)
	}
}

// TestLoadProgramFromFileEmptyPath pins that an empty path
// returns an error before any file I/O.
func TestLoadProgramFromFileEmptyPath(t *testing.T) {
	app := core.New(1024, cpu.SpecConfig())
	if err := app.LoadProgramFromFile(""); err == nil {
		t.Fatal("LoadProgramFromFile(\"\"): expected error, got nil")
	}
}

// TestLoadProgramFromFileMissingFile pins that a non-existent
// path returns a descriptive error mentioning the path.
func TestLoadProgramFromFileMissingFile(t *testing.T) {
	app := core.New(1024, cpu.SpecConfig())
	err := app.LoadProgramFromFile("/no/such/file.s")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "/no/such/file.s") {
		t.Errorf("error %q should mention the path", err)
	}
}

// TestRebuildEmptyProgram pins that Rebuild on a fresh App
// (no program loaded) succeeds and reflects the new config.
func TestRebuildEmptyProgram(t *testing.T) {
	app := core.New(1024, cpu.SpecConfig())
	newCfg := cpu.SpecConfig()
	newCfg.ALURSCount = 8
	if err := app.Rebuild(newCfg); err != nil {
		t.Fatalf("Rebuild: %v", err)
	}
	if got := app.Cfg().ALURSCount; got != 8 {
		t.Errorf("ALURSCount = %d, want 8", got)
	}
}

// TestRebuildChangesConfig pins that Cfg() reflects the new
// configuration after Rebuild.
func TestRebuildChangesConfig(t *testing.T) {
	app := core.New(1024, cpu.SpecConfig())
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

// TestRebuildPreservesProgram pins that a program loaded
// before Rebuild survives the reconfiguration. We load a
// program, rebuild, step, and check that the register holds
// the value the assembly wrote.
func TestRebuildPreservesProgram(t *testing.T) {
	app := core.New(1024, cpu.SpecConfig())
	if err := app.LoadProgram("addi x1, x0, 7"); err != nil {
		t.Fatalf("LoadProgram: %v", err)
	}
	if err := app.Rebuild(cpu.DefaultConfig()); err != nil {
		t.Fatalf("Rebuild: %v", err)
	}
	if err := app.Step(3); err != nil {
		t.Fatalf("Step(3): %v", err)
	}
	if got := app.Snapshot().Registers[1]; got != 7 {
		t.Errorf("x1 after rebuild = %d, want 7", got)
	}
}
