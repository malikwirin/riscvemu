package trace

import (
	"strings"
	"testing"
)

// TestParseErrorString verifies that the error message contains the
// reason, the line number, and the offending line text. The exact
// format is part of the public contract; downstream code matches on
// substrings rather than full equality to stay robust against tweaks.
func TestParseErrorString(t *testing.T) {
	cases := []struct {
		name     string
		err      *ParseError
		mustHave []string
	}{
		{
			name:     "basic",
			err:      &ParseError{LineNum: 5, Line: "ADD R1, R2, R3", Reason: "register out of range"},
			mustHave: []string{"register out of range", "5", "ADD R1, R2, R3"},
		},
		{
			name:     "line_1",
			err:      &ParseError{LineNum: 1, Line: "v", Reason: "unsupported mnemonic"},
			mustHave: []string{"unsupported mnemonic", "1", "v"},
		},
		{
			name:     "line_42",
			err:      &ParseError{LineNum: 42, Line: "STORE R6, 8(R3)", Reason: "offset overflow"},
			mustHave: []string{"offset overflow", "42", "STORE R6, 8(R3)"},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			AssertContainsAll(t, tc.err.Error(), tc.mustHave...)
		})
	}
}

// TestLineKindDistinguishes verifies that a Line carrying an Instr
// reports LineInstr, and a Line carrying a Control reports
// LineControl. The other fields must be zero-valued respectively.
func TestLineKindDistinguishes(t *testing.T) {
	instr := Line{Kind: LineInstr, LineNum: 3, Instr: Instr{Mnemonic: "ADD", Args: [3]string{"R1", "R2", "R3"}}}
	if instr.Kind != LineInstr {
		t.Errorf("Kind = %d, want LineInstr", instr.Kind)
	}
	if instr.Instr.Mnemonic != "ADD" {
		t.Errorf("Instr.Mnemonic = %q, want ADD", instr.Instr.Mnemonic)
	}
	if instr.Control.Op != 0 {
		t.Errorf("Control.Op = %d, want 0", instr.Control.Op)
	}

	ctrl := Line{Kind: LineControl, LineNum: 7, Control: Control{Op: 's'}}
	if ctrl.Kind != LineControl {
		t.Errorf("Kind = %d, want LineControl", ctrl.Kind)
	}
	if ctrl.Control.Op != 's' {
		t.Errorf("Control.Op = %c, want s", ctrl.Control.Op)
	}
	if ctrl.Instr.Mnemonic != "" {
		t.Errorf("Instr.Mnemonic = %q, want empty", ctrl.Instr.Mnemonic)
	}
}

// TestLineKindValues pins the two LineKind constants. If a future
// change ever needs to reorder or insert values, this test forces a
// conscious update instead of a silent regression.
func TestLineKindValues(t *testing.T) {
	if LineInstr != 0 {
		t.Errorf("LineInstr = %d, want 0", LineInstr)
	}
	if LineControl != 1 {
		t.Errorf("LineControl = %d, want 1", LineControl)
	}
}

// TestInstrArgsOrder is a documentation test: the project
// specification's STORE mnemonic lists its operands as
// "Rs, offset(Rd)" while RISC-V's STORE convention is
// "rs2, offset(rs1)". Downstream code (the encoder in A.4, and the
// driver in a later change) must swap the arguments when crossing
// the trace/encoding boundary. This test pins the on-the-wire
// trace order so the swap happens in exactly one place.
func TestInstrArgsOrder(t *testing.T) {
	cases := []struct {
		name        string
		mnemonic    string
		args        [3]string
		checkFirst  bool
		firstIsReg  string
		checkSecond bool
		secondIsMem string
	}{
		{
			name:        "ADD_three_registers",
			mnemonic:    "ADD",
			args:        [3]string{"R1", "R2", "R3"},
			checkFirst:  true,
			firstIsReg:  "R1",
			checkSecond: true,
			secondIsMem: "R2",
		},
		{
			name:        "LOAD_destination_then_memref",
			mnemonic:    "LOAD",
			args:        [3]string{"R6", "4(R2)"},
			checkFirst:  true,
			firstIsReg:  "R6",
			checkSecond: true,
			secondIsMem: "4(R2)",
		},
		{
			name:        "STORE_source_then_memref",
			mnemonic:    "STORE",
			args:        [3]string{"R6", "8(R3)"},
			checkFirst:  true,
			firstIsReg:  "R6",
			checkSecond: true,
			secondIsMem: "8(R3)",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			instr := Instr{Mnemonic: tc.mnemonic, Args: tc.args}
			if tc.checkFirst && instr.Args[0] != tc.firstIsReg {
				t.Errorf("Args[0] = %q, want %q", instr.Args[0], tc.firstIsReg)
			}
			if tc.checkSecond && instr.Args[1] != tc.secondIsMem {
				t.Errorf("Args[1] = %q, want %q", instr.Args[1], tc.secondIsMem)
			}
		})
	}
}

