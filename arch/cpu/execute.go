package cpu

// execute scans reservation stations for entries whose operands are ready and
// dispatches each into a free functional unit. After dispatch, all FUs
// advance their latency countdown (step). Dispatch order is out-of-order:
// any ready entry may be picked up regardless of its position in the issue
// queue. Stall-on-branch applies only to issue (handled there); dispatch is
// always free so that predecessors of an unresolved branch can still resolve
// their operands.
func (c *CPU) execute() {
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
	for _, a := range c.alus {
		a.Step()
	}
	for _, l := range c.lsus {
		l.Step()
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
	for _, a := range c.alus {
		if !a.IsBusy() {
			a.StartAtPC(entry.Tag(), entry.Kind, entry.Rd, entry.Vj, entry.Vk, entry.Imm, entry.Pc)
			c.stats.FunctionalBusyCycles[entry.Kind]++
			c.rs.FreeALU(entry.Tag())
			return true
		}
	}
	return false
}

// dispatchToLSU hands the RS entry to a free LSU. Returns true on success.
func (c *CPU) dispatchToLSU(rsIdx int) bool {
	entry := &c.rs.lsu[rsIdx]
	for _, l := range c.lsus {
		if l.IsFree() {
			l.StartAtPC(entry.Tag(), entry.Kind, entry.Rd, entry.Vj, entry.Vk, entry.Imm, 0)
			c.stats.FunctionalBusyCycles[entry.Kind]++
			c.rs.FreeLSU(entry.Tag())
			return true
		}
	}
	return false
}
