package cpu

import "github.com/malikwirin/riscvemu/assembler"

// issue decodes the instruction word and dispatches it to a reservation station.
// In-order issue: if the matching RS pool is full, the instruction stalls.
func (c *CPU) issue(word uint32) error {
	c.stats.Issued++
	meta := decode(assembler.Instruction(word))
	if meta.Kind == OpInvalid {
		return nil
	}
	entry := c.buildEntry(meta)
	var tag RSTag
	var ok bool
	if isLSUOp(meta.Kind) {
		tag, ok = c.rs.AllocateLSU(entry)
		if !ok {
			c.stats.StructuralStalls++
			return nil
		}
	} else {
		tag, ok = c.rs.AllocateALU(entry)
		if !ok {
			c.stats.StructuralStalls++
			return nil
		}
	}
	if meta.Rd != 0 {
		c.rf.Qi[meta.Rd] = tag
	}
	return nil
}

// buildEntry constructs the RS entry for a decoded instruction, including
// current operand values or rename tags from the register status.
func (c *CPU) buildEntry(m InstrMeta) RSEntry {
	entry := RSEntry{
		Kind: m.Kind,
		Rd:   m.Rd,
		Imm:  m.Imm,
	}
	if m.Rs1 != 0 {
		if q := c.rf.Qi[m.Rs1]; q != NoTag {
			entry.Qj = q
		} else {
			entry.Vj = c.rf.V[m.Rs1]
		}
	}
	if m.Rs2 != 0 {
		if q := c.rf.Qi[m.Rs2]; q != NoTag {
			entry.Qk = q
		} else {
			entry.Vk = c.rf.V[m.Rs2]
		}
	}
	return entry
}

// lastIssuedTag returns the tag of the most recently allocated RS entry.
// For simplicity the next-available ALU/LSU tag is used; in a full implementation
// this would track the last allocated tag per pool.
func (c *CPU) lastIssuedTag() RSTag {
	for i := len(c.rs.alu) - 1; i >= 0; i-- {
		if c.rs.alu[i].Busy {
			return aluTagBase + RSTag(i)
		}
	}
	for i := len(c.rs.lsu) - 1; i >= 0; i-- {
		if c.rs.lsu[i].Busy {
			return lsuTagBase + RSTag(i)
		}
	}
	return NoTag
}
