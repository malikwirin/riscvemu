package cpu

// execute scans reservation stations for entries whose operands are ready and
// dispatches each into a free functional unit. Running FUs advance their
// latency countdown. Dispatch order is out-of-order: any ready entry may be
// picked up regardless of its position in the issue queue.
func (c *CPU) execute() {
	for _, a := range c.alus {
		a.Step()
	}
	for _, l := range c.lsus {
		l.Step()
	}
    for i := range c.rs.alu {
        entry := &c.rs.alu[i]
        if !entry.Busy || !entry.OperandsReady() {
            continue
        }
        if c.kindBelongsToALU(entry.Kind) {
            c.dispatchToALU(i)
        }
    }
    for i := range c.rs.lsu {
        entry := &c.rs.lsu[i]
        if !entry.Busy || !entry.OperandsReady() {
            continue
        }
        if c.kindBelongsToLSU(entry.Kind) {
            c.dispatchToLSU(i)
        }
    }
}

// kindBelongsToALU reports whether the operation is computed by the ALU pool.
func (c *CPU) kindBelongsToALU(k OpKind) bool {
	switch k {
	case OpADD, OpADDI, OpSUB, OpSLT, OpSLLI, OpBEQ, OpBNE, OpBLT, OpJAL, OpJALR:
		return true
	}
	return false
}

// kindBelongsToLSU reports whether the operation is computed by the LSU pool.
func (c *CPU) kindBelongsToLSU(k OpKind) bool {
	return k == OpLOAD || k == OpSTORE
}

// dispatchToALU hands the RS entry to a free ALU. Returns true on success.
func (c *CPU) dispatchToALU(rsIdx int) bool {
	entry := &c.rs.alu[rsIdx]
	tag := aluTagBase + RSTag(rsIdx)
	for _, a := range c.alus {
		if !a.IsBusy() {
			a.Start(tag, entry.Kind, entry.Rd, entry.Vj, entry.Vk, entry.Imm)
			c.stats.FunctionalBusyCycles[entry.Kind]++
			c.rs.FreeALU(tag)
			return true
		}
	}
	return false
}

// dispatchToLSU hands the RS entry to a free LSU. Returns true on success.
func (c *CPU) dispatchToLSU(rsIdx int) bool {
	entry := &c.rs.lsu[rsIdx]
	tag := lsuTagBase + RSTag(rsIdx)
	for _, l := range c.lsus {
		if !l.IsBusy() {
			l.Start(tag, entry.Kind, entry.Rd, entry.Vj, entry.Vk, entry.Imm)
			c.stats.FunctionalBusyCycles[entry.Kind]++
			c.rs.FreeLSU(tag)
			return true
		}
	}
	return false
}
