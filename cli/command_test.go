package cli

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/malikwirin/riscvemu/arch"
)

type testOwner struct {
	m *arch.Machine
}

func (t *testOwner) Machine() *arch.Machine { return t.m }

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

func TestCmdQuit(t *testing.T) {
	out := captureOutput(func() {
		err := cmdQuit(nil, nil)
		if err != nil {
			t.Errorf("cmdQuit returned unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "Goodbye!") {
		t.Errorf("cmdQuit output missing 'Goodbye!': %q", out)
	}
}

func TestCmdHelp(t *testing.T) {
	out := captureOutput(func() {
		cmdHelp(nil, nil)
	})
	if !strings.Contains(out, "Available commands:") {
		t.Errorf("cmdHelp output missing 'Available commands': %q", out)
	}
	// Test help for valid command
	out = captureOutput(func() {
		cmdHelp(nil, []string{"step"})
	})
	if !strings.Contains(out, "step") || !strings.Contains(out, "Execute n steps") {
		t.Errorf("cmdHelp output for step missing: %q", out)
	}
	// Test help for unknown command
	out = captureOutput(func() {
		cmdHelp(nil, []string{"unknowncmd"})
	})
	if !strings.Contains(out, "Unknown command") {
		t.Errorf("cmdHelp output for unknown command missing: %q", out)
	}
}

func TestCmdPC(t *testing.T) {
	m := arch.NewMachine(64)
	m.CPU.PC = 1234
	owner := &testOwner{m}
	out := captureOutput(func() {
		cmdPC(owner, nil)
	})
	if !strings.Contains(out, "PC: 1234") {
		t.Errorf("cmdPC output missing correct PC: %q", out)
	}
}

func TestCmdStep(t *testing.T) {
	m := arch.NewMachine(64)
	owner := &testOwner{m}
	// Step n=1 (default)
	out := captureOutput(func() {
		err := cmdStep(owner, nil)
		if err != nil {
			t.Errorf("cmdStep unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "Executed 1 step") {
		t.Errorf("cmdStep output for default step missing: %q", out)
	}

	// Step n=3
	out = captureOutput(func() {
		err := cmdStep(owner, []string{"3"})
		if err != nil {
			t.Errorf("cmdStep unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "Executed 3 step") {
		t.Errorf("cmdStep output for 3 steps missing: %q", out)
	}

	// Invalid argument
	err := cmdStep(owner, []string{"NaN"})
	if err == nil || !strings.Contains(err.Error(), "invalid step count") {
		t.Errorf("cmdStep should fail for invalid input, got: %v", err)
	}
}

func TestCmdRegs(t *testing.T) {
	m := arch.NewMachine(64)
	m.CPU.Reg[0] = 42
	m.CPU.Reg[31] = 99
	owner := &testOwner{m}
	out := captureOutput(func() {
		cmdRegs(owner, nil)
	})
	if !strings.Contains(out, "x0") || !strings.Contains(out, "x31") {
		t.Errorf("cmdRegs output missing register labels: %q", out)
	}
	if !strings.Contains(out, "42") || !strings.Contains(out, "99") {
		t.Errorf("cmdRegs output missing register values: %q", out)
	}
}

func TestCmdReset(t *testing.T) {
	m := arch.NewMachine(64)
	m.CPU.PC = 123
	owner := &testOwner{m}
	out := captureOutput(func() {
		err := cmdReset(owner, nil)
		if err != nil {
			t.Errorf("cmdReset unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "CPU and memory reset") {
		t.Errorf("cmdReset output missing: %q", out)
	}
}
