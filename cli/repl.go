package cli

import (
	"errors"
	"fmt"
	"strings"

	"github.com/chzyer/readline"
	"github.com/malikwirin/riscvemu/arch"
	"github.com/malikwirin/riscvemu/arch/cpu"
)

var ErrQuit = errors.New("quit command")

type REPL struct {
	machine *arch.Machine
	cfg     cpu.Config
	rl      *readline.Instance
}

func NewREPL(machine *arch.Machine, cfg cpu.Config) (*REPL, error) {
	rl, err := readline.New("> ")
	if err != nil {
		return nil, err
	}
	repl := &REPL{
		machine: machine,
		cfg:     cfg,
		rl:      rl,
	}
	return repl, nil
}

func (r *REPL) Machine() *arch.Machine {
	return r.machine
}

// Cfg returns the configuration that was used to build the
// machine. The REPL needs it for commands that iterate over the
// register file (cmdRegs) and for displaying the active
// configuration (cmdConfig).
func (r *REPL) Cfg() cpu.Config {
	return r.cfg
}

func (r *REPL) Start() {
	defer r.rl.Close()
	fmt.Println("Simple CPU REPL. Type 'step', 'reset', 'quit' or 'help'.")

	for {
		line, err := r.rl.Readline()
		if err != nil {
			fmt.Println("Goodbye!")
			break
		}
		if strings.TrimSpace(line) == "" {
			continue
		}
		tokens := strings.Fields(line)
		cmdName := tokens[0]
		args := tokens[1:]

		cmd, ok := commands[cmdName]
		if !ok {
			fmt.Println("Unknown command. Type 'help' for help.")
			continue
		}
		err = cmd.Handler(r, args)
		if errors.Is(err, ErrQuit) {
			fmt.Println("Goodbye!")
			break
		} else if err != nil {
			fmt.Printf("Error: %v\n", err)
		}
	}
}
