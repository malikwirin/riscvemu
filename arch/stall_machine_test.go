package arch

import (
	"testing"

	"github.com/malikwirin/riscvemu/arch/cpu"
	"github.com/stretchr/testify/assert"
)

func TestMachineStallOnBranch(t *testing.T) {
	m := NewMachine(64)
	// PC=0:  addi x1, x0, 5
	// PC=4:  addi x2, x0, 5
	// PC=8:  beq x1, x2, 12  (taken, target=20)
	// PC=12: addi x3, x0, 99  (must be skipped)
	// PC=16: addi x4, x0, 1   (must be skipped)
	// PC=20: addi x5, x0, 42  (target after taken branch)
	assert.NoError(t, m.Memory.WriteWord(0, encodeCPU(t, "addi x1, x0, 5")))
	assert.NoError(t, m.Memory.WriteWord(4, encodeCPU(t, "addi x2, x0, 5")))
	assert.NoError(t, m.Memory.WriteWord(8, encodeCPU(t, "beq x1, x2, 12")))
	assert.NoError(t, m.Memory.WriteWord(12, encodeCPU(t, "addi x3, x0, 99")))
	assert.NoError(t, m.Memory.WriteWord(16, encodeCPU(t, "addi x4, x0, 1")))
	assert.NoError(t, m.Memory.WriteWord(20, encodeCPU(t, "addi x5, x0, 42")))

	for i := 0; i < 50; i++ {
		if err := m.Step(); err != nil {
			t.Fatalf("Step %d: %v", i, err)
		}
		if m.CPU.Reg(5) == 42 {
			// x3 and x4 must remain zero: the beq was taken, and stall-on-branch
			// prevented the addi at PC=12 and PC=16 from being issued before the
			// branch resolved.
			assert.Equal(t, uint32(0), m.CPU.Reg(3), "x3 should be 0 (branch skipped the addi)")
			assert.Equal(t, uint32(0), m.CPU.Reg(4), "x4 should be 0 (branch skipped the addi)")
			return
		}
	}
	t.Fatalf("branch did not reach target; final PC=0x%x, x1=%d x2=%d x3=%d x4=%d x5=%d",
		m.PC, m.CPU.Reg(1), m.CPU.Reg(2), m.CPU.Reg(3), m.CPU.Reg(4), m.CPU.Reg(5))
}

func TestMachineBranchStallCounter(t *testing.T) {
	m := NewMachineWithConfig(64, cpu.Config{
		ALURSCount: 1, ALULatency: 3, LoadLatency: 2, StoreLatency: 2,
		InstructionQueueSize: 8,
	})
	// With ALULatency=3 the addi takes three cycles. The beq issued in
	// cycle 2 has a pending operand x1, so the beq itself is the stalled
	// branch — every subsequent issue while it sits in the IQ counts as a
	// branch stall.
	// PC=0:  addi x1, x0, 5
	// PC=4:  beq x1, x0, 12
	// PC=8:  addi x3, x0, 1
	// PC=12: addi x4, x0, 2
	assert.NoError(t, m.Memory.WriteWord(0, encodeCPU(t, "addi x1, x0, 5")))
	assert.NoError(t, m.Memory.WriteWord(4, encodeCPU(t, "beq x1, x0, 12")))
	assert.NoError(t, m.Memory.WriteWord(8, encodeCPU(t, "addi x3, x0, 1")))
	assert.NoError(t, m.Memory.WriteWord(12, encodeCPU(t, "addi x4, x0, 2")))

	for i := 0; i < 10; i++ {
		if err := m.Step(); err != nil {
			t.Fatalf("Step %d: %v", i, err)
		}
	}
	if got := m.CPU.Stats().BranchStalls; got == 0 {
		t.Errorf("BranchStalls = 0, want > 0 (beq stalled by pending operand)")
	}
}
