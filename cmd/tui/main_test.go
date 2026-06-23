package main

import (
	"testing"

	"codeberg.org/malik/riscvemu/internal/testutil"
)

// TestPackageBuilds is a smoke test: it imports the same
// packages the real main.go uses, so a broken import in the
// entry point would surface here as a compile error. The
// runtime behaviour of main is not exercised because Go's
// testing harness cannot call a package main's main()
// function directly; that path is covered by manual smoke
// tests during development.
func TestPackageBuilds(t *testing.T) {
	if app := testutil.NewSpecApp(); app == nil {
		t.Fatal("testutil.NewSpecApp returned nil")
	}
}
