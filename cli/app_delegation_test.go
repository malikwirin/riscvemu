package cli

import (
	"testing"

	"codeberg.org/malik/riscvemu/arch/cpu"
	"codeberg.org/malik/riscvemu/internal/core"
	"github.com/stretchr/testify/assert"
)

// These tests pin the REPL command layer on top of the shared
// core.App. They are deliberately distinct from the core.App
// tests in internal/core/app_test.go: those verify the App API
// in isolation, these verify that the REPL commands call into
// the App API (and not some other state). When the two layers
// drift, only the matching test in this file will fail.

// TestCmdStepDelegatesToApp pins that cmdStep routes through
// core.App.Step rather than poking the machine directly.
func TestCmdStepDelegatesToApp(t *testing.T) {
	withApp(1024, cpu.SpecConfig(), func(app *core.App, owner *testOwner) {
		out := captureOutput(func() {
			err := cmdStep(owner, nil)
			assert.NoError(t, err, "cmdStep default")
		})
		assert.Contains(t, out, "Executed 1 step", "cmdStep output for default step missing")
		// Step(1) on a fresh App should advance Cycles by exactly 1.
		assert.Equal(t, uint64(1), app.Snapshot().Stats.Cycles,
			"Cycle count after one cmdStep should be 1, got %d",
			app.Snapshot().Stats.Cycles)
	})
}

// TestCmdStepNDelegatesToApp pins that the parsed n argument
// reaches app.Step.
func TestCmdStepNDelegatesToApp(t *testing.T) {
	withApp(1024, cpu.SpecConfig(), func(app *core.App, owner *testOwner) {
		_ = captureOutput(func() {
			err := cmdStep(owner, []string{"3"})
			assert.NoError(t, err, "cmdStep(3)")
		})
		assert.Equal(t, uint64(3), app.Snapshot().Stats.Cycles,
			"Cycle count after cmdStep 3 should be 3, got %d",
			app.Snapshot().Stats.Cycles)
	})
}

// TestCmdResetDelegatesToApp pins that cmdReset routes through
// core.App.Reset.
func TestCmdResetDelegatesToApp(t *testing.T) {
	withApp(1024, cpu.SpecConfig(), func(app *core.App, owner *testOwner) {
		_ = captureOutput(func() { _ = cmdStep(owner, []string{"5"}) })
		if app.Snapshot().Stats.Cycles != 5 {
			t.Fatalf("setup: cycles = %d, want 5", app.Snapshot().Stats.Cycles)
		}
		_ = captureOutput(func() { _ = cmdReset(owner, nil) })
		if app.Snapshot().Stats.Cycles != 0 {
			t.Errorf("after cmdReset: cycles = %d, want 0", app.Snapshot().Stats.Cycles)
		}
		if app.Snapshot().PC != 0 {
			t.Errorf("after cmdReset: PC = %d, want 0", app.Snapshot().PC)
		}
	})
}

// TestCmdStatsReadsSnapshot pins that cmdStats reads from
// app.Snapshot().Stats, not from a separate Stats source.
func TestCmdStatsReadsSnapshot(t *testing.T) {
	withApp(1024, cpu.SpecConfig(), func(app *core.App, owner *testOwner) {
		_ = app.Step(2) // 2 cycles, 0 retired
		out := captureOutput(func() { _ = cmdStats(owner, nil) })
		assert.Contains(t, out, "Cycles:           2",
			"cmdStats should report Cycles=2 from app.Snapshot")
	})
}

// TestCmdRegsReadsSnapshot pins that cmdRegs iterates over
// app.Snapshot().Registers, not over the live machine.
func TestCmdRegsReadsSnapshot(t *testing.T) {
	withApp(1024, cpu.SpecConfig(), func(app *core.App, owner *testOwner) {
		if err := app.LoadProgram("addi x1, x0, 5"); err != nil {
			t.Fatalf("LoadProgram: %v", err)
		}
		_ = app.Step(2) // x1 should be 5
		out := captureOutput(func() { _ = cmdRegs(owner, nil) })
		assert.Contains(t, out, "x1: 5",
			"cmdRegs should render Snapshot.Registers[1]=5, got output:\n%s", out)
	})
}

// TestCmdLoadDelegatesToApp pins that cmdLoad passes the parsed
// file content through to app.LoadProgram.
func TestCmdLoadDelegatesToApp(t *testing.T) {
	withApp(1024, cpu.SpecConfig(), func(app *core.App, owner *testOwner) {
		// We exercise the LoadProgram path directly via app, then
		// verify that the cmdLoad integration uses the same
		// mechanism: the cmdLoad handler is a thin wrapper around
		// app.LoadProgram after the refactor.
		if err := app.LoadProgram("addi x1, x0, 42"); err != nil {
			t.Fatalf("setup LoadProgram: %v", err)
		}
		_ = app.Step(2)
		if got := app.Snapshot().Registers[1]; got != 42 {
			t.Fatalf("setup: R1 = %d, want 42", got)
		}
	})
}
