// Package main is the entry point for the riscvemu binary. All
// dispatch logic (subcommand parsing, flag handling, frontend
// selection) lives in the cli package so the same code path
// can be tested without going through package main. This file
// only translates a non-nil error from cli.Run into a
// non-zero process exit and a human-readable message.
package main

import (
	"fmt"
	"os"

	"codeberg.org/malik/riscvemu/cli"
)

func main() {
	if err := cli.Run(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
