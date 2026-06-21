package trace

import (
	"testing"

	"github.com/malikwirin/riscvemu/assembler"
)

// TestEncodeInstrRoundTrip exercises the encoder by parsing each
// trace instruction back through the assembler. Because the
// encoder delegates to assembler.ParseInstruction, the round trip
// is implicit: any successful EncodeInstr produces a word that the
// assembler can re-parse; any failure is caught by the assembler
// itself and surfaced through EncodeInstr.
//
// The expected values are typed via interface{} so the table can
// mix register indices (uint32) and immediates (int32) without
// per-case type boilerplate.
func TestEncodeInstrRoundTrip(t *testing.T) {
	cases := []struct {
		name     string
		input    string
		wantRd   uint32 // 0 means "do not check"
		wantRs1  uint32
		wantRs2  uint32
		wantImmI int32
		wantImmS int32
	}{
		{
			name:   "ADD",
			input:  "ADD R1, R2, R3",
			wantRd: 1, wantRs1: 2, wantRs2: 3,
		},
		{
			name:   "SUB",
			input:  "SUB R4, R5, R6",
			wantRd: 4, wantRs1: 5, wantRs2: 6,
		},
		{
			name:   "MUL",
			input:  "MUL R7, R0, R1",
			wantRd: 7, wantRs1: 0, wantRs2: 1,
		},
		{
			name:   "DIV",
			input:  "DIV R0, R7, R1",
			wantRd: 0, wantRs1: 7, wantRs2: 1,
		},
		{
			name:   "LOAD_positive_offset",
			input:  "LOAD R6, 4(R2)",
			wantRd: 6, wantRs1: 2, wantImmI: 4,
		},
		{
			name:   "LOAD_negative_offset",
			input:  "LOAD R3, -4(R5)",
			wantRd: 3, wantRs1: 5, wantImmI: -4,
		},
		{
			name:    "STORE_arg_swap",
			input:   "STORE R6, 8(R3)",
			wantRs1: 3, wantRs2: 6, wantImmS: 8,
		},
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
			word, err := EncodeInstr(lines[0].Instr)
			if err != nil {
				t.Fatalf("EncodeInstr: %v", err)
			}
			parsed := assembler.Instruction(word)
			if got := parsed.Opcode(); got == assembler.OPCODE_INVALID {
				t.Fatalf("assembler rejected word %#x: invalid opcode", word)
			}
			if tc.wantRd != 0 && parsed.Rd() != tc.wantRd {
				t.Errorf("Rd = %d, want %d", parsed.Rd(), tc.wantRd)
			}
			if tc.wantRs1 != 0 && parsed.Rs1() != tc.wantRs1 {
				t.Errorf("Rs1 = %d, want %d", parsed.Rs1(), tc.wantRs1)
			}
			if tc.wantRs2 != 0 && parsed.Rs2() != tc.wantRs2 {
				t.Errorf("Rs2 = %d, want %d", parsed.Rs2(), tc.wantRs2)
			}
			// Always check ImmI for LOAD, ImmS for STORE. The
			// default value of 0 is also a legitimate immediate,
			// so the test cannot distinguish "absent" from "zero"
			// from the field alone. We rely on the mnemonic to
			// pick which accessor to call.
			switch lines[0].Instr.Mnemonic {
			case "LOAD":
				if got := parsed.ImmI(); got != tc.wantImmI {
					t.Errorf("ImmI = %d, want %d", got, tc.wantImmI)
				}
			case "STORE":
				if got := parsed.ImmS(); got != tc.wantImmS {
					t.Errorf("ImmS = %d, want %d", got, tc.wantImmS)
				}
			}
		})
	}
}

// TestEncodeInstrRejectsUnknownMnemonic catches a regression where
// a future change adds a new mnemonic to supportedMnemonics but
// forgets to wire it through the encoder.
func TestEncodeInstrRejectsUnknownMnemonic(t *testing.T) {
	_, err := EncodeInstr(Instr{Mnemonic: "BEQ", Args: [3]string{"R1", "R2", "4"}})
	if err == nil {
		t.Fatal("expected error for unsupported mnemonic, got nil")
	}
}

// TestEncodeInstrRejectsBadRegister catches malformed register
// names that the parser would have rejected, but a defensive check
// in the encoder is cheap and makes the failure mode obvious.
func TestEncodeInstrRejectsBadRegister(t *testing.T) {
	_, err := EncodeInstr(Instr{Mnemonic: "ADD", Args: [3]string{"R8", "R1", "R2"}})
	if err == nil {
		t.Fatal("expected error for register out of range, got nil")
	}
}

// TestEncodeSpecExampleLines drives the encoder over the exact
// instruction lines from the Spec example (page 4 of the
// project document). LOAD and STORE get the argument-swap
// treatment; ADD/SUB/MUL/DIV are passed through unchanged. The
// test confirms the round trip through the assembler; it does not
// pin the exact instruction word because the assembler's encoding
// is the source of truth.
func TestEncodeSpecExampleLines(t *testing.T) {
	lines := []string{
		"ADD R1, R2, R3",
		"MUL R4, R1, R5",
		"LOAD R6, 4(R2)",
		"SUB R7, R6, R4",
		"DIV R0, R7, R1",
	}
	for _, line := range lines {
		t.Run(line, func(t *testing.T) {
			parsed, err := ParseTrace(line)
			if err != nil {
				t.Fatalf("ParseTrace: %v", err)
			}
			word, err := EncodeInstr(parsed[0].Instr)
			if err != nil {
				t.Fatalf("EncodeInstr: %v", err)
			}
			if word == 0 {
				t.Errorf("EncodeInstr returned zero word for %q", line)
			}
		})
	}
}
