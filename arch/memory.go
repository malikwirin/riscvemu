package arch

import (
    "encoding/binary"
	"fmt"
)

type Memory struct {
    Data []byte
}

func NewMemory(size int) *Memory {
    return &Memory{
        Data: make([]byte, size),
    }
}

func (m *Memory) LoadWord(addr uint32) (int32, error) {
	if addr+4 > uint32(len(m.Data)) {
        return 0, fmt.Errorf("address %d out of bounds", addr)
    }
    return int32(binary.LittleEndian.Uint32(m.Data[addr : addr+4])), nil
}

func (m *Memory) StoreWord(addr uint32, value int32) error {
	if addr+4 > uint32(len(m.Data)) {
        return fmt.Errorf("address %d out of bounds", addr)
    }
    binary.LittleEndian.PutUint32(m.Data[addr:addr+4], uint32(value))
	return nil
}

func (m *Memory) WriteWord(addr uint32, value uint32) error {
    if addr+4 > uint32(len(m.Data)) {
        return fmt.Errorf("address %d out of bounds", addr)
    }
    binary.LittleEndian.PutUint32(m.Data[addr:addr+4], value)
    return nil
}

func (m *Memory) ReadWord(addr uint32) (uint32, error) {
    if addr+4 > uint32(len(m.Data)) {
        return 0, fmt.Errorf("address %d out of bounds", addr)
    }
    return binary.LittleEndian.Uint32(m.Data[addr : addr+4]), nil
}
