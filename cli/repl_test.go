package cli

import (
    "github.com/malikwirin/riscvemu/arch"
    "strings"
    "testing"
    "bytes"
    "os"
)

func TestREPLHandleCommand(t *testing.T) {
    m := arch.NewMachine(64)
    repl := &REPL{machine: m}

    tests := []struct {
        cmd      string
        expected string
    }{
        {"help", "Commands"},
        {"step", "Step executed."},
        {"reset", "CPU and memory reset."},
        {"regs", "Registers:"},
        {"pc", "PC:"},
        {"foobar", "Unknown command"},
    }

    for _, tc := range tests {
        output := captureOutput(func() {
            repl.handleCommand(tc.cmd)
        })
        if !strings.Contains(output, tc.expected) {
            t.Errorf("For '%s' expected output to contain '%s', got: %s", tc.cmd, tc.expected, output)
        }
    }
}

// captureOutput captures stdout for testing.
func captureOutput(f func()) string {
    old := os.Stdout
    r, w, _ := os.Pipe()
    os.Stdout = w

    f()

    w.Close()
    var buf bytes.Buffer
    _, _ = buf.ReadFrom(r)
    os.Stdout = old
    return buf.String()
}
