package arch

import "testing"

func TestMemoryInitialization(t *testing.T) {
    mem := NewMemory(4096)
    if len(mem.Data) != 4096 {
        t.Errorf("Expected memory size 4096 bytes, got %d", len(mem.Data))
    }
    for i, b := range mem.Data {
        if b != 0 {
            t.Errorf("Expected memory at address %d to be 0 on init, got %d", i, b)
            break
        }
    }
}

func TestMemoryWriteAndReadWord(t *testing.T) {
    mem := NewMemory(4096)
    addr := uint32(100)
    value := int32(0x12345678)
    if err := mem.StoreWord(addr, value); err != nil {
        t.Fatalf("Unexpected error on store: %v", err)
    }
    got, err := mem.LoadWord(addr)
	if err != nil {
        t.Fatalf("Unexpected error on load: %v", err)
    }
    if got != value {
        t.Errorf("Expected 0x%X at address %d, got 0x%X", value, addr, got)
    }
}

func TestMemoryOverwrite(t *testing.T) {
    mem := NewMemory(4096)
    addr := uint32(200)
    if err := mem.StoreWord(addr, 0x11111111); err != nil {
        t.Fatalf("Unexpected error on first store: %v", err)
    }
    if err := mem.StoreWord(addr, 0x22222222); err != nil {
        t.Fatalf("Unexpected error on second store: %v", err)
    }
    got, err := mem.LoadWord(addr)
    if err != nil {
        t.Fatalf("Unexpected error on load: %v", err)
    }
    if got != 0x22222222 {
        t.Errorf("Expected 0x22222222 at address %d, got 0x%X", addr, got)
    }
}

func TestMemoryReadWord(t *testing.T) {
    mem := NewMemory(4096)
    addr := uint32(120)
    value := int32(0x1EADBEEF)
    if err := mem.StoreWord(addr, value); err != nil {
        t.Fatalf("Unexpected error on store: %v", err)
    }
    uval, err := mem.ReadWord(addr)
    if err != nil {
        t.Fatalf("Unexpected error on ReadWord: %v", err)
    }
    if uval != uint32(value) {
        t.Errorf("Expected 0x%X at address %d, got 0x%X", uint32(value), addr, uval)
    }
}

func TestMemoryOutOfBounds(t *testing.T) {
    mem := NewMemory(4096)
    _, err := mem.LoadWord(4096)
    if err == nil {
        t.Error("Expected error on out-of-bounds load, got nil")
    }
    err = mem.StoreWord(4096, 123)
    if err == nil {
        t.Error("Expected error on out-of-bounds store, got nil")
    }
	_, err = mem.ReadWord(4096)
    if err == nil {
        t.Error("Expected error on out-of-bounds ReadWord, got nil")
    }
}
