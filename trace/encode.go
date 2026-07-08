package trace

import (
	"fmt"
	"strconv"
	"strings"

	"codeberg.org/malik/riscvemu/assembler"
)

// EncodeInstr turns a parsed trace instruction into the equivalent
// 32-bit RISC-V instruction word. The function delegates the
// bit-level encoding to the assembler package: it translates the
// trace's surface syntax (R0..R7, STORE-with-args-swapped) into the
// assembler's input syntax (x0..x7, STORE rs2,offset(rs1)), lets
// the assembler do the encoding, and returns the resulting word.
func EncodeInstr(instr Instr) (uint32, error) {
	asmLine, err := traceToAsmLine(instr)
	if err != nil {
		return 0, err
	}
	word, err := assembler.ParseInstruction(asmLine)
	if err != nil {
		return 0, fmt.Errorf("encoder rejected %q: %w", asmLine, err)
	}
	return uint32(word), nil
}

// traceToAsmLine rewrites a parsed trace instruction into the
// assembler's mnemonics and argument conventions. The R0..R7 names
// become x0..x7. STORE has its two argument slots swapped because
// the Spec writes "Rs, offset(Rd)" while the assembler expects
// "rs2, offset(rs1)".
func traceToAsmLine(instr Instr) (string, error) {
	switch instr.Mnemonic {
	case "ADD", "SUB", "MUL", "DIV":
		return rTypeAsmLine(instr)
	case "LOAD":
		return loadAsmLine(instr)
	case "STORE":
		return storeAsmLine(instr)
	}
	return "", fmt.Errorf("unsupported mnemonic: %q", instr.Mnemonic)
}

// rTypeAsmLine maps an R-type trace instruction to its assembler
// equivalent. ADD/SUB/MUL/DIV share the same R-Type encoding; the
// only difference is the funct3/funct7 pair, which the assembler
// owns.
func rTypeAsmLine(instr Instr) (string, error) {
	rd, err := traceRegToAsm(instr.Args[0])
	if err != nil {
		return "", err
	}
	rs1, err := traceRegToAsm(instr.Args[1])
	if err != nil {
		return "", err
	}
	rs2, err := traceRegToAsm(instr.Args[2])
	if err != nil {
		return "", err
	}
	return strings.ToLower(instr.Mnemonic) + " " + rd + ", " + rs1 + ", " + rs2, nil
}

// loadAsmLine maps a trace LOAD to an assembler LW. The mnemonic
// argument order is identical in the Spec and the assembler, so
// the only rewrite is R0..R7 -> x0..x7.
func loadAsmLine(instr Instr) (string, error) {
	rd, offset, base, err := splitMemRef(instr.Args[0], instr.Args[1])
	if err != nil {
		return "", err
	}
	rdAsm, err := traceRegToAsm(rd)
	if err != nil {
		return "", err
	}
	baseAsm, err := traceRegToAsm(base)
	if err != nil {
		return "", err
	}
	return "lw " + rdAsm + ", " + strconv.FormatInt(offset, 10) + "(" + baseAsm + ")", nil
}

// storeAsmLine maps a trace STORE to an assembler SW. The trace
// surface writes the operands as "Rs, offset(Rd)" but the
// assembler expects "rs2, offset(rs1)": the first argument in the
// trace is the value to store (rs2) and the second carries the
// base address (rs1). The swap happens here.
func storeAsmLine(instr Instr) (string, error) {
	value, offset, base, err := splitMemRef(instr.Args[0], instr.Args[1])
	if err != nil {
		return "", err
	}
	valueAsm, err := traceRegToAsm(value)
	if err != nil {
		return "", err
	}
	baseAsm, err := traceRegToAsm(base)
	if err != nil {
		return "", err
	}
	return "sw " + valueAsm + ", " + strconv.FormatInt(offset, 10) + "(" + baseAsm + ")", nil
}

// splitMemRef picks the value (Args[0] from the Instr) apart from
// the "offset(Rbase)" form (Args[1]). The two Args slots are kept
// separately during parsing so we can re-validate the base register
// independently from the value register.
func splitMemRef(reg, memRef string) (string, int64, string, error) {
	base, offset, err := parseMemRefBase(memRef)
	if err != nil {
		return "", 0, "", err
	}
	return reg, offset, base, nil
}

// parseMemRefBase extracts "offset(Rbase)" into (offset, baseReg).
// The string is assumed to have been produced by parseMemRef and
// therefore already validated; this helper only re-splits.
func parseMemRefBase(memRef string) (string, int64, error) {
	open := strings.Index(memRef, "(")
	close := strings.LastIndex(memRef, ")")
	if open < 0 || close <= open+1 {
		return "", 0, fmt.Errorf("malformed memory reference: %q", memRef)
	}
	base := memRef[open+1 : close]
	offset, err := strconv.ParseInt(strings.TrimSpace(memRef[:open]), 10, 64)
	if err != nil {
		return "", 0, fmt.Errorf("malformed offset in %q: %w", memRef, err)
	}
	return base, offset, nil
}

// traceRegToAsm maps a single trace register name "R0".."R7" to
// the assembler's "x0".."x7" naming. Empty strings and other
// identifiers surface as errors so the encoder never silently
// produces a word for a malformed argument.
func traceRegToAsm(reg string) (string, error) {
	if len(reg) != 2 || reg[0] != 'R' {
		return "", fmt.Errorf("not a trace register name: %q", reg)
	}
	idx, err := strconv.Atoi(string(reg[1]))
	if err != nil || idx < 0 || idx > 7 {
		return "", fmt.Errorf("trace register index out of range: %q", reg)
	}
	return "x" + strconv.Itoa(idx), nil
}
