package main

import (
	"fmt"
	"os"

	"codeberg.org/malik/riscvemu/arch"
	"codeberg.org/malik/riscvemu/cli"
)

func main() {
	cfg, memSize, err := cli.ConfigFromFlags(os.Args[1:])
	if err != nil {
		if cli.HelpRequested(err) {
			return
		}
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	machine := arch.NewMachineWithConfig(memSize, cfg)

	repl, err := cli.NewREPL(machine, cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to start REPL: %v\n", err)
		os.Exit(1)
	}

	repl.Start()
}
