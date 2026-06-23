package cli

import (
	"errors"
	"fmt"
	"os"

	"codeberg.org/malik/riscvemu/internal/core"
)

// RunScript loads a program or trace from the given path,
// runs the emulator for the requested number of cycles,
// and prints a one-line stats summary. This is the
// non-interactive CLI frontend: it does not enter a REPL
// or start the TUI, and it returns as soon as the cycles
// have elapsed (or the program is exhausted when cycles
// is 0).
//
// The file is first parsed as assembly; if the assembler
// rejects the input, the loader falls back to the trace
// format. The two parsers are independent, so the same
// path can be used for both program files and trace
// files without an explicit -type flag.
func RunScript(app *core.App, path string, cycles int) error {
	if path == "" {
		return errNoProgram
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read %s: %w", path, err)
	}
	src := string(data)
	if perr := app.LoadProgram(src); perr != nil {
		if terr := app.LoadTrace(src); terr != nil {
			return fmt.Errorf("load %s: not assembly (%v) or trace (%v)", path, perr, terr)
		}
	}
	if cycles < 0 {
		return errors.New("run: cycle count must be non-negative")
	}
	if cycles > 0 {
		if err := app.Step(cycles); err != nil {
			return fmt.Errorf("step %d: %w", cycles, err)
		}
	}
	s := app.Snapshot().Stats
	stalls := s.StructuralStalls + s.BranchStalls
	fmt.Printf("cycles=%d retired=%d IPC=%.3f stalls=%d\n",
		s.Cycles, s.Retired, s.IPC(), stalls)
	return nil
}
