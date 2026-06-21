package main

import (
	"fmt"
	"os"

	"codeberg.org/malik/riscvemu/cli"
	"codeberg.org/malik/riscvemu/internal/core"
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

	app := core.New(memSize, cfg)

	repl, err := cli.NewREPL(app, cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to start REPL: %v\n", err)
		os.Exit(1)
	}

	repl.Start()
}
