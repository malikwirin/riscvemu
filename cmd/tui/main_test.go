package main

import (
	"testing"

	"codeberg.org/malik/riscvemu/arch/cpu"
	"codeberg.org/malik/riscvemu/internal/core"
)

// TestPackageBuilds is a smoke test: it imports the same
// packages the real main.go uses, so a broken import in the
// entry point would surface here as a compile error. The
// runtime behaviour of main is not exercised because Go's
// testing harness cannot call a package main's main()
// function directly; that path is covered by manual smoke
// tests during development.
func TestPackageBuilds(t *testing.T) {
	app := core.New(1024, cpu.SpecConfig())
	if app == nil {
		t.Fatal("core.New returned nil")
	}
}
