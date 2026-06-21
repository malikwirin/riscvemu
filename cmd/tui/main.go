// Package main is the entry point for the riscvemu-tui binary.
// It builds on the shared internal/core layer so the TUI and
// the REPL observe the same emulator state through the same
// App API.
package main

import (
	"fmt"
	"os"

	"codeberg.org/malik/riscvemu/cli"
	"codeberg.org/malik/riscvemu/internal/core"
	"codeberg.org/malik/riscvemu/tui"
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
	if err := tui.Run(app); err != nil {
		fmt.Fprintf(os.Stderr, "TUI error: %v\n", err)
		os.Exit(1)
	}
}
