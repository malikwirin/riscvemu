package cli

import (
	"bytes"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/malikwirin/riscvemu/arch"
	"github.com/malikwirin/riscvemu/assembler"
)

// testOwner is a test double for machineOwner.
type testOwner struct {
	m *arch.Machine
}

func (t *testOwner) Machine() *arch.Machine { return t.m }

// captureOutput runs f and returns everything it prints to os.Stdout.
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

// withMachine runs f with a new machine of given size and returns the machine and testOwner.
func withMachine(memSize int, f func(m *arch.Machine, owner *testOwner)) {
	m := arch.NewMachine(memSize)
	owner := &testOwner{m}
	f(m, owner)
}

// Test helper to check if a value is "randomized", i.e. not zero and plausibly not all same.
func isRandomized(values []uint32) bool {
	allZero := true
	first := values[0]
	allSame := true
	for _, v := range values {
		if v != 0 {
			allZero = false
		}
		if v != first {
			allSame = false
		}
	}
	return !allZero && !allSame
}

func TestCmdRandStore(t *testing.T) {
	withMachine(128, func(m *arch.Machine, owner *testOwner) {
		startAddr := uint32(32)
		count := 5

		// Seed so we get different results on every run (for realism, but see below!)
		rand.Seed(time.Now().UnixNano())

		err := cmdRandStore(owner, []string{fmt.Sprintf("%d", startAddr), fmt.Sprintf("%d", count)})
		assertNoErr(t, err, "cmdRandStore")

		var values []uint32
		for i := 0; i < count; i++ {
			word, err := m.Memory.ReadWord(startAddr + uint32(i)*4)
			assertNoErr(t, err, "ReadWord after randstore")
			values = append(values, word)
		}

		// Check that the values are plausibly random (not all zero, not all the same)
		if !isRandomized(values) {
			t.Errorf("Expected random values at %d..%d, got: %#v", startAddr, startAddr+uint32((count-1)*4), values)
		}

		// Error: not enough args
		err = cmdRandStore(owner, []string{"20"})
		if err == nil || !strings.Contains(err.Error(), "usage") {
			t.Error("Expected usage error for too few args")
		}
		// Error: invalid address
		err = cmdRandStore(owner, []string{"notanaddr", "3"})
		if err == nil || !strings.Contains(err.Error(), "invalid address") {
			t.Error("Expected error for invalid address")
		}
		// Error: invalid count
		err = cmdRandStore(owner, []string{"20", "NaN"})
		if err == nil || !strings.Contains(err.Error(), "invalid count") {
			t.Error("Expected error for invalid count")
		}
		// Error: negative count
		err = cmdRandStore(owner, []string{"20", "-4"})
		if err == nil || !strings.Contains(err.Error(), "count must be positive") {
			t.Error("Expected error for negative count")
		}
	})
}

// assertContains reports an error if want is not a substring of got.
func assertContains(t *testing.T, got, want string, msg string) {
	t.Helper()
	if !strings.Contains(got, want) {
		t.Errorf("%s: got %q, want substring %q", msg, got, want)
	}
}

// assertNoErr fails the test if err is not nil.
func assertNoErr(t *testing.T, err error, context string) {
	t.Helper()
	if err != nil {
		t.Fatalf("%s: unexpected error: %v", context, err)
	}
}

func TestCmdHelp(t *testing.T) {
	out := captureOutput(func() { _ = cmdHelp(nil, nil) })
	assertContains(t, out, "Available commands:", "cmdHelp output missing")

	out = captureOutput(func() { _ = cmdHelp(nil, []string{"step"}) })
	assertContains(t, out, "step", "cmdHelp output for step missing")
	assertContains(t, out, "step", "cmdHelp output for step missing")

	out = captureOutput(func() { _ = cmdHelp(nil, []string{"unknowncmd"}) })
	assertContains(t, out, "Unknown command", "cmdHelp output for unknown command missing")
}

func TestCmdMem(t *testing.T) {
	withMachine(64, func(m *arch.Machine, owner *testOwner) {
		_ = m.Memory.WriteWord(0, 0xDEADBEEF)
		_ = m.Memory.WriteWord(4, 0x12345678)
		_ = m.Memory.WriteWord(8, 0x0BADBEEF)

		out := captureOutput(func() {
			err := cmdMem(owner, []string{"0", "3"})
			assertNoErr(t, err, "cmdMem")
		})
		assertContains(t, out, "0x00000000: 0xdeadbeef", "cmdMem output missing first word")
		assertContains(t, out, "0x00000004: 0x12345678", "cmdMem output missing second word")
		assertContains(t, out, "0x00000008: 0x0badbeef", "cmdMem output missing third word")

		out = captureOutput(func() {
			err := cmdMem(owner, []string{"60", "2"})
			assertNoErr(t, err, "cmdMem OOB")
		})
		assertContains(t, out, "ERROR", "cmdMem should print ERROR for out-of-bounds access")
	})
}

func TestCmdPC(t *testing.T) {
	withMachine(64, func(m *arch.Machine, owner *testOwner) {
		m.CPU.PC = 1234
		out := captureOutput(func() { _ = cmdPC(owner, nil) })
		assertContains(t, out, "PC: 1234", "cmdPC output missing correct PC")
	})
}

