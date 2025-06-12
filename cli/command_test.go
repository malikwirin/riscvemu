package cli

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/malikwirin/riscvemu/arch"
	"github.com/malikwirin/riscvemu/assembler"
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
	instr, _ := assembler.ParseInstruction("addi x0, x0, 0")
	for i := 0; i < 4; i++ {
		base := i * 4
		m.Memory.Data[base+0] = byte(instr)
		m.Memory.Data[base+1] = byte(instr >> 8)
		m.Memory.Data[base+2] = byte(instr >> 16)
		m.Memory.Data[base+3] = byte(instr >> 24)
	}
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

func TestCmdLoad(t *testing.T) {
	m := arch.NewMachine(64)
	owner := &testOwner{m}

	asmCode := "addi x0, x0, 0\n"
	tmpfile, err := os.CreateTemp("", "testprog-*.asm")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.WriteString(asmCode); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}
	tmpfile.Close()

	out := captureOutput(func() {
		err := cmdLoad(owner, []string{tmpfile.Name()})
		if err != nil {
			t.Errorf("cmdLoad unexpected error: %v", err)
		}
	})

	if !strings.Contains(out, "Program loaded") {
		t.Errorf("cmdLoad output missing 'Program loaded': %q", out)
	}

	instr, _ := assembler.ParseInstruction("addi x0, x0, 0")
	mem := uint32(m.Memory.Data[0]) | uint32(m.Memory.Data[1])<<8 | uint32(m.Memory.Data[2])<<16 | uint32(m.Memory.Data[3])<<24
	if mem != uint32(instr) {
		t.Errorf("Loaded instruction does not match, got %08x, want %08x", mem, instr)
	}
}
