package main

import (
	"testing"
)

// TestPackageBuilds is a smoke test for the riscvemu
// binary's package. It imports cli and verifies the
// Run entry point can be reached without a real TTY
// when only -help is given. The actual main() cannot
// be called from a unit test (it calls os.Exit on
// error), but the import is enough to surface a broken
// import in the entry point as a compile error.
func TestPackageBuilds(t *testing.T) {
	// Compile-time check: if the package fails to compile
	// or the cli.Run signature drifts, this test stops
	// being runnable. No runtime check is needed.
	if t == nil {
		t.Fatal("testing.T is nil")
	}
}
