package arch

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMemoryInitialization(t *testing.T) {
	mem := NewMemory(4096)
	assert.Equal(t, 4096, len(mem.Data), "Expected memory size 4096 bytes")
	for i, b := range mem.Data {
		assert.Equalf(t, byte(0), b, "Expected memory at address %d to be 0 on init", i)
	}
}

func TestMemoryWriteAndReadWord(t *testing.T) {
	mem := NewMemory(4096)
	addr := uint32(100)
	value := int32(0x12345678)
	assert.NoError(t, mem.StoreWord(addr, value), "Unexpected error on store")
	got, err := mem.LoadWord(addr)
	assert.NoError(t, err, "Unexpected error on load")
	assert.Equalf(t, value, got, "Expected 0x%X at address %d", value, addr)
}

func TestMemoryOverwrite(t *testing.T) {
	mem := NewMemory(4096)
	addr := uint32(200)
	assert.NoError(t, mem.StoreWord(addr, 0x11111111), "Unexpected error on first store")
	assert.NoError(t, mem.StoreWord(addr, 0x22222222), "Unexpected error on second store")
	got, err := mem.LoadWord(addr)
	assert.NoError(t, err, "Unexpected error on load")
	assert.Equalf(t, int32(0x22222222), got, "Expected 0x22222222 at address %d", addr)
}

func TestMemoryReadWord(t *testing.T) {
	mem := NewMemory(4096)
	addr := uint32(120)
	value := int32(0x1EADBEEF)
	assert.NoError(t, mem.StoreWord(addr, value), "Unexpected error on store")
	uval, err := mem.ReadWord(addr)
	assert.NoError(t, err, "Unexpected error on ReadWord")
	assert.Equalf(t, uint32(value), uval, "Expected 0x%X at address %d", uint32(value), addr)
}

func TestMemoryOutOfBounds(t *testing.T) {
	mem := NewMemory(4096)
	_, err := mem.LoadWord(4096)
	assert.Error(t, err, "Expected error on out-of-bounds load")
	err = mem.StoreWord(4096, 123)
	assert.Error(t, err, "Expected error on out-of-bounds store")
	_, err = mem.ReadWord(4096)
	assert.Error(t, err, "Expected error on out-of-bounds ReadWord")
	err = mem.WriteWord(4096, 0xDEADBEEF)
	assert.Error(t, err, "Expected error on out-of-bounds WriteWord")
}

func TestMemoryWriteWord(t *testing.T) {
	mem := NewMemory(4096)
	addr := uint32(256)
	val := uint32(0xDEADBEEF)

	assert.NoError(t, mem.WriteWord(addr, val), "Unexpected error on WriteWord")

	got, err := mem.ReadWord(addr)
	assert.NoError(t, err, "Unexpected error on ReadWord")
	assert.Equalf(t, val, got, "Expected 0x%X at address %d", val, addr)
}

func TestMemory_ReadWordReturnsWrittenInstruction(t *testing.T) {
	mem := NewMemory(4096)
	addr := uint32(0x100)
	expected := uint32(0x00112023) // Typical RISC-V sw instruction

	assert.NoError(t, mem.WriteWord(addr, expected), "WriteWord failed")

	actual, err := mem.ReadWord(addr)
	assert.NoError(t, err, "ReadWord failed")

	// Check for ASCII "STOR" and "TORE"
	assert.NotEqual(t, uint32(0x53544F52), actual, "Unexpected ASCII value 'STOR' read from memory")
	assert.NotEqual(t, uint32(0x544F5245), actual, "Unexpected ASCII value 'TORE' read from memory")
	assert.Equalf(t, expected, actual, "Expected 0x%X at address %d", expected, addr)
}
