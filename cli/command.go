package cli

import (
	"fmt"
	"github.com/malikwirin/riscvemu/arch"
	"strconv"
)

type Command struct {
	Handler func(owner machineOwner, args []string) error
	Help    string
}

var commands map[string]Command

func init() {
	commands = map[string]Command{
		"exit": {
			Handler: cmdQuit,
			Help:    "exit: Exit the CLI",
		},
		"help": {
			Handler: cmdHelp,
			Help:    "help [command]: Show help for a command",
		},
		"load": {
			Handler: cmdLoad,
			Help:    "load <filename> [address]: Load a binary program into memory at an optional address (default 0)",
		},
		"pc": {
			Handler: cmdPC,
			Help:    "pc: Print the current program counter",
		},
		"step": {
			Handler: cmdStep,
			Help:    "step [n]: Execute n steps (default 1)",
		},
		"regs": {
			Handler: cmdRegs,
			Help:    "regs: Print the current state of the registers",
		},
		"reset": {
			Handler: cmdReset,
			Help:    "reset: Reset the CPU and memory to initial state",
		},
	}
}

type machineOwner interface {
	Machine() *arch.Machine
}

func cmdQuit(_ machineOwner, _ []string) error {
	fmt.Println("Goodbye!")
	return nil
}

func cmdHelp(_ machineOwner, args []string) error {
	if len(args) == 0 {
		fmt.Println("Available commands:")
		for name := range commands {
			fmt.Printf("  %s\n", name)
		}
		fmt.Println("Type 'help <command>' for details.")
		return nil
	}
	cmdName := args[0]
	cmd, ok := commands[cmdName]
	if !ok {
		fmt.Printf("Unknown command: %s\n", cmdName)
		return nil
	}
	fmt.Printf("Help for '%s':\n  %s\n", cmdName, cmd.Help)
	return nil
}

func cmdLoad(owner machineOwner, args []string) error {
	return nil
}

func cmdPC(owner machineOwner, _ []string) error {
	fmt.Printf("PC: %d\n", owner.Machine().CPU.PC)
	return nil
}

func cmdStep(owner machineOwner, args []string) error {
	n := 1
	if len(args) > 0 {
		parsed, err := strconv.Atoi(args[0])
		if err != nil || parsed < 1 {
			return fmt.Errorf("invalid step count: %q", args[0])
		}
		n = parsed
	}
	m := owner.Machine()
	for i := 0; i < n; i++ {
		if err := m.Step(); err != nil {
			return fmt.Errorf("error during Step %d: %w", i+1, err)
		}
	}
	fmt.Printf("Executed %d step(s).\n", n)
	return nil
}

func cmdRegs(owner machineOwner, _ []string) error {
	m := owner.Machine()
	fmt.Println("Registers:")
	for i, v := range m.CPU.Reg {
		fmt.Printf("x%-2d: %d\n", i, v)
	}
	return nil
}

func cmdReset(owner machineOwner, _ []string) error {
	m := owner.Machine()
	if err := m.Reset(); err != nil {
		return fmt.Errorf("error during Reset: %w", err)
	}
	fmt.Println("CPU and memory reset.")
	return nil
}
