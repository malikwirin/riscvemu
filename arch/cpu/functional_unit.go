package cpu

// ALU executes integer arithmetic and logical operations.
type ALU struct {
	latency int
	busy    bool
	remain  int
	tag     RSTag
}

func newALU(latency int) *ALU {
	return &ALU{latency: latency}
}

func (a *ALU) IsFree() bool { return !a.busy }

func (a *ALU) Start(tag RSTag) {
	a.busy = true
	a.remain = a.latency
	a.tag = tag
}

func (a *ALU) Step() {
	if a.busy {
		a.remain--
		if a.remain <= 0 {
			a.busy = false
		}
	}
}
