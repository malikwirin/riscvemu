package cli

import (
	"bytes"
	"os"

	"codeberg.org/malik/riscvemu/arch"
	"codeberg.org/malik/riscvemu/arch/cpu"
	"codeberg.org/malik/riscvemu/internal/core"
)

// testOwner is a test double for machineOwner. It carries the
// shared core.App so REPL commands can be verified against the
// action layer rather than poking the underlying arch.Machine
// directly. Machine() and Cfg() remain as escape hatches for
// tests that need the raw machine or the configuration.
type testOwner struct {
	app *core.App
	cfg cpu.Config
}

func (t *testOwner) App() *core.App         { return t.app }
func (t *testOwner) Machine() *arch.Machine { return t.app.Machine() }
func (t *testOwner) Cfg() cpu.Config        { return t.cfg }

// captureOutput runs f and returns what is printed to os.Stdout as a string.
func captureOutput(f func()) string {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	f()

	w.Close()
	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	os.Stdout = old
	return buf.String()
}

// withApp runs f with a fresh core.App and a testOwner. Use this
// for tests that verify the REPL command layer against the shared
// action layer.
func withApp(memSize int, cfg cpu.Config, f func(app *core.App, owner *testOwner)) {
	app := core.New(memSize, cfg)
	owner := &testOwner{app: app, cfg: cfg}
	f(app, owner)
}

// withMachineConfig runs f with a machine built from the given
// config. Backward-compatible: it builds the same core.App under
// the hood, so existing tests that need *arch.Machine continue to
// work without refactoring.
func withMachineConfig(memSize int, cfg cpu.Config, f func(m *arch.Machine, owner *testOwner)) {
	withApp(memSize, cfg, func(app *core.App, owner *testOwner) {
		f(app.Machine(), owner)
	})
}

// withMachine runs f with a new Machine of given size and a
// testOwner. The Machine is built from a SpecConfig so the
// historical default layout (32 registers) keeps working.
func withMachine(memSize int, f func(m *arch.Machine, owner *testOwner)) {
	withMachineConfig(memSize, cpu.DefaultConfig(), f)
}

// isRandomized returns true if the values slice is plausibly random (not all zero, not all the same).
func isRandomized(values []uint32) bool {
	if len(values) == 0 {
		return false
	}
	allZero := true
	first := values[0]
	allSame := true
	for _, v := range values {
		if v != 0 {
			allZero = false
		}
		if v != first {
			allSame = false
		}
	}
	return !allZero && !allSame
}
