package cpu

// ReservationStation is the pool of reservation station entries for ALU and LSU instructions.
type ReservationStation struct {
	aluCount int
	lsuCount int
}

func NewReservationStation(aluCount, lsuCount int) *ReservationStation {
	return &ReservationStation{aluCount: aluCount, lsuCount: lsuCount}
}

func (rs *ReservationStation) ALURSCapacity() int { return rs.aluCount }
func (rs *ReservationStation) LSURSCapacity() int { return rs.lsuCount }
