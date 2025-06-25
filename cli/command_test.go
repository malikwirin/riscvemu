package cli

import (
	"fmt"
	"math/rand"
	"os"
	"testing"
	"time"

	"github.com/malikwirin/riscvemu/arch"
	"github.com/malikwirin/riscvemu/assembler"
	"github.com/stretchr/testify/assert"
)

func TestCmdRandStore(t *testing.T) {
	withMachine(128, func(m *arch.Machine, owner *testOwner) {
		startAddr := uint32(32)
		count := 5

		// Seed so we get different results on every run (for realism, but see below!)
		rand.Seed(time.Now().UnixNano())

		err := cmdRandStore(owner, []string{fmt.Sprintf("%d", startAddr), fmt.Sprintf("%d", count)})
		assert.NoError(t, err, "cmdRandStore")

		var values []uint32
		for i := 0; i < count; i++ {
			word, err := m.Memory.ReadWord(startAddr + uint32(i)*4)
			assert.NoError(t, err, "ReadWord after randstore")
			values = append(values, word)
		}

		// Check that the values are plausibly random (not all zero, not all the same)
		assert.Truef(t, isRandomized(values), "Expected random values at %d..%d, got: %#v", startAddr, startAddr+uint32((count-1)*4), values)

		// Error: not enough args
		err = cmdRandStore(owner, []string{"20"})
		assert.Error(t, err, "Expected usage error for too few args")
		assert.Contains(t, err.Error(), "usage", "Expected usage error for too few args")

		// Error: invalid address
		err = cmdRandStore(owner, []string{"notanaddr", "3"})
		assert.Error(t, err, "Expected error for invalid address")
		assert.Contains(t, err.Error(), "invalid address", "Expected error for invalid address")

		// Error: invalid count
		err = cmdRandStore(owner, []string{"20", "NaN"})
		assert.Error(t, err, "Expected error for invalid count")
		assert.Contains(t, err.Error(), "invalid count", "Expected error for invalid count")

		// Error: negative count
		err = cmdRandStore(owner, []string{"20", "-4"})
		assert.Error(t, err, "Expected error for negative count")
		assert.Contains(t, err.Error(), "count must be positive", "Expected error for negative count")
	})
}

func TestCmdHelp(t *testing.T) {
	out := captureOutput(func() { _ = cmdHelp(nil, nil) })
	assert.Contains(t, out, "Available commands:", "cmdHelp output missing")

	out = captureOutput(func() { _ = cmdHelp(nil, []string{"step"}) })
	assert.Contains(t, out, "step", "cmdHelp output for step missing")

	out = captureOutput(func() { _ = cmdHelp(nil, []string{"unknowncmd"}) })
	assert.Contains(t, out, "Unknown command", "cmdHelp output for unknown command missing")
}

func TestCmdMem(t *testing.T) {
	withMachine(64, func(m *arch.Machine, owner *testOwner) {
		_ = m.Memory.WriteWord(0, 0xDEADBEEF)
		_ = m.Memory.WriteWord(4, 0x12345678)
		_ = m.Memory.WriteWord(8, 0x0BADBEEF)

		out := captureOutput(func() {
			err := cmdMem(owner, []string{"0", "3"})
			assert.NoError(t, err, "cmdMem")
		})
		assert.Contains(t, out, "0x00000000: 0xdeadbeef", "cmdMem output missing first word")
		assert.Contains(t, out, "0x00000004: 0x12345678", "cmdMem output missing second word")
		assert.Contains(t, out, "0x00000008: 0x0badbeef", "cmdMem output missing third word")

		out = captureOutput(func() {
			err := cmdMem(owner, []string{"60", "2"})
			assert.NoError(t, err, "cmdMem OOB")
		})
		assert.Contains(t, out, "ERROR", "cmdMem should print ERROR for out-of-bounds access")
	})
}

func TestCmdPC(t *testing.T) {
	withMachine(64, func(m *arch.Machine, owner *testOwner) {
		m.CPU.PC = 1234
		out := captureOutput(func() { _ = cmdPC(owner, nil) })
		assert.Contains(t, out, "PC: 1234", "cmdPC output missing correct PC")
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
			assert.NoError(t, err, "cmdStep (default)")
		})
		assert.Contains(t, out, "Executed 1 step", "cmdStep output for default step missing")

		out = captureOutput(func() {
			err := cmdStep(owner, []string{"3"})
			assert.NoError(t, err, "cmdStep (3 steps)")
		})
		assert.Contains(t, out, "Executed 3 step", "cmdStep output for 3 steps missing")

		err := cmdStep(owner, []string{"NaN"})
		assert.Error(t, err, "cmdStep should fail for invalid input")
		assert.Contains(t, err.Error(), "invalid step count", "cmdStep should fail for invalid input")
	})
}

