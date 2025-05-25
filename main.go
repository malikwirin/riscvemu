package main

import (
	"fmt"
	"os"

	"github.com/malikwirin/riscvemu/arch"
	"github.com/malikwirin/riscvemu/cli"
)

func main() {
	machine := arch.NewMachine(64 * 1024)

	repl, err := cli.NewREPL(machine)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to start REPL: %v\n", err)
		os.Exit(1)
	}

	repl.Start()
}
