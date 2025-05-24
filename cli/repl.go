package cli

import (
    "fmt"
    "strings"
	"errors"

    "github.com/chzyer/readline"
    "github.com/malikwirin/riscvemu/arch"
)

var ErrQuit = errors.New("quit command")

type REPL struct {
    machine *arch.Machine
    rl      *readline.Instance
}

func NewREPL(machine *arch.Machine) (*REPL, error) {
    rl, err := readline.New("> ")
    if err != nil {
        return nil, err
    }
	repl := &REPL{
        machine: machine,
        rl:      rl,
    }
	return repl, nil
}

func (r *REPL) Machine() *arch.Machine {
    return r.machine
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
        cmd := strings.TrimSpace(line)
        if cmd == "" {
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
