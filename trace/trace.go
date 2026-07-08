// The parser produces a stream of Line values that downstream code
// (the trace driver, in a later change) turns into actual RISC-V
// instruction words. Parsing is independent of execution: this file
// only knows the trace syntax, not the simulation model.
package trace

import (
	"os"
	"strings"
)

// LineKind discriminates between an instruction and a control command.
type LineKind uint8

const (
	// LineInstr marks a parsed instruction line (ADD/SUB/MUL/DIV/LOAD/STORE).
	LineInstr LineKind = iota
	// LineControl marks a single-character control command (v/s/h/i).
	LineControl
)

// Instr is a single trace instruction. The Mnemonic is the
// upper-cased canonical name; Args preserves the raw operand strings
// in the order they appeared on the line:
//
//   - ADD, SUB, MUL, DIV:  three register names R0..R7.
//   - LOAD:  Args[0] is the destination register; Args[1] is the
//     memory reference "offset(Rs)" in raw form.
//   - STORE:  Args[0] is the source register (the value to be stored);
//     Args[1] is the memory reference "offset(Rd)" in raw form. Note
//     that STORE is the only mnemonic whose first operand is a value
//     rather than a destination, which is the inverse of the RISC-V
//     STORE encoding convention.
type Instr struct {
	Mnemonic string
	Args     [3]string
}

// Control is a single trace control command. Op is one of 'v', 's',
// 'h', 'i'. The toggle semantics (v on/off, s/h/i stateless) are
// implemented by the trace driver, not the parser.
type Control struct {
	Op byte
}

// Line is one parsed trace line. Either Instr or Control is set
// according to Kind.
type Line struct {
	Kind    LineKind
	LineNum int // 1-based, for error messages
	Instr   Instr
	Control Control
}

// ParseError carries the line number and offending text so the trace
// driver (or the user) can report exactly where parsing broke.
type ParseError struct {
	LineNum int
	Line    string
	Reason  string
}

func (e *ParseError) Error() string {
	return e.Reason + " at line " + itoa(e.LineNum) + ": " + e.Line
}

// ParseTraceFile reads the trace file at path and returns its parsed
// lines. On the first parse error it returns *ParseError and the lines
// collected so far are discarded.
//
// I/O errors (file not found, permission denied, ...) are wrapped in
// *ParseError with LineNum 0 to signal that the failure happened
// before any line was examined.
func ParseTraceFile(path string) ([]Line, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, &ParseError{LineNum: 0, Reason: "read " + path + ": " + err.Error()}
	}
	return ParseTrace(string(data))
}

// ParseTrace parses a trace from a string. Useful for tests and for
// callers that already have the trace content in memory.
//
// Empty lines and lines starting with '#' (after trimming) are
// silently skipped; they do not advance the line counter reported in
// ParseError. The first parse error short-circuits the run.
func ParseTrace(src string) ([]Line, error) {
	var lines []Line
	for i, raw := range strings.Split(src, "\n") {
		line, err := parseLine(raw, i+1)
		if err != nil {
			return nil, err
		}
		if line.Kind == 0 && line.LineNum == 0 {
			// Empty or comment line; skip.
			continue
		}
		lines = append(lines, line)
	}
	return lines, nil
}

// itoa is a small signed-decimal-to-string helper used by ParseError.
// It avoids pulling in strconv at the error site, which keeps the
// dependency surface of this package minimal.
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var buf [16]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}
