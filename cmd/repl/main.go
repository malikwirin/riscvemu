// Package main is the entry point for the riscvemu REPL/CLI
// binary. The same core.App that drives this REPL is also used
// by the TUI binary in cmd/tui; both frontends are thin shells
// over the shared internal/core layer.
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
