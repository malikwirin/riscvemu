package cpu

// CDBResult is a single broadcast on the common data bus.
type CDBResult struct {
	Tag   RSTag
	Value uint32
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
