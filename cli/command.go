package cli

import (
	"fmt"
	"github.com/malikwirin/riscvemu/arch"
	"github.com/malikwirin/riscvemu/assembler"
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
		"quit": {
			Handler: cmdQuit,
			Help:    "quit: Exit the CLI",
		},
		"help": {
			Handler: cmdHelp,
			Help:    "help [command]: Show help for a command",
		},
		"load": {
			Handler: cmdLoad,
			Help:    "load <filename> [address]: Load a binary program into memory at an optional address (default 0)",
		},
		"mem": {Handler: cmdMem, Help: "mem [start [length]]: Dump memory (default: start=0, length=16 words)"},
		"pc": {
			Handler: cmdPC,
			Help:    "pc: Print the current program counter",
		},
		"peek": {
			Handler: cmdPeek,
			Help:    "peek: Show the next instruction at the current PC",
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
	return ErrQuit
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
	if len(args) < 1 {
		return fmt.Errorf("usage: load <filename> [address]")
	}

	filename := args[0]
	address := uint32(0)

	if len(args) > 1 {
		addr, err := strconv.ParseUint(args[1], 0, 32)
		if err != nil {
			return fmt.Errorf("invalid address: %q", args[1])
		}
		address = uint32(addr)
	}

	prog, err := assembler.AssembleFile(filename)
	if err != nil {
		fmt.Printf("Failed to assemble: %v\n", err)
		return err
	}

	m := owner.Machine()
	if err := m.LoadProgram(prog, address); err != nil {
		fmt.Printf("Failed to load program: %v\n", err)
		return err
	}

	fmt.Println("Program loaded")
	return nil
}

func cmdMem(owner machineOwner, args []string) error {
	start := uint32(0)
	length := 16 // number of words (4 bytes each)

	// Parse optional arguments: start and length
	if len(args) >= 1 {
		var s int
		_, err := fmt.Sscanf(args[0], "%d", &s)
		if err == nil && s >= 0 {
			start = uint32(s)
		}
	}
	if len(args) >= 2 {
		var l int
		_, err := fmt.Sscanf(args[1], "%d", &l)
		if err == nil && l > 0 {
			length = l
		}
	}

	// Dump memory
	for i := 0; i < length; i++ {
		addr := start + uint32(i*4)
		word, err := owner.Machine().Memory.ReadWord(addr)
		if err != nil {
			fmt.Printf("0x%08x: ERROR (%v)\n", addr, err)
		} else {
			fmt.Printf("0x%08x: 0x%08x\n", addr, word)
		}
	}
	return nil
}

func cmdPC(owner machineOwner, _ []string) error {
	fmt.Printf("PC: %d\n", owner.Machine().CPU.PC)
	return nil
}

// cmdPeek prints the next instruction at the current PC as a hex value.
func cmdPeek(owner machineOwner, args []string) error {
	m := owner.Machine()
	pc := m.CPU.PC
	word, err := m.Memory.ReadWord(pc)
	if err != nil {
		fmt.Printf("Error reading memory at 0x%08x: %v\n", pc, err)
		return err
	}
	fmt.Printf("Next instruction at 0x%08x: 0x%08x\n", pc, word)
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