func TestCmdStep(t *testing.T) {
	withMachine(64, func(m *arch.Machine, owner *testOwner) {
		instr, _ := assembler.ParseInstruction("addi x0, x0, 0")
		for i := 0; i < 4; i++ {
			base := i * 4
			m.Memory.Data[base+0] = byte(instr)
			m.Memory.Data[base+1] = byte(instr >> 8)
			m.Memory.Data[base+2] = byte(instr >> 16)
			m.Memory.Data[base+3] = byte(instr >> 24)
		}

		out := captureOutput(func() {
			err := cmdStep(owner, nil)
			assertNoErr(t, err, "cmdStep (default)")
		})
		assertContains(t, out, "Executed 1 step", "cmdStep output for default step missing")

		out = captureOutput(func() {
			err := cmdStep(owner, []string{"3"})
			assertNoErr(t, err, "cmdStep (3 steps)")
		})
		assertContains(t, out, "Executed 3 step", "cmdStep output for 3 steps missing")

		err := cmdStep(owner, []string{"NaN"})
		if err == nil || !strings.Contains(err.Error(), "invalid step count") {
			t.Errorf("cmdStep should fail for invalid input, got: %v", err)
		}
	})
}

func TestCmdRegs(t *testing.T) {
	withMachine(64, func(m *arch.Machine, owner *testOwner) {
		m.CPU.Reg[0] = 42
		m.CPU.Reg[31] = 99
		out := captureOutput(func() { _ = cmdRegs(owner, nil) })
		assertContains(t, out, "x0", "cmdRegs output missing register labels (x0)")
		assertContains(t, out, "x31", "cmdRegs output missing register labels (x31)")
		assertContains(t, out, "42", "cmdRegs output missing register value 42")
		assertContains(t, out, "99", "cmdRegs output missing register value 99")
	})
}

func TestCmdReset(t *testing.T) {
	withMachine(64, func(m *arch.Machine, owner *testOwner) {
		m.CPU.PC = 123
		out := captureOutput(func() {
			err := cmdReset(owner, nil)
			assertNoErr(t, err, "cmdReset")
		})
		assertContains(t, out, "CPU and memory reset", "cmdReset output missing")
	})
}

func TestCmdLoad(t *testing.T) {
	withMachine(64, func(m *arch.Machine, owner *testOwner) {
		asmCode := "addi x0, x0, 0\n"
		tmpfile, err := os.CreateTemp("", "testprog-*.asm")
		assertNoErr(t, err, "create temp file")
		defer os.Remove(tmpfile.Name())

		_, err = tmpfile.WriteString(asmCode)
		assertNoErr(t, err, "write temp file")
		tmpfile.Close()

		out := captureOutput(func() {
			err := cmdLoad(owner, []string{tmpfile.Name()})
			assertNoErr(t, err, "cmdLoad")
		})
		assertContains(t, out, "Program loaded", "cmdLoad output missing 'Program loaded'")

		instr, _ := assembler.ParseInstruction("addi x0, x0, 0")
		mem := uint32(m.Memory.Data[0]) | uint32(m.Memory.Data[1])<<8 | uint32(m.Memory.Data[2])<<16 | uint32(m.Memory.Data[3])<<24
		if mem != uint32(instr) {
			t.Errorf("Loaded instruction does not match, got %08x, want %08x", mem, instr)
		}
	})
}

func TestCmdPeek(t *testing.T) {
	withMachine(64, func(m *arch.Machine, owner *testOwner) {
		_ = m.Memory.WriteWord(0, 0xDEADBEEF)
		m.CPU.PC = 0

		out := captureOutput(func() {
			err := cmdPeek(owner, nil)
			assertNoErr(t, err, "cmdPeek")
		})
		assertContains(t, out, "Next instruction at 0x00000000: 0xdeadbeef", "cmdPeek output missing or incorrect")

		m.CPU.PC = 1000 // outside allocated memory
		out = captureOutput(func() {
			_ = cmdPeek(owner, nil)
		})
		assertContains(t, out, "Error reading memory", "cmdPeek should print error when PC is out of bounds")
	})
}

func TestCmdStore(t *testing.T) {
	withMachine(128, func(m *arch.Machine, owner *testOwner) {
		err := cmdStore(owner, []string{"100", "42"})
		assertNoErr(t, err, "cmdStore single")

		word, err := m.Memory.ReadWord(100)
		assertNoErr(t, err, "ReadWord after store single")
		if word != 42 {
			t.Errorf("Expected memory at 100 to be 42, got %d", word)
		}

		err = cmdStore(owner, []string{"104", "1", "2", "3"})
		assertNoErr(t, err, "cmdStore multiple")
		for i, want := range []uint32{1, 2, 3} {
			word, err := m.Memory.ReadWord(104 + uint32(i)*4)
			assertNoErr(t, err, "ReadWord after store multiple")
			if word != want {
				t.Errorf("Expected memory at %d to be %d, got %d", 104+uint32(i)*4, want, word)
			}
		}

		err = cmdStore(owner, []string{})
		if err == nil || !strings.Contains(err.Error(), "usage") {
			t.Error("Expected usage error for no args")
		}

		err = cmdStore(owner, []string{"notanaddr", "1"})
		if err == nil || !strings.Contains(err.Error(), "invalid address") {
			t.Error("Expected error for invalid address")
		}

		err = cmdStore(owner, []string{"100", "nope"})
		if err == nil || !strings.Contains(err.Error(), "invalid value") {
			t.Error("Expected error for invalid value")
		}
	})
}
