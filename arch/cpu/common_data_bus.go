package cpu

// CDBResult is a single broadcast on the common data bus.
type CDBResult struct {
	// Tag is the rename tag of the RS entry that produced the value.
	Tag RSTag
	// Value is the computed 32-bit result being broadcast.
	Value uint32
	// Rd is the architectural destination register of the in-flight operation.
	Rd uint32
}

// CommonDataBus arbitrates one CDBResult per cycle.
type CommonDataBus struct {
	slot CDBResult
	busy bool
}

func NewCommonDataBus() *CommonDataBus {
	return &CommonDataBus{}
}

func (c *CommonDataBus) Broadcast(r CDBResult) {
	c.slot = r
	c.busy = true
}

func (c *CommonDataBus) Consume() (CDBResult, bool) {
	if !c.busy {
		return CDBResult{}, false
	}
	c.busy = false
	return c.slot, true
}

// Last returns the most recent broadcast without consuming it.
func (c *CommonDataBus) Last() (CDBResult, bool) {
	if !c.busy {
		return CDBResult{}, false
	}
	return c.slot, true
}
