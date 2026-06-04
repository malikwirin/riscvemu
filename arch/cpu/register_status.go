package cpu

// RegisterStatus tracks the rename tag currently producing each architectural register.
type RegisterStatus struct {
	Qi [32]RSTag
	V  [32]uint32
}

func NewRegisterStatus() *RegisterStatus {
	return &RegisterStatus{}
}

func (rf *RegisterStatus) Read(idx uint32) uint32 {
	if idx > 31 {
		return 0
	}
	return rf.V[idx]
}

func (rf *RegisterStatus) Write(idx uint32, value uint32) {
	if idx == 0 || idx > 31 {
		return
	}
	rf.V[idx] = value
	rf.Qi[idx] = 0
}