// TestStubsCompile verifies that the empty ParseTrace / ParseTraceFile
// stubs exist with the right signatures and return nil nil at this
// stage. Later changes replace the bodies.
func TestStubsCompile(t *testing.T) {
	if _, err := ParseTrace(""); err != nil {
		t.Errorf("ParseTrace empty stub returned error: %v", err)
	}
}

// TestParseTraceFileEmptyStub checks that an empty path is
// reported via *ParseError with LineNum 0, so callers can
// distinguish I/O errors from syntax errors.
func TestParseTraceFileEmptyStub(t *testing.T) {
	_, err := ParseTraceFile("")
	if err == nil {
		t.Fatal("expected error for empty path, got nil")
	}
	pe, ok := err.(*ParseError)
	if !ok {
		t.Fatalf("error type = %T, want *ParseError", err)
	}
	if pe.LineNum != 0 {
		t.Errorf("LineNum = %d, want 0", pe.LineNum)
	}
}

// TestParseRType exercises the R-type subset of the Spec.
func TestParseRType(t *testing.T) {
	cases := []struct {
		name     string
		input    string
		wantMnem string
		wantArgs [3]string
	}{
		{"ADD_basic", "ADD R1, R2, R3", "ADD", [3]string{"R1", "R2", "R3"}},
		{"SUB", "SUB R0, R7, R0", "SUB", [3]string{"R0", "R7", "R0"}},
		{"MUL", "MUL R4, R1, R5", "MUL", [3]string{"R4", "R1", "R5"}},
		{"DIV", "DIV R0, R7, R1", "DIV", [3]string{"R0", "R7", "R1"}},
		{"ADD_no_spaces_around_commas", "ADD R1,R2,R3", "ADD", [3]string{"R1", "R2", "R3"}},
		{"ADD_extra_spaces", "ADD    R1,  R2,   R3", "ADD", [3]string{"R1", "R2", "R3"}},
		{"lowercase_normalised", "add r1, r2, r3", "ADD", [3]string{"R1", "R2", "R3"}},
		{"mixed_case_normalised", "Add r1, R2, r3", "ADD", [3]string{"R1", "R2", "R3"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			lines, err := ParseTrace(tc.input)
			if err != nil {
				t.Fatalf("ParseTrace: %v", err)
			}
			if len(lines) != 1 {
				t.Fatalf("got %d lines, want 1", len(lines))
			}
			got := lines[0]
			if got.Kind != LineInstr {
				t.Errorf("Kind = %d, want LineInstr", got.Kind)
			}
			if got.Instr.Mnemonic != tc.wantMnem {
				t.Errorf("Mnemonic = %q, want %q", got.Instr.Mnemonic, tc.wantMnem)
			}
			if got.Instr.Args != tc.wantArgs {
				t.Errorf("Args = %v, want %v", got.Instr.Args, tc.wantArgs)
			}
		})
	}
}

