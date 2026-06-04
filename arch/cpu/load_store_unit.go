package cpu

// LSU is the load/store unit and bridges to memory.
type LSU struct {
	mem WordHandler
}

func NewLSU() *LSU {
	return &LSU{}
}

func (l *LSU) AttachMemory(mem WordHandler) {
	l.mem = mem
}
