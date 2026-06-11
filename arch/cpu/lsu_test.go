package cpu

import "testing"

// MockWordHandler is a simple in-memory backend for CPU tests.
type MockWordHandler struct {
	Instr uint32
	Err   error
	Mem   map[uint32]uint32
}

func (m *MockWordHandler) ReadWord(addr uint32) (uint32, error) {
	if m.Err != nil {
		return 0, m.Err
	}
	if m.Mem != nil {
		if val, ok := m.Mem[addr]; ok {
			return val, nil
		}
	}
	return m.Instr, nil
}

func (m *MockWordHandler) WriteWord(addr uint32, value uint32) error {
	if m.Err != nil {
		return m.Err
	}
	if m.Mem != nil {
		m.Mem[addr] = value
	}
	return nil
}

func TestLSULoadReadsFromMemory(t *testing.T) {
	core := NewCPU(DefaultConfig())
	mem := &MockWordHandler{Mem: map[uint32]uint32{100: 0xDEADBEEF}}
	core.AttachMemory(mem)
	if err := core.ReceiveInstruction(encode(t, "addi x2, x0, 100")); err != nil {
		t.Fatal(err)
	}
	if err := core.ReceiveInstruction(encode(t, "lw x1, 0(x2)")); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 10; i++ {
		core.RunCycle()
		if core.Reg(1) == 0xDEADBEEF {
			return
		}
	}
	t.Errorf("x1 = %#x, want 0xDEADBEEF", core.Reg(1))
}

func TestLSUStoreWritesToMemory(t *testing.T) {
	core := NewCPU(DefaultConfig())
	mem := &MockWordHandler{Mem: map[uint32]uint32{}}
	core.AttachMemory(mem)
	if err := core.ReceiveInstruction(encode(t, "addi x2, x0, 200")); err != nil {
		t.Fatal(err)
	}
	if err := core.ReceiveInstruction(encode(t, "addi x3, x0, 1234")); err != nil {
		t.Fatal(err)
	}
	if err := core.ReceiveInstruction(encode(t, "sw x3, 0(x2)")); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 15; i++ {
		core.RunCycle()
		if mem.Mem[200] == 1234 {
			return
		}
	}
	t.Errorf("mem[200] = %#x, want 1234", mem.Mem[200])
}

func TestLSULoadWaitsForPendingAddress(t *testing.T) {
	core := NewCPU(DefaultConfig())
	mem := &MockWordHandler{Mem: map[uint32]uint32{100: 0x42}}
	core.AttachMemory(mem)
	if err := core.ReceiveInstruction(encode(t, "addi x2, x0, 100")); err != nil {
		t.Fatal(err)
	}
	if err := core.ReceiveInstruction(encode(t, "lw x1, 0(x2)")); err != nil {
		t.Fatal(err)
	}
	// Run enough cycles for the dependency to resolve and the load to complete.
	for i := 0; i < 10; i++ {
		core.RunCycle()
		if core.Reg(1) == 0x42 {
			return
		}
	}
	t.Errorf("x1 = %#x, want 0x42 (load should complete after x2 is ready)", core.Reg(1))
}

func TestLSUStoreLatency(t *testing.T) {
	cfg := DefaultConfig()
	cfg.StoreLatency = 3
	core := NewCPU(cfg)
	mem := &MockWordHandler{Mem: map[uint32]uint32{}}
	core.AttachMemory(mem)
	if err := core.ReceiveInstruction(encode(t, "addi x2, x0, 300")); err != nil {
		t.Fatal(err)
	}
	if err := core.ReceiveInstruction(encode(t, "addi x3, x0, 1500")); err != nil {
		t.Fatal(err)
	}
	if err := core.ReceiveInstruction(encode(t, "sw x3, 0(x2)")); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 20; i++ {
		core.RunCycle()
		if mem.Mem[300] == 1500 {
			return
		}
	}
	t.Errorf("mem[300] = %#x, want 1500", mem.Mem[300])
}

func TestLSULoadWithNonZeroImmediateOffset(t *testing.T) {
	core := NewCPU(DefaultConfig())
	mem := &MockWordHandler{Mem: map[uint32]uint32{64: 0xABCD}}
	core.AttachMemory(mem)
	if err := core.ReceiveInstruction(encode(t, "addi x2, x0, 60")); err != nil {
		t.Fatal(err)
	}
	if err := core.ReceiveInstruction(encode(t, "lw x1, 4(x2)")); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 10; i++ {
		core.RunCycle()
		if core.Reg(1) == 0xABCD {
			return
		}
	}
	t.Errorf("x1 = %#x, want 0xABCD (address 60+4=64)", core.Reg(1))
}

func TestLSUStoreDoesNotBroadcastResult(t *testing.T) {
	core := NewCPU(DefaultConfig())
	mem := &MockWordHandler{Mem: map[uint32]uint32{}}
	core.AttachMemory(mem)
	if err := core.ReceiveInstruction(encode(t, "addi x2, x0, 400")); err != nil {
		t.Fatal(err)
	}
	if err := core.ReceiveInstruction(encode(t, "addi x3, x0, 1000")); err != nil {
		t.Fatal(err)
	}
	if err := core.ReceiveInstruction(encode(t, "sw x3, 0(x2)")); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 20; i++ {
		core.RunCycle()
		if mem.Mem[400] == 1000 {
			break
		}
	}
	// After two addi broadcast (Retired=2) and one store (no broadcast),
	// Retired must still equal 2.
	if got := core.Stats().Retired; got != 2 {
		t.Errorf("Retired = %d, want 2 (store should not broadcast)", got)
	}
}