// TestParseRTypeRejects verifies that R-type parsing fails cleanly
// when the operand string is malformed.
func TestParseRTypeRejects(t *testing.T) {
	cases := []struct {
		name      string
		input     string
		wantSubst string
	}{
		{"missing_comma", "ADD R1 R2 R3", "invalid R-type operands"},
		{"too_few_args", "ADD R1, R2", "invalid R-type operands"},
		{"too_many_args", "ADD R1, R2, R3, R4", "invalid R-type operands"},
		{"register_out_of_range", "ADD R1, R2, R8", "invalid R-type operands"},
		{"low_register_digit", "ADD Rx, R2, R3", "invalid R-type operands"},
		{"no_register_prefix", "ADD 1, 2, 3", "invalid R-type operands"},
		{"empty_operands", "ADD", "expected mnemonic and operands"},
		{"whitespace_only_operands", "ADD   ", "expected mnemonic and operands"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := ParseTrace(tc.input)
			if err == nil {
				t.Fatalf("expected error, got nil")
			}
			pe, ok := err.(*ParseError)
			if !ok {
				t.Fatalf("error type = %T, want *ParseError", err)
			}
			if !strings.Contains(pe.Reason, tc.wantSubst) {
				t.Errorf("Reason = %q, want substring %q", pe.Reason, tc.wantSubst)
			}
			if pe.LineNum != 1 {
				t.Errorf("LineNum = %d, want 1", pe.LineNum)
			}
		})
	}
}

// TestParseRejectsUnknownMnemonic covers BEQ, JAL, ADDI, and other
// RISC-V instructions that are valid encodings but outside the Spec
// subset. The driver (later change) only consumes what the parser
// produces, so rejecting here is the right place to enforce the
// Spec's instruction list.
func TestParseRejectsUnknownMnemonic(t *testing.T) {
	cases := []string{
		"BEQ R1, R2, 4",
		"BNE R1, R2, 4",
		"ADDI R1, R0, 1",
		"JAL R0, 100",
		"JALR R1, 0(R2)",
		"SW R1, 0(R2)",
		"LW R1, 0(R2)",
	}
	for _, in := range cases {
		t.Run(in, func(t *testing.T) {
			_, err := ParseTrace(in)
			if err == nil {
				t.Fatalf("expected error for %q, got nil", in)
			}
			pe, ok := err.(*ParseError)
			if !ok {
				t.Fatalf("error type = %T, want *ParseError", err)
			}
			if !strings.Contains(pe.Reason, "unsupported mnemonic") {
				t.Errorf("Reason = %q, want substring %q", pe.Reason, "unsupported mnemonic")
			}
		})
	}
}

// TestParseSkipsEmptyAndComment verifies that blank lines and lines
// starting with '#' do not produce a Line and do not consume a line
// number for error reporting.
func TestParseSkipsEmptyAndComment(t *testing.T) {
	input := `# A comment line
ADD R1, R2, R3

# Another comment after a blank line
SUB R4, R5, R6
`
	lines, err := ParseTrace(input)
	if err != nil {
		t.Fatalf("ParseTrace: %v", err)
	}
	if len(lines) != 2 {
		t.Fatalf("got %d lines, want 2", len(lines))
	}
	if lines[0].LineNum != 2 {
		t.Errorf("first instr LineNum = %d, want 2 (after comment line)", lines[0].LineNum)
	}
	if lines[1].LineNum != 5 {
		t.Errorf("second instr LineNum = %d, want 5 (after blank + comment)", lines[1].LineNum)
	}
}

// TestParseUnknownControlRejects verifies that a single-character
// line that is not one of v/s/h/i is a parse error.
func TestParseUnknownControlRejects(t *testing.T) {
	for _, in := range []string{"x", "V", "!"} {
		t.Run(in, func(t *testing.T) {
			_, err := ParseTrace(in)
			if err == nil {
				t.Fatalf("expected error for %q, got nil", in)
			}
			pe, ok := err.(*ParseError)
			if !ok {
				t.Fatalf("error type = %T, want *ParseError", err)
			}
			if !strings.Contains(pe.Reason, "unknown control command") {
				t.Errorf("Reason = %q, want substring %q", pe.Reason, "unknown control command")
			}
		})
	}
}

