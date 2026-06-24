// Package core is the shared application layer for the three
// project frontends (REPL, TUI, future). It owns an arch.Machine
// and exposes actions plus a value-type Snapshot. Frontends call
// into App; the emulator's internal state stays in core and the
// arch package.
package core

import (
	"fmt"
	"strings"

	"codeberg.org/malik/riscvemu/arch"
	"codeberg.org/malik/riscvemu/arch/cpu"
	"codeberg.org/malik/riscvemu/assembler"
	"codeberg.org/malik/riscvemu/trace"
)

// App owns an arch.Machine and exposes the actions all frontends
// share. The Machine itself is private so frontends cannot
// accidentally bypass the action layer; tests that need the
// raw Machine can use the Machine() escape hatch.
type App struct {
	machine     *arch.Machine
	cfg         cpu.Config
	memorySize  int
	lastProgram []assembler.Instruction
}

// New builds a fresh App with the given memory size and pipeline
// configuration. The SpecConfig and DefaultConfig helpers live
// in the arch/cpu package; this constructor does not pick a
// default because the call site (REPL startup, TUI startup)
// should choose deliberately.
func New(memSize int, cfg cpu.Config) *App {
	m := arch.NewMachineWithConfig(memSize, cfg)
	return &App{machine: m, cfg: cfg, memorySize: memSize}
}

// Cfg returns the configuration the App was built with. Frontends
// use it to render the active pipeline (e.g. the 'config' REPL
// command or the TUI's settings panel).
func (a *App) Cfg() cpu.Config { return a.cfg }

// Machine returns the underlying arch.Machine. This is an
// escape hatch for tests and for code that genuinely needs the
// raw machine; production frontends should call the App's
// action methods instead.
func (a *App) Machine() *arch.Machine { return a.machine }

// Step advances the emulator by n clock cycles. n must be
// non-negative; n == 0 is a no-op. Each cycle performs fetch,
// issue, execute, and writeback via the underlying Machine.
func (a *App) Step(n int) error {
	if n < 0 {
		return fmt.Errorf("Step: count must be non-negative, got %d", n)
	}
	for i := 0; i < n; i++ {
		if err := a.machine.Step(); err != nil {
			return fmt.Errorf("Step cycle %d: %w", i+1, err)
		}
	}
	return nil
}

// Reset returns the emulator to its initial state. The CPU and
// memory are rebuilt from the original configuration.
func (a *App) Reset() error {
	return a.machine.Reset()
}

// LoadProgram parses a RISC-V assembly program and writes it to
// memory starting at address 0. Multi-line input is supported
// (one instruction per non-blank line). An empty source or a
// parse error returns a descriptive error; the caller is
// expected to surface it to the user.
func (a *App) LoadProgram(src string) error {
	if strings.TrimSpace(src) == "" {
		return fmt.Errorf("LoadProgram: empty source")
	}
	var prog []assembler.Instruction
	for i, line := range strings.Split(src, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		instr, err := assembler.ParseInstruction(trimmed)
		if err != nil {
			return fmt.Errorf("LoadProgram line %d: %w", i+1, err)
		}
		prog = append(prog, instr)
	}
	if len(prog) == 0 {
		return fmt.Errorf("LoadProgram: no instructions in source")
	}
	a.lastProgram = prog
	return a.machine.LoadProgram(prog, 0)
}

// LoadProgramFromProg writes a pre-parsed program to memory. Use
// this when the caller already has []assembler.Instruction
// (e.g. from assembler.AssembleFile).
func (a *App) LoadProgramFromProg(prog []assembler.Instruction) error {
	if len(prog) == 0 {
		return fmt.Errorf("LoadProgramFromProg: empty program")
	}
	a.lastProgram = prog
	return a.machine.LoadProgram(prog, 0)
}

// LoadProgramFromFile reads a RISC-V assembly source from disk
// and writes the parsed instructions to memory. The file may
// contain one instruction per line; comments and blank lines
// are handled by the assembler package. An empty path or a
// missing/unreadable file returns a descriptive error. The
// Machine's PC is reset to 0 after a successful load.
func (a *App) LoadProgramFromFile(path string) error {
	if path == "" {
		return fmt.Errorf("LoadProgramFromFile: empty path")
	}
	prog, err := assembler.AssembleFile(path)
	if err != nil {
		return fmt.Errorf("LoadProgramFromFile %q: %w", path, err)
	}
	a.lastProgram = prog
	if err := a.machine.LoadProgram(prog, 0); err != nil {
		return fmt.Errorf("LoadProgramFromFile: %w", err)
	}
	return nil
}

// Rebuild reconstructs the underlying machine with a new
// pipeline configuration. The *App pointer identity is
// preserved so every frontend (REPL, TUI) sees the new state
// through the same reference. Any program currently in memory
// is re-loaded from the cached parse tree so the user's
// program survives the reconfiguration. Calling Rebuild with
// the same configuration is cheap (one machine rebuild + one
// program reload) and is supported.
func (a *App) Rebuild(newCfg cpu.Config) error {
	a.machine = arch.NewMachineWithConfig(a.memorySize, newCfg)
	a.cfg = newCfg
	if len(a.lastProgram) == 0 {
		return nil
	}
	return a.machine.LoadProgram(a.lastProgram, 0)
}

// LoadTrace parses a Spec-format trace and writes the encoded
// instructions to memory starting at address 0. The trace may
// contain ADD/SUB/MUL/DIV/LOAD/STORE instructions plus the v/s/h/i
// control commands; control commands are skipped here (the
// caller can read them from the parsed Line stream if it cares).
// The Machine's PC is reset to 0 after a successful load.
func (a *App) LoadTrace(src string) error {
	if strings.TrimSpace(src) == "" {
		return fmt.Errorf("LoadTrace: empty source")
	}
	lines, err := trace.ParseTrace(src)
	if err != nil {
		return err
	}
	wordIdx := 0
	for _, line := range lines {
		if line.Kind != trace.LineInstr {
			continue
		}
		word, err := trace.EncodeInstr(line.Instr)
		if err != nil {
			return fmt.Errorf("LoadTrace line %d: %w", line.LineNum, err)
		}
		if err := a.machine.Memory.WriteWord(uint32(wordIdx*4), word); err != nil {
			return fmt.Errorf("LoadTrace write at 0x%x: %w", wordIdx*4, err)
		}
		wordIdx++
	}
	a.machine.PC = 0
	return nil
}

// Snapshot returns a value-type copy of the emulator's current
// state. Frontends cache the last snapshot and render it; the
// copy is independent of the live machine, so mutating the
// returned Snapshot has no effect on the emulator.
func (a *App) Snapshot() Snapshot {
	regs := a.machine.CPU.SnapshotRegisters()
	return Snapshot{
		PC:        a.machine.PC,
		Registers: regs.V,
		Qi:        regs.Qi,
		RS:        a.machine.CPU.SnapshotRS(),
		Stats:     a.machine.CPU.Stats(),
	}
}
