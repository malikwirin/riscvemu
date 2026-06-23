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

// TestMemoryStoreLoadRoundtrip pins the three "write then
// read" shapes (StoreWord/LoadWord, WriteWord/ReadWord,
// StoreWord+overwrite/LoadWord) and the out-of-bounds guard
// for all four entry points. Each case is one row of the
// table, so adding a new entry point only needs one more row.
func TestMemoryStoreLoadRoundtrip(t *testing.T) {
	mem := NewMemory(4096)
	cases := []struct {
		name string
		do   func() error
	}{
		{"StoreWord/LoadWord", func() error {
			if err := mem.StoreWord(100, int32(0x12345678)); err != nil {
				return err
			}
			got, err := mem.LoadWord(100)
			if err != nil {
				return err
			}
			assert.Equalf(t, int32(0x12345678), got, "LoadWord(100) = %#x", got)
			return nil
		}},
		{"StoreWord+overwrite/LoadWord", func() error {
			if err := mem.StoreWord(200, int32(0x11111111)); err != nil {
				return err
			}
			if err := mem.StoreWord(200, int32(0x22222222)); err != nil {
				return err
			}
			got, err := mem.LoadWord(200)
			if err != nil {
				return err
			}
			assert.Equalf(t, int32(0x22222222), got, "LoadWord(200) = %#x", got)
			return nil
		}},
		{"StoreWord/ReadWord unsigned", func() error {
			if err := mem.StoreWord(120, int32(0x1EADBEEF)); err != nil {
				return err
			}
			uval, err := mem.ReadWord(120)
			if err != nil {
				return err
			}
			assert.Equalf(t, uint32(0x1EADBEEF), uval, "ReadWord(120) = %#x", uval)
			return nil
		}},
		{"WriteWord/ReadWord", func() error {
			if err := mem.WriteWord(256, uint32(0xDEADBEEF)); err != nil {
				return err
			}
			got, err := mem.ReadWord(256)
			if err != nil {
				return err
			}
			assert.Equalf(t, uint32(0xDEADBEEF), got, "ReadWord(256) = %#x", got)
			return nil
		}},
		{"WriteWord preserves bit pattern", func() error {
			// A typical RISC-V sw instruction: must not be
			// reinterpreted as ASCII "STOR" / "TORE".
			expected := uint32(0x00112023)
			if err := mem.WriteWord(0x100, expected); err != nil {
				return err
			}
			actual, err := mem.ReadWord(0x100)
			if err != nil {
				return err
			}
			assert.NotEqual(t, uint32(0x53544F52), actual, "ASCII 'STOR' leaked from WriteWord")
			assert.NotEqual(t, uint32(0x544F5245), actual, "ASCII 'TORE' leaked from WriteWord")
			assert.Equalf(t, expected, actual, "ReadWord(0x100) = %#x", actual)
			return nil
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.NoError(t, tc.do())
		})
	}
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
