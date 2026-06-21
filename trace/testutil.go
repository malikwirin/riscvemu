package trace

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/malikwirin/riscvemu/arch"
)

// Test helpers shared across the trace package's *_test.go
// files. They live in production code (not _test.go suffix) so
// that downstream consumers who write integration tests against
// the trace package can reuse them. The helpers are deliberately
// small: each captures one repeated setup or assertion shape that
// appeared in three or more test files before the consolidation.

// NewTestMachine builds a Machine with a small memory and an
// attached CPU-Memory wiring, ready for a Driver. Tests that need
// a non-default Config should call arch.NewMachineWithConfig
// directly; this helper covers the common "small machine for a
// trace smoke test" case.
func NewTestMachine() *arch.Machine {
	m := arch.NewMachine(1024)
	m.CPU.AttachMemory(m.Memory)
	return m
}

// RunDriver parses src, drives the given machine through the
// resulting trace, and returns the captured output. t.Fatal is
// called on parse or run errors so the test fails fast at the
// point of misuse rather than later in an assert.
func RunDriver(t *testing.T, machine *arch.Machine, src string) *bytes.Buffer {
	t.Helper()
	lines, err := ParseTrace(src)
	if err != nil {
		t.Fatalf("ParseTrace: %v", err)
	}
	var buf bytes.Buffer
	d := NewDriver(machine, &buf)
	if err := d.Run(lines); err != nil {
		t.Fatalf("Run: %v", err)
	}
	return &buf
}

// AssertContainsAll fails the test if any of wants is missing
// from got. The full string is included in the error message
// because the substring-miss is usually easier to diagnose in
// context than in isolation.
func AssertContainsAll(t *testing.T, got string, wants ...string) {
	t.Helper()
	for _, want := range wants {
		if !strings.Contains(got, want) {
			t.Errorf("output missing %q\nfull output:\n%s", want, got)
		}
	}
}

// WriteTempTrace writes content to a fresh tempfile and returns
// the path. The tempfile is cleaned up automatically by t.TempDir
// when the test ends. file_test.go uses this for every happy-path
// and error-path file scenario.
func WriteTempTrace(t *testing.T, name, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	return path
}
