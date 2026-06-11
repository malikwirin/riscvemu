package cpu

import "testing"

func TestReceiveInstructionAcceptsPC(t *testing.T) {
    core := NewCPU(DefaultConfig())
    if err := core.ReceiveInstruction(encode(t, "addi x1, x0, 1"), 0x100); err != nil {
        t.Fatalf("ReceiveInstruction: %v", err)
    }
}

func TestReceiveInstructionAtZero(t *testing.T) {
    core := NewCPU(DefaultConfig())
    if err := core.ReceiveInstruction(encode(t, "addi x1, x0, 1"), 0); err != nil {
        t.Fatalf("ReceiveInstruction at PC=0: %v", err)
    }
}
