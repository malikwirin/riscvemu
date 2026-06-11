package cpu

// writeback picks at most one completed result per cycle, broadcasts it on
// the common data bus, wakes any reservation station entries waiting on its
// tag, and commits the value to the architectural register file.
func (c *CPU) writeback() {
	c.lastBranch = BranchInfo{}
	for _, a := range c.alus {
		tag, value, rd, ok := a.TakeResult()
		if !ok {
			continue
		}
		c.cdb.Broadcast(CDBResult{Tag: tag, Value: value, Rd: rd})
		c.stats.Retired++
		c.wakeReservationStations(tag, value)
		if rd != 0 {
			c.rf.V[rd] = value
			c.rf.Qi[rd] = NoTag
		}
		if a.IsBranchTaken() {
			link := a.LinkInfo()
			info := BranchInfo{
				IsBranch:  true,
				Taken:     true,
				Target:    a.BranchTarget(),
				LinkReg:   link.Reg,
				LinkValue: link.Value,
			}
			if link.Active && link.Reg != 0 {
				c.rf.V[link.Reg] = link.Value
				c.rf.Qi[link.Reg] = NoTag
			}
			c.lastBranch = info
		} else if isBranchKind(a.Kind()) {
			c.lastBranch = BranchInfo{IsBranch: true, Taken: false}
		}
		// Clear the issue-stage gate only for the branches that set it.
		if isConditionalBranchKind(a.Kind()) {
			c.hasUnresolvedBranch = false
		}
		return // only one broadcast per cycle
	}
	for _, l := range c.lsus {
		tag, value, rd, ok := l.TakeResult()
		if !ok {
			continue
		}
		c.cdb.Broadcast(CDBResult{Tag: tag, Value: value, Rd: rd})
		// Stores have no destination register; only loads retire.
		if rd != 0 {
			c.stats.Retired++
		}
		c.wakeReservationStations(tag, value)
		if rd != 0 {
			c.rf.V[rd] = value
			c.rf.Qi[rd] = NoTag
		}
		return // only one broadcast per cycle
	}
}

// wakeReservationStations scans every RS entry for pending operands that
// reference the broadcast tag. When found, the operand is filled in and
// the rename tag is cleared. RAWResolved counts each such resolution.
func (c *CPU) wakeReservationStations(tag RSTag, value uint32) {
	for i := range c.rs.alu {
		if !c.rs.alu[i].Busy {
			continue
		}
		if c.rs.alu[i].Qj == tag {
			c.rs.alu[i].Vj = value
			c.rs.alu[i].Qj = NoTag
			c.stats.RAWResolved++
		}
		if c.rs.alu[i].Qk == tag {
			c.rs.alu[i].Vk = value
			c.rs.alu[i].Qk = NoTag
			c.stats.RAWResolved++
		}
	}
	for i := range c.rs.lsu {
		if !c.rs.lsu[i].Busy {
			continue
		}
		if c.rs.lsu[i].Qj == tag {
			c.rs.lsu[i].Vj = value
			c.rs.lsu[i].Qj = NoTag
			c.stats.RAWResolved++
		}
		if c.rs.lsu[i].Qk == tag {
			c.rs.lsu[i].Vk = value
			c.rs.lsu[i].Qk = NoTag
			c.stats.RAWResolved++
		}
	}
}
