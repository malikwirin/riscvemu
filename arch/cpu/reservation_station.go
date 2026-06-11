package cpu

// RSEntry is a single reservation station slot.
type RSEntry struct {
	// Busy indicates whether this slot currently holds a dispatched instruction.
	Busy bool
	// Kind identifies which functional unit and operation this entry will execute.
	Kind OpKind
	// Rd is the architectural destination register (x0..x31) of the instruction.
	Rd uint32
	// Imm is the decoded immediate value carried by the instruction, if any.
	Imm int32
	// Vj is the resolved value of the first source operand.
	Vj uint32
	// Vk is the resolved value of the second source operand.
	Vk uint32
	// Qj is the rename tag of the first source operand when it is not yet resolved
	// (i.e. waiting for the RS entry identified by this tag to write back).
	Qj RSTag
    // Qk is the rename tag of the second source operand when it is not yet resolved.
    Qk RSTag
}

// OperandsReady reports whether both source operands of this entry are
// available: the rename tag of the first source (Qj) and the rename tag of
// the second source (Qk) are both NoTag.
func (e *RSEntry) OperandsReady() bool {
    return e.Qj == NoTag && e.Qk == NoTag
}

// ReservationStation is the pool of reservation station entries for ALU and LSU instructions.
type ReservationStation struct {
	alu        []RSEntry
	lsu        []RSEntry
	nextALUTag RSTag
	nextLSUTag RSTag
}

// aluTagBase and lsuTagBase separate the tag namespaces for ALU and LSU entries.
const (
	aluTagBase RSTag = 1
	lsuTagBase RSTag = 1 << 16
)

func NewReservationStation(aluCount, lsuCount int) *ReservationStation {
	return &ReservationStation{
		alu: make([]RSEntry, aluCount),
		lsu: make([]RSEntry, lsuCount),
	}
}

func (rs *ReservationStation) ALURSCapacity() int { return len(rs.alu) }
func (rs *ReservationStation) LSURSCapacity() int { return len(rs.lsu) }

// AllocateALU places a new entry in the ALU pool and returns its tag, or false if the pool is full.
func (rs *ReservationStation) AllocateALU(entry RSEntry) (RSTag, bool) {
	for i := range rs.alu {
		if !rs.alu[i].Busy {
			rs.alu[i] = entry
			rs.alu[i].Busy = true
			return aluTagBase + RSTag(i), true
		}
	}
	return NoTag, false
}

// AllocateLSU places a new entry in the LSU pool and returns its tag, or false if the pool is full.
func (rs *ReservationStation) AllocateLSU(entry RSEntry) (RSTag, bool) {
	for i := range rs.lsu {
		if !rs.lsu[i].Busy {
			rs.lsu[i] = entry
			rs.lsu[i].Busy = true
			return lsuTagBase + RSTag(i), true
		}
	}
	return NoTag, false
}

// FreeALU releases an ALU reservation station entry by tag.
func (rs *ReservationStation) FreeALU(tag RSTag) {
	if tag < aluTagBase {
		return
	}
	i := int(tag - aluTagBase)
	if i >= 0 && i < len(rs.alu) {
		rs.alu[i] = RSEntry{}
	}
}

// FreeLSU releases an LSU reservation station entry by tag.
func (rs *ReservationStation) FreeLSU(tag RSTag) {
	if tag < lsuTagBase {
		return
	}
	i := int(tag - lsuTagBase)
	if i >= 0 && i < len(rs.lsu) {
		rs.lsu[i] = RSEntry{}
	}
}

// ALUCountBusy returns the number of busy ALU entries.
func (rs *ReservationStation) ALUCountBusy() int {
	n := 0
	for i := range rs.alu {
		if rs.alu[i].Busy {
			n++
		}
	}
	return n
}

// LSUCountBusy returns the number of busy LSU entries.
func (rs *ReservationStation) LSUCountBusy() int {
	n := 0
	for i := range rs.lsu {
		if rs.lsu[i].Busy {
			n++
		}
	}
	return n
}
