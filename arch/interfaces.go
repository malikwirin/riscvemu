package arch

type WordReader interface {
    ReadWord(addr uint32) (uint32, error)
}

// Compile-time check: *Memory implements WordReader
var _ WordReader = (*Memory)(nil)
