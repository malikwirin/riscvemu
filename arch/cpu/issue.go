package cpu

import (
	"github.com/malikwirin/riscvemu/assembler"
)

// issueStage pulls one entry from the instruction queue and tries to dispatch
// it into a reservation station. In-order issue: if the matching RS pool is
// full, the instruction stalls and stays in the IQ for the next attempt.
// Stall-on-branch: while a branch or jump is unresolved the issue stage is
// gated and the cycle is counted in BranchStalls.
func (c *CPU) issueStage() {
	if c.hasUnresolvedBranch {
		c.stats.BranchStalls++
		return
	}
	word, pc, ok := c.iq.Dequeue()
	if !ok {
		return
	}
	meta := decode(assembler.Instruction(word))
	if meta.Kind == OpInvalid {
		c.stats.Issued++
		return
	}
	c.instrPC = pc
	c.stats.Issued++
	entry := c.buildEntry(meta)
	var tag RSTag
	var allocated bool
	if isLSUOp(meta.Kind) {
		tag, allocated = c.rs.AllocateLSU(entry)
	} else {
		tag, allocated = c.rs.AllocateALU(entry)
	}
	if !allocated {
		// Reservation station full; put the instruction back at the head
		// of the queue so it is the next one tried.
		c.stats.StructuralStalls++
		c.iq.RequeueHead(word, pc)
		return
	}
	if meta.Rd != 0 {
		c.rf.Qi[meta.Rd] = tag
	}
	// JAL and JALR are unconditional; the only outcome is "taken" and the
	// link register, so they do not block the issue stage. Only conditional
	// branches (BEQ/BNE/BLT) need to stall issue until they resolve.
	if isConditionalBranchKind(meta.Kind) {
		c.hasUnresolvedBranch = true
	}
}

// buildEntry constructs the RS entry for a decoded instruction, including
// current operand values or rename tags from the register status.
func (c *CPU) buildEntry(m InstrMeta) RSEntry {
	entry := RSEntry{
		Kind: m.Kind,
		Rd:   m.Rd,
		Imm:  m.Imm,
		Pc:   c.instrPC,
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
