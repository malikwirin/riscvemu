package cli

import (
	"errors"
	"fmt"
	"strings"

	"codeberg.org/malik/riscvemu/arch"
	"codeberg.org/malik/riscvemu/arch/cpu"
	"codeberg.org/malik/riscvemu/internal/core"
	"github.com/chzyer/readline"
)

var ErrQuit = errors.New("quit command")

type REPL struct {
	app *core.App
	rl  *readline.Instance
}

// NewREPL builds a REPL that drives the given core.App. The cfg
// is held by the App and surfaced via Cfg() for the 'config' and
// 'regs' commands.
func NewREPL(app *core.App, cfg cpu.Config) (*REPL, error) {
	rl, err := readline.New("> ")
	if err != nil {
		return nil, err
	}
	return &REPL{app: app, rl: rl}, nil
}

// App returns the shared core.App. The command handlers receive
// the REPL through the machineOwner interface, which delegates
// to App().
func (r *REPL) App() *core.App { return r.app }

// Machine returns the raw arch.Machine. Commands that need
// direct memory access (peek, mem, store) use this; production
// commands should go through App() instead.
func (r *REPL) Machine() *arch.Machine { return r.app.Machine() }

// Cfg returns the pipeline configuration. The REPL holds it
// alongside the App so it can render the active settings without
// peeking into internal state.
func (r *REPL) Cfg() cpu.Config { return r.app.Cfg() }

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