func TestCmdRegs(t *testing.T) {
	withMachine(64, func(m *arch.Machine, owner *testOwner) {
		m.CPU.Reg[0] = 42
		m.CPU.Reg[31] = 99
		out := captureOutput(func() { _ = cmdRegs(owner, nil) })
		assert.Contains(t, out, "x0", "cmdRegs output missing register labels (x0)")
		assert.Contains(t, out, "x31", "cmdRegs output missing register labels (x31)")
		assert.Contains(t, out, "42", "cmdRegs output missing register value 42")
		assert.Contains(t, out, "99", "cmdRegs output missing register value 99")
	})
}

func TestCmdReset(t *testing.T) {
	withMachine(64, func(m *arch.Machine, owner *testOwner) {
		m.CPU.PC = 123
		out := captureOutput(func() {
			err := cmdReset(owner, nil)
			assert.NoError(t, err, "cmdReset")
		})
		assert.Contains(t, out, "CPU and memory reset", "cmdReset output missing")
	})
}

func TestCmdLoad(t *testing.T) {
	withMachine(64, func(m *arch.Machine, owner *testOwner) {
		asmCode := "addi x0, x0, 0\n"
		tmpfile, err := os.CreateTemp("", "testprog-*.asm")
		assert.NoError(t, err, "create temp file")
		defer os.Remove(tmpfile.Name())

		_, err = tmpfile.WriteString(asmCode)
		assert.NoError(t, err, "write temp file")
		tmpfile.Close()

		out := captureOutput(func() {
			err := cmdLoad(owner, []string{tmpfile.Name()})
			assert.NoError(t, err, "cmdLoad")
		})
		assert.Contains(t, out, "Program loaded", "cmdLoad output missing 'Program loaded'")

		instr, _ := assembler.ParseInstruction("addi x0, x0, 0")
		mem := uint32(m.Memory.Data[0]) | uint32(m.Memory.Data[1])<<8 | uint32(m.Memory.Data[2])<<16 | uint32(m.Memory.Data[3])<<24
		assert.Equalf(t, uint32(instr), mem, "Loaded instruction does not match, got %08x, want %08x", mem, instr)
	})
}

func TestCmdPeek(t *testing.T) {
	withMachine(64, func(m *arch.Machine, owner *testOwner) {
		_ = m.Memory.WriteWord(0, 0xDEADBEEF)
		m.CPU.PC = 0

		out := captureOutput(func() {
			err := cmdPeek(owner, nil)
			assert.NoError(t, err, "cmdPeek")
		})
		assert.Contains(t, out, "Next instruction at 0x00000000: 0xdeadbeef", "cmdPeek output missing or incorrect")

		m.CPU.PC = 1000 // outside allocated memory
		out = captureOutput(func() {
			_ = cmdPeek(owner, nil)
		})
		assert.Contains(t, out, "Error reading memory", "cmdPeek should print error when PC is out of bounds")
	})
}

func TestCmdStore(t *testing.T) {
	withMachine(128, func(m *arch.Machine, owner *testOwner) {
		err := cmdStore(owner, []string{"100", "42"})
		assert.NoError(t, err, "cmdStore single")

		word, err := m.Memory.ReadWord(100)
		assert.NoError(t, err, "ReadWord after store single")
		assert.Equal(t, uint32(42), word, "Expected memory at 100 to be 42")

		err = cmdStore(owner, []string{"104", "1", "2", "3"})
		assert.NoError(t, err, "cmdStore multiple")
		for i, want := range []uint32{1, 2, 3} {
			word, err := m.Memory.ReadWord(104 + uint32(i)*4)
			assert.NoError(t, err, "ReadWord after store multiple")
			assert.Equalf(t, want, word, "Expected memory at %d to be %d, got %d", 104+uint32(i)*4, want, word)
		}

		err = cmdStore(owner, []string{})
		assert.Error(t, err, "Expected usage error for no args")
		assert.Contains(t, err.Error(), "usage", "Expected usage error for no args")

		err = cmdStore(owner, []string{"notanaddr", "1"})
		assert.Error(t, err, "Expected error for invalid address")
		assert.Contains(t, err.Error(), "invalid address", "Expected error for invalid address")

		err = cmdStore(owner, []string{"100", "nope"})
		assert.Error(t, err, "Expected error for invalid value")
		assert.Contains(t, err.Error(), "invalid value", "Expected error for invalid value")
	})
}
