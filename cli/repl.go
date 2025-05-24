package cli

import (
    "fmt"
    "strings"

    "github.com/chzyer/readline"
    "github.com/malikwirin/riscvemu/arch"
)

type REPL struct {
    machine *arch.Machine
    rl      *readline.Instance
	commands map[string]func()
}

func NewREPL(machine *arch.Machine) (*REPL, error) {
    rl, err := readline.New("> ")
    if err != nil {
        return nil, err
    }
	repl := &REPL{
        machine: machine,
        rl:      rl,
		commands: make(map[string]func()),
    }
	repl.initCommands()
	return repl, nil
}

func (r *REPL) initCommands() {
    r.commands = map[string]func(){
        "quit":  r.cmdQuit,
        "exit":  r.cmdQuit,
        "help":  r.cmdHelp,
        "step":  r.cmdStep,
        "reset": r.cmdReset,
        "regs":  r.cmdRegs,
        "pc":    r.cmdPC,
    }
}

func (r *REPL) handleCommand(cmd string) {
    if fn, ok := r.commands[cmd]; ok {
        fn()
    } else {
        fmt.Println("Unknown command. Type 'help' for help.")
    }
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
        if fn, ok := r.commands[cmd]; ok {
            fn()
            if cmd == "quit" || cmd == "exit" {
                break
            }
        } else {
            fmt.Println("Unknown command. Type 'help' for help.")
        }
    }
}

func (r *REPL) cmdQuit() {
    fmt.Println("Goodbye!")
}

func (r *REPL) cmdHelp() {
    fmt.Print("Commands:")
    first := true
    for name := range r.commands {
        if !first {
            fmt.Print(",")
        }
        fmt.Printf(" %s", name)
        first = false
    }
    fmt.Println()
}

func (r *REPL) cmdStep() {
    if err := r.machine.Step(); err != nil {
        fmt.Printf("Error during Step: %v\n", err)
    } else {
        fmt.Println("Step executed.")
    }
}

func (r *REPL) cmdReset() {
    if err := r.machine.Reset(); err != nil {
        fmt.Printf("Error during Reset: %v\n", err)
    } else {
        fmt.Println("CPU and memory reset.")
    }
}

func (r *REPL) cmdRegs() {
    fmt.Println("Registers:")
    for i, v := range r.machine.CPU.Registers {
        fmt.Printf("x%-2d: %d\n", i, v)
    }
}

func (r *REPL) cmdPC() {
    fmt.Printf("PC: %d\n", r.machine.CPU.PC)
}
