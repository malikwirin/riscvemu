package cli

import (
	"errors"
	"fmt"

	"codeberg.org/malik/riscvemu/arch/cpu"
	"codeberg.org/malik/riscvemu/internal/core"
	"codeberg.org/malik/riscvemu/tui"
)

// Run is the canonical entry point for the riscvemu binary.
// It splits the supplied command-line arguments into a
// frontend selector and the remaining pipeline-flag
// overrides, then dispatches to the matching frontend:
//
//   - subcommand "tui" or no subcommand (when no positional
//     argument is given) -> TUI
//   - subcommand "repl" -> REPL
//   - subcommand "run" or a positional argument -> CLI
//     (load + run + print stats, then exit)
//
// All subcommands and positional arguments can be combined
// with the pipeline flags understood by ConfigFromFlags;
// the flags may appear before or after the subcommand.
//
// Run returns any error from the chosen frontend. A
// -help / help request returns nil after printing the
// help text, so the caller can exit cleanly.
func Run(args []string) error {
	sub, rest := splitSubcommand(args)
	if sub == "help" || sub == "-help" || sub == "--help" {
		fmt.Print(FlagHelp)
		return nil
	}
	// The empty subcommand from splitSubcommand means the
	// first token was neither a recognised subcommand nor
	// a -help flag. In that case the dispatcher falls
	// through to the CLI path: any token that does not
	// look like a -flag is a positional program file, and
	// the rest are pipeline flags. This is what makes
	// "riscvemu program.s" behave as "riscvemu run program.s".
	if sub == "" {
		if isPositionalRunInvocation(args) {
			cfg, memSize, program, cycles, err := ParseRunArgs(args)
			if err != nil {
				if HelpRequested(err) {
					fmt.Print(FlagHelp)
					return nil
				}
				return err
			}
			return RunScript(core.New(memSize, cfg), program, cycles)
		}
	}
	switch sub {
	case "", "tui", "-tui", "--tui":
		cfg, memSize, err := ConfigFromFlags(rest)
		if err != nil {
			if HelpRequested(err) {
				fmt.Print(FlagHelp)
				return nil
			}
			return err
		}
		return tui.Run(core.New(memSize, cfg))
	case "repl", "-repl", "--repl":
		cfg, memSize, err := ConfigFromFlags(rest)
		if err != nil {
			if HelpRequested(err) {
				fmt.Print(FlagHelp)
				return nil
			}
			return err
		}
		repl, err := NewREPL(core.New(memSize, cfg), cfg)
		if err != nil {
			return fmt.Errorf("start REPL: %w", err)
		}
		repl.Start()
		return nil
	case "run":
		cfg, memSize, program, cycles, err := ParseRunArgs(rest)
		if err != nil {
			if HelpRequested(err) {
				fmt.Print(FlagHelp)
				return nil
			}
			return err
		}
		return RunScript(core.New(memSize, cfg), program, cycles)
	default:
		return fmt.Errorf("unknown subcommand %q (try -help)", sub)
	}
}

// splitSubcommand inspects args and returns the recognised
// subcommand ("tui", "repl", "run", "help") and the
// remaining tokens. If args is empty, or the first token
// does not look like a subcommand, the empty string is
// returned and the entire args slice is the rest. A
// positional program file then falls through to the CLI
// dispatch in Run.
func splitSubcommand(args []string) (string, []string) {
	if len(args) == 0 {
		return "", nil
	}
	switch args[0] {
	case "tui", "repl", "run", "help":
		return args[0], args[1:]
	}
	return "", args
}

// ParseRunArgs is the CLI-side flag parser used by the "run"
// subcommand and by positional-argument invocations
// (e.g. `riscvemu program.s`). It walks args, separating
// -name value pipeline flags from the positional program
// file and an optional cycle count. The first positional
// token is the program file; the second, if present and
// numeric, is the cycle count; anything else is rejected.
// Pipeline flags go through ConfigFromFlags so the same
// -alu-rs, -alu-lat, etc. knobs work everywhere.
func ParseRunArgs(args []string) (cfg cpu.Config, memSize int, program string, cycles int, err error) {
	var flags []string
	for i := 0; i < len(args); i++ {
		a := args[i]
		if len(a) >= 2 && a[0] == '-' && a[1] != '-' {
			flags = append(flags, a)
			// Consume the value if the next token does not
			// look like another flag. -help / -h are value-less
			// and stay single-token.
			if a == "-help" {
				continue
			}
			if i+1 < len(args) {
				flags = append(flags, args[i+1])
				i++
			}
			continue
		}
		if program == "" {
			program = a
			continue
		}
		var n int
		n, err = parseCycles(a)
		if err != nil {
			return cpu.Config{}, 0, "", 0, err
		}
		cycles = n
	}
	c, m, ferr := ConfigFromFlags(flags)
	if ferr != nil {
		return c, m, "", 0, ferr
	}
	return c, m, program, cycles, nil
}

// isPositionalRunInvocation reports whether the args
// should be dispatched to the CLI as a positional-program
// invocation. The signal is simple: at least one token is
// present that does not start with "-" and is not the
// only token (a bare "-" would otherwise look like a
// positional).
func isPositionalRunInvocation(args []string) bool {
	for _, a := range args {
		if len(a) == 0 {
			return true
		}
		if a[0] != '-' {
			return true
		}
	}
	return false
}

// errNoProgram is returned by RunScript when the caller
// invoked the CLI without naming a program file. The error
// message is short enough to print to stderr and is the
// only way to distinguish a missing program from a parse
// failure inside the chosen frontend.
var errNoProgram = errors.New("run: no program file given (try -help)")

// parseCycles converts a string to a non-negative integer
// cycle count. It returns an error for negative values or
// non-numeric input so the CLI can give the user a precise
// failure mode.
func parseCycles(s string) (int, error) {
	n := 0
	for _, r := range s {
		if r < '0' || r > '9' {
			return 0, fmt.Errorf("run: bad cycle count %q (must be a non-negative integer)", s)
		}
		n = n*10 + int(r-'0')
	}
	return n, nil
}
