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
}

func NewREPL(machine *arch.Machine) (*REPL, error) {
    rl, err := readline.New("> ")
    if err != nil {
        return nil, err
    }
    return &REPL{
        machine: machine,
        rl:      rl,
    }, nil
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
        r.handleCommand(cmd)
        if cmd == "quit" || cmd == "exit" {
            break
        }
    }
}

func (r *REPL) handleCommand(cmd string) {
    switch cmd {
    case "quit", "exit":
        fmt.Println("Goodbye!")
    case "help":
        fmt.Println("Commands: step, reset, regs, pc, quit, help")
    case "step":
        if err := r.machine.Step(); err != nil {
            fmt.Printf("Error during Step: %v\n", err)
        } else {
            fmt.Println("Step executed.")
        }
    case "reset":
        if err := r.machine.Reset(); err != nil {
            fmt.Printf("Error during Reset: %v\n", err)
        } else {
            fmt.Println("CPU and memory reset.")
        }
    case "regs":
        fmt.Println("Registers:")
        for i, v := range r.machine.CPU.Registers {
            fmt.Printf("x%-2d: %d\n", i, v)
        }
    case "pc":
        fmt.Printf("PC: %d\n", r.machine.CPU.PC)
    case "":
        // do nothing on empty line
    default:
        fmt.Println("Unknown command. Type 'help' for help.")
    }
}