// TestParseControl verifies that v/s/h/i are accepted as control
// commands and the rest of the line is empty. The toggle semantics
// of 'v' live in the driver, not the parser.
func TestParseControl(t *testing.T) {
	for _, op := range []byte{'v', 's', 'h', 'i'} {
		t.Run(string(op), func(t *testing.T) {
			lines, err := ParseTrace(string(op))
			if err != nil {
				t.Fatalf("ParseTrace: %v", err)
			}
			if len(lines) != 1 {
				t.Fatalf("got %d lines, want 1", len(lines))
			}
			if lines[0].Kind != LineControl {
				t.Errorf("Kind = %d, want LineControl", lines[0].Kind)
			}
			if lines[0].Control.Op != op {
				t.Errorf("Control.Op = %c, want %c", lines[0].Control.Op, op)
			}
		})
	}
}

// TestParseLineNumberOnError pins that the LineNum reported in a
// ParseError matches the offending line, including when the error
// sits several lines after comments and blank lines.
func TestParseLineNumberOnError(t *testing.T) {
	input := `# header
ADD R1, R2, R3

# another comment
BEQ R1, R2, 4
`
	_, err := ParseTrace(input)
	pe, ok := err.(*ParseError)
	if !ok {
		t.Fatalf("error type = %T, want *ParseError", err)
	}
	if pe.LineNum != 5 {
		t.Errorf("LineNum = %d, want 5", pe.LineNum)
	}
}

// TestParseMultiLine is a smoke test for the full Spec example
// shape, restricted to the R-type subset (LOAD/STORE come in A.3,
// control commands are covered by TestParseControl).
func TestParseMultiLine(t *testing.T) {
	input := `ADD R1, R2, R3
MUL R4, R1, R5
SUB R7, R6, R4
DIV R0, R7, R1
`
	lines, err := ParseTrace(input)
	if err != nil {
		t.Fatalf("ParseTrace: %v", err)
	}
	if len(lines) != 4 {
		t.Fatalf("got %d lines, want 4", len(lines))
	}
	wantMnems := []string{"ADD", "MUL", "SUB", "DIV"}
	for i, want := range wantMnems {
		if lines[i].Instr.Mnemonic != want {
			t.Errorf("line %d: Mnemonic = %q, want %q", i, lines[i].Instr.Mnemonic, want)
		}
		if lines[i].LineNum != i+1 {
			t.Errorf("line %d: LineNum = %d, want %d", i, lines[i].LineNum, i+1)
		}
	}
}

// TestParseLOADSTORE exercises the LOAD and STORE operand parsers.
// The two mnemonics share the same surface syntax; the difference
// is which register is the destination (LOAD) versus the value
// (STORE). The Spec writes the operands as "Rd, offset(Rs)" for
// LOAD and "Rs, offset(Rd)" for STORE; the actual RISC-V STORE
// argument order is the inverse, but that inversion is the
// encoder's job (A.4), not the parser's.
func TestParseLOADSTORE(t *testing.T) {
	cases := []struct {
		name       string
		input      string
		wantReg    string
		wantMemRef string
	}{
		{"LOAD_basic", "LOAD R6, 4(R2)", "R6", "4(R2)"},
		{"LOAD_negative_offset", "LOAD R3, -4(R5)", "R3", "-4(R5)"},
		{"LOAD_zero_offset", "LOAD R0, 0(R0)", "R0", "0(R0)"},
		{"LOAD_positive_offset_max", "LOAD R7, 2047(R1)", "R7", "2047(R1)"},
		{"LOAD_negative_offset_min", "LOAD R7, -2048(R1)", "R7", "-2048(R1)"},
		{"STORE_basic", "STORE R6, 8(R3)", "R6", "8(R3)"},
		{"STORE_negative_offset", "STORE R1, -16(R4)", "R1", "-16(R4)"},
		{"STORE_zero_offset", "STORE R0, 0(R0)", "R0", "0(R0)"},
		{"lowercase_normalised", "load r6, 4(r2)", "R6", "4(R2)"},
		{"mixed_case_normalised", "Load R6, 4(r2)", "R6", "4(R2)"},
		{"whitespace_tolerated", "LOAD  R6 , 4 ( R2 )", "R6", "4(R2)"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			lines, err := ParseTrace(tc.input)
			if err != nil {
				t.Fatalf("ParseTrace: %v", err)
			}
			if len(lines) != 1 {
				t.Fatalf("got %d lines, want 1", len(lines))
			}
			got := lines[0].Instr
			if got.Mnemonic != mnemonicFromInput(tc.input) {
				t.Errorf("Mnemonic = %q, want %q", got.Mnemonic, mnemonicFromInput(tc.input))
			}
			if got.Args[0] != tc.wantReg {
				t.Errorf("Args[0] = %q, want %q", got.Args[0], tc.wantReg)
			}
			if got.Args[1] != tc.wantMemRef {
				t.Errorf("Args[1] = %q, want %q", got.Args[1], tc.wantMemRef)
			}
		})
	}
}

