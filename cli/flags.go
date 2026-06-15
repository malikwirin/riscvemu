package cli

import (
	"fmt"
	"strconv"

	"github.com/malikwirin/riscvemu/arch/cpu"
)

// ConfigFromFlags takes a SpecConfig() baseline and applies the
// supplied flag overrides. Unknown flags are reported as errors.
// A flag with the suffix "-help" prints the help text to stdout
// and returns a sentinel error so the caller can exit without
// starting the REPL.
type FlagOverride struct {
	Name  string
	Value int
}

// FlagHelp lists every supported flag with its effect. Printed
// when the user passes -help.
const FlagHelp = `Pipeline configuration flags (override SpecConfig defaults):
  -alu-rs N      ALU reservation station count (default 4)
  -lsu-rs N      Load/Store reservation station count (default 3)
  -mul-rs N      MUL reservation station count (default 1)
  -div-rs N      DIV reservation station count (default 1)
  -alu-lat N     ALU latency in cycles (default 1)
  -load-lat N    Load latency in cycles (default 2)
  -store-lat N   Store latency in cycles (default 2)
  -mul-lat N     MUL latency in cycles (default 3)
  -div-lat N     DIV latency in cycles (default 5)
  -iq N          Instruction queue size (default 8)
  -regs N        Architectural register count (default 8; set 32 for the
                 full RISC-V register file)
  -mem N         Memory size in bytes (default 65536)
  -help          Show this help and exit
`

// ConfigFromFlags parses os.Args-style flag pairs of the form
// "-name value". Returns the resulting Config, the memory size,
// and an error if any flag is unknown or has an invalid value.
// The first return is also printed in help text.
func ConfigFromFlags(args []string) (cpu.Config, int, error) {
	cfg := cpu.SpecConfig()
	memSize := 64 * 1024
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg == "-help" || arg == "--help" {
			fmt.Print(FlagHelp)
			return cfg, memSize, errHelp
		}
		if len(arg) < 2 || arg[0] != '-' {
			return cfg, memSize, fmt.Errorf("expected flag, got %q", arg)
		}
		name := arg[1:]
		if i+1 >= len(args) {
			return cfg, memSize, fmt.Errorf("flag %q requires a value", arg)
		}
		val, err := strconv.Atoi(args[i+1])
		if err != nil {
			return cfg, memSize, fmt.Errorf("flag %q: invalid value %q: %w", arg, args[i+1], err)
		}
		if val < 0 {
			return cfg, memSize, fmt.Errorf("flag %q: value must be non-negative, got %d", arg, val)
		}
		i++ // consume value
		switch name {
		case "alu-rs":
			cfg.ALURSCount = val
		case "lsu-rs":
			cfg.LSURSCount = val
		case "mul-rs":
			cfg.MulRSCount = val
		case "div-rs":
			cfg.DivRSCount = val
		case "alu-lat":
			cfg.ALULatency = val
		case "load-lat":
			cfg.LoadLatency = val
		case "store-lat":
			cfg.StoreLatency = val
		case "mul-lat":
			cfg.MulLatency = val
		case "div-lat":
			cfg.DivLatency = val
		case "iq":
			cfg.InstructionQueueSize = val
		case "regs":
			cfg.RegisterCount = val
		case "mem":
			memSize = val
		default:
			return cfg, memSize, fmt.Errorf("unknown flag: %q (try -help)", arg)
		}
	}
	if err := validateConfig(cfg); err != nil {
		return cfg, memSize, err
	}
	return cfg, memSize, nil
}

// errHelp is a sentinel that signals the user asked for the
// help text. Callers should exit without starting the REPL.
type helpError struct{}

func (helpError) Error() string { return "help requested" }

// errHelp is exported through HelpRequested below.
var errHelp = helpError{}

// HelpRequested returns true if the error came from -help.
func HelpRequested(err error) bool {
	_, ok := err.(helpError)
	return ok
}

// validateConfig checks that the post-override configuration is
// internally consistent. Zero values for the size fields are
// recovered by NewCPU (RegisterCount=0 -> 32) but the RS counts
// and latencies must be at least 1 to avoid divide-by-zero in
// the reservation-station pools.
func validateConfig(cfg cpu.Config) error {
	if cfg.ALURSCount < 1 {
		return fmt.Errorf("alu-rs must be >= 1, got %d", cfg.ALURSCount)
	}
	if cfg.LSURSCount < 1 {
		return fmt.Errorf("lsu-rs must be >= 1, got %d", cfg.LSURSCount)
	}
	if cfg.MulRSCount < 1 {
		return fmt.Errorf("mul-rs must be >= 1, got %d", cfg.MulRSCount)
	}
	if cfg.DivRSCount < 1 {
		return fmt.Errorf("div-rs must be >= 1, got %d", cfg.DivRSCount)
	}
	if cfg.ALULatency < 1 {
		return fmt.Errorf("alu-lat must be >= 1, got %d", cfg.ALULatency)
	}
	if cfg.LoadLatency < 1 {
		return fmt.Errorf("load-lat must be >= 1, got %d", cfg.LoadLatency)
	}
	if cfg.StoreLatency < 1 {
		return fmt.Errorf("store-lat must be >= 1, got %d", cfg.StoreLatency)
	}
	if cfg.MulLatency < 1 {
		return fmt.Errorf("mul-lat must be >= 1, got %d", cfg.MulLatency)
	}
	if cfg.DivLatency < 1 {
		return fmt.Errorf("div-lat must be >= 1, got %d", cfg.DivLatency)
	}
	if cfg.InstructionQueueSize < 1 {
		return fmt.Errorf("iq must be >= 1, got %d", cfg.InstructionQueueSize)
	}
	if cfg.RegisterCount < 0 {
		return fmt.Errorf("regs must be >= 0, got %d", cfg.RegisterCount)
	}
	return nil
}
