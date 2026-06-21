package trace

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// supportedMnemonics is the Spec-mandated subset of RISC-V
// instructions that the trace parser accepts. The keys are upper-case
// canonical names; the parser lowercases the input before lookup.
var supportedMnemonics = map[string]bool{
	"ADD":   true,
	"SUB":   true,
	"MUL":   true,
	"DIV":   true,
	"LOAD":  true,
	"STORE": true,
}

// rTypeArgs matches "Rd, Rs1, Rs2" with register indices 0..7. Whitespace
// around commas is tolerated so that the parser is forgiving of
// "ADD R1,R2,R3" and "ADD R1, R2, R3" alike. R0 is intentionally
// accepted (and writes are silently discarded on the encoding side).
var rTypeArgs = regexp.MustCompile(`^R([0-7]),\s*R([0-7]),\s*R([0-7])$`)

// parseRType decodes the operand string of an R-type instruction
// (ADD/SUB/MUL/DIV) into three register names. The caller has already
// validated the mnemonic. Operand register letters are case-insensitive
// (so "r1" and "R1" both yield "R1" in the output), matching the
// overall case-insensitive policy of the trace format.
func parseRType(operands string) ([3]string, error) {
	// Normalize operand register letters to upper case so the regex
	// can stay strict ("R[0-7]"). The trace format is fully
	// case-insensitive per the Spec.
	m := rTypeArgs.FindStringSubmatch(strings.ToUpper(operands))
	if m == nil {
		return [3]string{}, fmt.Errorf("invalid R-type operands: %q", operands)
	}
	return [3]string{"R" + m[1], "R" + m[2], "R" + m[3]}, nil
}

// parseLine dispatches one non-empty, non-comment line to the
// appropriate sub-parser. Comment handling is "lines starting with
// '#' (after trimming) are skipped"; this convention is a superset
// of the Spec example, which has no comments. Empty lines are also
// skipped.
func parseLine(raw string, lineNum int) (Line, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" || strings.HasPrefix(trimmed, "#") {
		return Line{}, nil
	}

	// Single-character line is a control command.
	if len(trimmed) == 1 {
		op := trimmed[0]
		switch op {
		case 'v', 's', 'h', 'i':
			return Line{Kind: LineControl, LineNum: lineNum, Control: Control{Op: op}}, nil
		}
		return Line{}, &ParseError{LineNum: lineNum, Line: raw, Reason: fmt.Sprintf("unknown control command: %q", string(op))}
	}

	// Otherwise split mnemonic from operands. The Spec uses a single
	// space as the separator; we tolerate runs of whitespace.
	parts := strings.Fields(trimmed)
	if len(parts) < 2 {
		return Line{}, &ParseError{LineNum: lineNum, Line: raw, Reason: fmt.Sprintf("expected mnemonic and operands, got %q", trimmed)}
	}
	mnemonic := strings.ToUpper(parts[0])
	if !supportedMnemonics[mnemonic] {
		return Line{}, &ParseError{LineNum: lineNum, Line: raw, Reason: fmt.Sprintf("unsupported mnemonic: %q", parts[0])}
	}
	operands := strings.Join(parts[1:], "")
	args, err := parseInstrArgs(mnemonic, operands)
	if err != nil {
		return Line{}, &ParseError{LineNum: lineNum, Line: raw, Reason: err.Error()}
	}
	return Line{Kind: LineInstr, LineNum: lineNum, Instr: Instr{Mnemonic: mnemonic, Args: args}}, nil
}

// parseInstrArgs dispatches by mnemonic to the operand-level parser.
// R-type (ADD/SUB/MUL/DIV) and memory-reference (LOAD/STORE)
// subsets are both implemented; A.3 closes the loop on the latter
// two. The memory-reference operands are stored as
// "Args[0] = reg, Args[1] = offset(Rbase)" where the two
// registers have been normalised to upper case.
func parseInstrArgs(mnemonic, operands string) ([3]string, error) {
	switch mnemonic {
	case "ADD", "SUB", "MUL", "DIV":
		return parseRType(operands)
	case "LOAD", "STORE":
		reg, offset, base, err := parseMemRef(operands)
		if err != nil {
			return [3]string{}, err
		}
		return [3]string{reg, formatMemRef(offset, base)}, nil
	}
	return [3]string{}, fmt.Errorf("unsupported mnemonic: %q", mnemonic)
}

// formatMemRef rebuilds the canonical "offset(Rbase)" form for the
// Instr.Args[1] slot. The offset is always written in decimal,
// negative numbers carry their sign.
func formatMemRef(offset int64, base string) string {
	return strconv.FormatInt(offset, 10) + "(" + base + ")"
}
