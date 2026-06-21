package trace

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// memRefArgs matches "Rd, offset(Rs)" for LOAD and "Rs, offset(Rd)"
// for STORE. Whitespace around the comma, around the offset, and
// inside the parentheses is tolerated. The offset is a signed
// decimal. The Spec does not pin the encoding width, but the
// underlying RISC-V I/S-type format uses a 12-bit signed immediate,
// which the encoder (A.4) enforces. We surface the overflow here so
// the trace line is identified up front.
var memRefArgs = regexp.MustCompile(`^R([0-7]),\s*(-?\d+)\s*\(\s*R([0-7])\s*\)$`)

// parseMemRef decodes a LOAD or STORE operand string. Both mnemonics
// share the same surface syntax in the Spec: a register, a comma,
// and a memory reference "offset(Rs)". The first captured register
// is what the Spec calls "Rd" for LOAD (the destination) and "Rs"
// for STORE (the value to be stored). The encoded RISC-V argument
// order is the inverse for STORE; that inversion happens in A.4.
//
// Operand register letters are normalised to upper case so the
// regex can stay strict ("R[0-7]").
func parseMemRef(operands string) (reg string, offset int64, baseReg string, err error) {
	m := memRefArgs.FindStringSubmatch(strings.ToUpper(operands))
	if m == nil {
		err = fmt.Errorf("invalid memory-reference operands: %q", operands)
		return
	}
	offset, perr := strconv.ParseInt(m[2], 10, 64)
	if perr != nil {
		err = fmt.Errorf("invalid offset: %q", m[2])
		return
	}
	if offset < -2048 || offset > 2047 {
		err = fmt.Errorf("offset out of range for memory reference: %d", offset)
		return
	}
	return "R" + m[1], offset, "R" + m[3], nil
}
