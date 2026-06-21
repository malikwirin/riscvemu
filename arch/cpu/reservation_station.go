package cpu

// RSEntry is a single reservation station slot.
type RSEntry struct {
	// Busy indicates whether this slot currently holds a dispatched instruction.
	Busy bool
	// tag is the unique rename tag assigned at allocation. It is preserved
	// even after the entry is "cleared" for re-use so that pending wakeups
	// can find the original producing entry.
	tag  RSTag
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
	// Pc is the program counter of the instruction that produced this entry;
	// needed so branch resolution can compute the target at dispatch time.
	Pc uint32
}

// Tag returns the unique rename tag assigned to this entry.
func (e *RSEntry) Tag() RSTag { return e.tag }

// Clear resets all operand state and Busy but preserves the tag.
func (e *RSEntry) Clear() {
	e.Busy = false
	e.Kind = 0
	e.Rd = 0
	e.Imm = 0
	e.Vj = 0
	e.Vk = 0
	e.Qj = NoTag
	e.Qk = NoTag
	e.Pc = 0
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

// nextALUTag and nextLSUTag are monotonically increasing counters that
// guarantee every issued instruction gets a unique tag within its pool,
// even when the same RS slot is reused.
var (
	nextALUTag RSTag = aluTagBase
	nextLSUTag RSTag = lsuTagBase
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
			tag := nextALUTag
			nextALUTag++
			entry.tag = tag
			entry.Busy = true
			rs.alu[i] = entry
			return tag, true
		}
	}
	return NoTag, false
}

// AllocateLSU places a new entry in the LSU pool and returns its tag, or false if the pool is full.
func (rs *ReservationStation) AllocateLSU(entry RSEntry) (RSTag, bool) {
	for i := range rs.lsu {
		if !rs.lsu[i].Busy {
			tag := nextLSUTag
			nextLSUTag++
			entry.tag = tag
			entry.Busy = true
			rs.lsu[i] = entry
			return tag, true
		}
	}
	return NoTag, false
}

// FreeALU releases an ALU reservation station entry by tag.
// The slot is identified by its stored tag, not by index.
func (rs *ReservationStation) FreeALU(tag RSTag) {
	for i := range rs.alu {
		if rs.alu[i].tag == tag {
			rs.alu[i].Busy = false
		}
	}
}

// FreeLSU releases an LSU reservation station entry by tag.
func (rs *ReservationStation) FreeLSU(tag RSTag) {
	for i := range rs.lsu {
		if rs.lsu[i].tag == tag {
			rs.lsu[i].Busy = false
		}
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