// TestParseLOADSTORERejects verifies that malformed memory-reference
// operands surface a ParseError.
func TestParseLOADSTORERejects(t *testing.T) {
	cases := []struct {
		name      string
		input     string
		wantSubst string
	}{
		{"missing_parenthesis", "LOAD R6, 4 R2", "invalid memory-reference operands"},
		{"missing_closing_paren", "LOAD R6, 4(R2", "invalid memory-reference operands"},
		{"missing_opening_paren", "LOAD R6, 4R2)", "invalid memory-reference operands"},
		{"register_out_of_range", "LOAD R6, 4(R8)", "invalid memory-reference operands"},
		{"offset_too_large", "LOAD R6, 2048(R2)", "offset out of range"},
		{"offset_too_negative", "LOAD R6, -2049(R2)", "offset out of range"},
		{"empty_operands", "LOAD", "expected mnemonic and operands"},
		{"missing_comma", "LOAD R6 4(R2)", "invalid memory-reference operands"},
		{"non_digit_offset", "LOAD R6, foo(R2)", "invalid memory-reference operands"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := ParseTrace(tc.input)
			if err == nil {
				t.Fatalf("expected error for %q, got nil", tc.input)
			}
			pe, ok := err.(*ParseError)
			if !ok {
				t.Fatalf("error type = %T, want *ParseError", err)
			}
			if !strings.Contains(pe.Reason, tc.wantSubst) {
				t.Errorf("Reason = %q, want substring %q", pe.Reason, tc.wantSubst)
			}
		})
	}
}

// mnemonicFromInput extracts the upper-cased mnemonic from a trace
// line, used by the LOAD/STORE tests to assert Mnemonic.
func mnemonicFromInput(line string) string {
	parts := strings.Fields(line)
	if len(parts) == 0 {
		return ""
	}
	return strings.ToUpper(parts[0])
}

// TestParseSpecExampleSubset is a final smoke test of the Spec
// example shape with LOAD/STORE and a control command mixed in.
func TestParseSpecExampleSubset(t *testing.T) {
	input := `LOAD R6, 4(R2)
STORE R6, 8(R3)
s
`
	lines, err := ParseTrace(input)
	if err != nil {
		t.Fatalf("ParseTrace: %v", err)
	}
	if len(lines) != 3 {
		t.Fatalf("got %d lines, want 3", len(lines))
	}
	if lines[0].Kind != LineInstr || lines[0].Instr.Mnemonic != "LOAD" {
		t.Errorf("line 0: kind=%d mnem=%q, want LineInstr/LOAD", lines[0].Kind, lines[0].Instr.Mnemonic)
	}
	if lines[1].Kind != LineInstr || lines[1].Instr.Mnemonic != "STORE" {
		t.Errorf("line 1: kind=%d mnem=%q, want LineInstr/STORE", lines[1].Kind, lines[1].Instr.Mnemonic)
	}
	if lines[2].Kind != LineControl || lines[2].Control.Op != 's' {
		t.Errorf("line 2: kind=%d op=%c, want LineControl/'s'", lines[2].Kind, lines[2].Control.Op)
	}
}
