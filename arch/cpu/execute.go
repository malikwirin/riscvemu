package cpu

// execute scans reservation stations for entries whose operands are ready and
// dispatches each into a free functional unit. Running FUs advance their
// latency countdown. Dispatch order is out-of-order: any ready entry may be
// picked up regardless of its position in the issue queue.
func (c *CPU) execute() {
	for _, a := range c.alus {
		a.Step()
	}
	for i := range c.rs.alu {
		entry := &c.rs.alu[i]
		if !entry.Busy {
			continue
		}
		if entry.Qj != NoTag || entry.Qk != NoTag {
			continue
		}
		if c.kindBelongsToALU(entry.Kind) {
			c.dispatchToALU(i)
		}
	}
}

// kindBelongsToALU reports whether the operation is computed by the ALU pool.
func (c *CPU) kindBelongsToALU(k OpKind) bool {
	switch k {
	case OpADD, OpSUB, OpSLT, OpSLLI, OpBEQ, OpBNE, OpBLT, OpJAL, OpJALR:
		return true
	}
	return false
}

// dispatchToALU hands the RS entry to a free ALU. Returns true on success.
func (c *CPU) dispatchToALU(rsIdx int) bool {
	entry := &c.rs.alu[rsIdx]
	tag := aluTagBase + RSTag(rsIdx)
	for _, a := range c.alus {
		if !a.IsBusy() {
			a.Start(tag, entry.Kind, entry.Rd, entry.Vj, entry.Vk, entry.Imm)
			c.rs.FreeALU(tag)
			c.stats.FunctionalBusyCycles[OpADD]++ // accounting by operation kind
			return true
		}
	}
	return false
}
