package cpu

import "testing"

func TestStatisticsIPC(t *testing.T) {
	core := NewCPU(DefaultConfig())
	// Issue 2 addi, let them complete
	if err := core.ReceiveInstruction(encode(t, "addi x1, x0, 1")); err != nil {
		t.Fatal(err)
	}
	if err := core.ReceiveInstruction(encode(t, "addi x2, x0, 2")); err != nil {
		t.Fatal(err)
	}
	core.RunCycle() // dispatch both
	core.RunCycle() // first broadcast
	core.RunCycle() // second broadcast
	s := core.Stats()
	if got := s.IPC(); got < 0.5 || got > 1.0 {
		t.Errorf("IPC = %f, want ~0.67 (2 retired / 3 cycles)", got)
	}
}

func TestStatisticsFUUtil(t *testing.T) {
	core := NewCPU(DefaultConfig())
	if err := core.ReceiveInstruction(encode(t, "addi x1, x0, 1")); err != nil {
		t.Fatal(err)
	}
	core.RunCycle() // dispatch
	core.RunCycle() // step, writeback
	s := core.Stats()
	// 1 functional busy cycle for ADDI, total 2 cycles => 0.5
	if got := s.FUUtil(OpADDI); got != 0.5 {
		t.Errorf("FUUtil(OpADDI) = %f, want 0.5", got)
	}
}
