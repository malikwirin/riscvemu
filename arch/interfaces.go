package arch

type WordHandler interface {
	ReadWord(addr uint32) (uint32, error)
	WriteWord(addr uint32, value uint32) error
}

// Compile-time check: *Memory implements interfaces
var _ WordHandler = (*Memory)(nil)
