package cpu

// RegisterStatus tracks the rename tag currently producing each architectural register.
type RegisterStatus struct {
	Qi [32]RSTag
	V  [32]uint32
	// limit is the configurable upper bound on writable register
	// indices. Writes to idx >= limit are silently dropped and
	// reads return zero. The underlying V/Qi storage remains
	// 32-wide so internal code can still touch the high half
	// for tests that opt out of the limit.
	limit uint32
}

func NewRegisterStatus() *RegisterStatus {
	rs := &RegisterStatus{}
	rs.limit = 32
	return rs
}

// SetLimit sets the upper bound on register indices that Read and
// Write will accept. Indices >= limit are treated as out-of-range:
// reads return zero, writes are dropped. Pass 0 to disable all
// writes and reads (only useful for tests that need a fully
// zeroed register file).
func (rs *RegisterStatus) SetLimit(limit uint32) {
	rs.limit = limit
}

func (rs *RegisterStatus) Read(idx uint32) uint32 {
	if idx >= rs.limit {
		return 0
	}
	return rs.V[idx]
}

func (rs *RegisterStatus) Write(idx uint32, value uint32) {
	if idx == 0 || idx >= rs.limit {
		return
	}
	rs.V[idx] = value
	rs.Qi[idx] = 0
}
