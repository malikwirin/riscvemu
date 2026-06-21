package arch

import (
	"testing"

	"codeberg.org/malik/riscvemu/assembler"
	"github.com/stretchr/testify/assert"
)

func encodeCPU(t *testing.T, line string) uint32 {
	t.Helper()
	instr, err := assembler.ParseInstruction(line)
	if err != nil {
		t.Fatalf("ParseInstruction(%q): %v", line, err)
	}
	return uint32(instr)
}

func TestMachineBranchTakenUpdatesPC(t *testing.T) {
	m := NewMachine(64)
	// PC=0:  addi x1, x0, 5
	// PC=4:  addi x2, x0, 5
	// PC=8:  beq x1, x2, 12   -> target = 8 + 12 = 20
	// PC=12: addi x3, x0, 99  (skipped when branch taken)
	// PC=16: addi x4, x0, 1
	// PC=20: addi x5, x0, 42  <-- we expect to land here
	assert.NoError(t, m.Memory.WriteWord(0, encodeCPU(t, "addi x1, x0, 5")))
	assert.NoError(t, m.Memory.WriteWord(4, encodeCPU(t, "addi x2, x0, 5")))
	assert.NoError(t, m.Memory.WriteWord(8, encodeCPU(t, "beq x1, x2, 12")))
	assert.NoError(t, m.Memory.WriteWord(12, encodeCPU(t, "addi x3, x0, 99")))
	assert.NoError(t, m.Memory.WriteWord(16, encodeCPU(t, "addi x4, x0, 1")))
	assert.NoError(t, m.Memory.WriteWord(20, encodeCPU(t, "addi x5, x0, 42")))

	for i := 0; i < 50; i++ {
		if err := m.Step(); err != nil {
			t.Fatalf("Step %d (PC=0x%x): %v", i, m.PC, err)
		}
		t.Logf("step %d PC=0x%x x1=%d x2=%d x3=%d x4=%d x5=%d lastBr=%+v",
			i, m.PC, m.CPU.Reg(1), m.CPU.Reg(2), m.CPU.Reg(3), m.CPU.Reg(4), m.CPU.Reg(5), m.CPU.LastBranch())
		if m.CPU.Reg(5) == 42 {
			assert.Equal(t, uint32(0), m.CPU.Reg(3), "x3 should be 0 (branch skipped the addi)")
			return
		}
	}
	t.Fatalf("branch did not reach target; final PC=0x%x, x1=%d x2=%d x3=%d x4=%d x5=%d",
		m.PC, m.CPU.Reg(1), m.CPU.Reg(2), m.CPU.Reg(3), m.CPU.Reg(4), m.CPU.Reg(5))
}

func TestMachineBranchNotTakenFallsThrough(t *testing.T) {
	m := NewMachine(64)
	// PC=0:  addi x1, x0, 5
	// PC=4:  addi x2, x0, 7   (different from x1)
	// PC=8:  beq x1, x2, 12   -> not taken
	// PC=12: addi x3, x0, 99  <-- we expect to land here
	assert.NoError(t, m.Memory.WriteWord(0, encodeCPU(t, "addi x1, x0, 5")))
	assert.NoError(t, m.Memory.WriteWord(4, encodeCPU(t, "addi x2, x0, 7")))
	assert.NoError(t, m.Memory.WriteWord(8, encodeCPU(t, "beq x1, x2, 12")))
	assert.NoError(t, m.Memory.WriteWord(12, encodeCPU(t, "addi x3, x0, 99")))

	for i := 0; i < 50; i++ {
		if err := m.Step(); err != nil {
			t.Fatalf("Step %d: %v", i, err)
		}
		if m.CPU.Reg(3) == 99 {
			return
		}
	}
	t.Fatalf("PC=0x%x, x1=%d x2=%d x3=%d", m.PC, m.CPU.Reg(1), m.CPU.Reg(2), m.CPU.Reg(3))
}

func TestMachineForwardBranch(t *testing.T) {
	// Verify that a forward branch (beq) with a non-taken path followed by
	// a taken branch updates the PC correctly twice in sequence. Stall-on-
	// branch is not yet implemented, so we only test a single forward branch.
	m := NewMachine(64)
	// PC=0:  addi x1, x0, 5
	// PC=4:  addi x2, x0, 5
	// PC=8:  beq x1, x2, 8  -> target = 8 + 8 = 16
	// PC=12: addi x3, x0, 1 (skipped)
	// PC=16: addi x3, x0, 99 <-- should reach
	assert.NoError(t, m.Memory.WriteWord(0, encodeCPU(t, "addi x1, x0, 5")))
	assert.NoError(t, m.Memory.WriteWord(4, encodeCPU(t, "addi x2, x0, 5")))
	assert.NoError(t, m.Memory.WriteWord(8, encodeCPU(t, "beq x1, x2, 8")))
	assert.NoError(t, m.Memory.WriteWord(12, encodeCPU(t, "addi x3, x0, 1")))
	assert.NoError(t, m.Memory.WriteWord(16, encodeCPU(t, "addi x3, x0, 99")))

	for i := 0; i < 50; i++ {
		if err := m.Step(); err != nil {
			t.Fatalf("Step %d: %v", i, err)
		}
		if m.CPU.Reg(3) == 99 {
			return
		}
	}
	t.Fatalf("forward branch test failed; PC=0x%x x1=%d x2=%d x3=%d",
		m.PC, m.CPU.Reg(1), m.CPU.Reg(2), m.CPU.Reg(3))
}
