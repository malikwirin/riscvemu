// Package testutil is the cross-paket helper module for the
// test suites of the riscvemu project. It exposes the few
// setup steps that appear in three or more packages' test
// files, so that the boilerplate stays in one place. Anything
// only one package uses lives in that package's own testutil
// file (e.g. cli/testutil.go, examples/testutil.go) instead.
package testutil

import (
	"testing"

	"codeberg.org/malik/riscvemu/arch/cpu"
	"codeberg.org/malik/riscvemu/internal/core"
)

// NewSpecApp builds a fresh *core.App with the Spec-mandated
// 1024-byte memory and configuration. This is the most common
// app construction in the test suites (tui, internal/core,
// cmd/tui, cli/app_delegation_test.go) and keeps the magic
// numbers in one place.
func NewSpecApp() *core.App {
	return core.New(1024, cpu.SpecConfig())
}

// NewAppWith builds a fresh *core.App with the given memory
// size and configuration. Use this for tests that need a
// non-default layout (e.g. reduced register count, custom
// latencies) and do not want to repeat the core.New call
// with the same arguments everywhere.
func NewAppWith(memSize int, cfg cpu.Config) *core.App {
	return core.New(memSize, cfg)
}

// LoadProgramOrFail loads the given assembly source into the
// app and fails the test on parse error. It hides the
// t.Helper + t.Fatalf boilerplate that appears in sixteen
// test functions across the tui, internal/core, and
// cli packages.
func LoadProgramOrFail(t *testing.T, app *core.App, src string) {
	t.Helper()
	if err := app.LoadProgram(src); err != nil {
		t.Fatalf("LoadProgram: %v", err)
	}
}

// StepOrFail advances the app by n cycles and fails the test
// on step error. Same rationale as LoadProgramOrFail.
func StepOrFail(t *testing.T, app *core.App, n int) {
	t.Helper()
	if err := app.Step(n); err != nil {
		t.Fatalf("Step(%d): %v", n, err)
	}
}
