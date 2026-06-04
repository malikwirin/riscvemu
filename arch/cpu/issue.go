package cpu

// InstrMeta is the decoded view of an instruction used inside the CPU.
type InstrMeta struct {
	Kind OpKind
	Rd   uint32
	Rs1  uint32
	Rs2  uint32
	Imm  int32
}

// issue accepts an encoded instruction word into the issue queue.
func (c *CPU) issue(_ uint32) error {
	c.stats.Issued++
	return nil
}
