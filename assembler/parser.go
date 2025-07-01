package assembler

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// --- Constants for immediate ranges and shifts ---
const (
	ITypeImmMin  = -2048
	ITypeImmMax  = 2047
	BTypeImmMin  = -4096
	BTypeImmMax  = 4095
	JTypeImmMin  = -(1 << 20)
	JTypeImmMax  = (1 << 20) - 1
	SLLIShamtMax = 31
)

// --- Shared regex patterns ---
var (
	reRType    = regexp.MustCompile(`^x(\d+),x(\d+),x(\d+)$`)
	reIType    = regexp.MustCompile(`^x(\d+),x(\d+),(-?\d+)$`)
	reBType    = regexp.MustCompile(`^x(\d+),x(\d+),(-?\d+)$`)
	reJal      = regexp.MustCompile(`^x(\d+),(-?\d+)$`)
	reJalrLwSw = regexp.MustCompile(`^x(\d+),(-?\d+)\(x(\d+)\)$`)
	reSlli     = regexp.MustCompile(`^x(\d+),x(\d+),(\d+)$`)
)

// --- Helper functions ---
func parseOperands(operands string, re *regexp.Regexp, mnemonic string) ([]string, error) {
	matches := re.FindStringSubmatch(operands)
	if matches == nil {
		return nil, fmt.Errorf("invalid %s operands: %q", mnemonic, operands)
	}
	return matches, nil
}

func parseInt(s string) int64 {
	v, _ := strconv.ParseInt(s, 10, 32)
	return v
}

func parseUint(s string) uint32 {
	v, _ := strconv.ParseUint(s, 10, 32)
	return uint32(v)
}

// --- Table-driven instruction patterns ---
type instrPattern struct {
	Mnemonics []string
	Regex     *regexp.Regexp
	Builder   func([]string, string) (Instruction, error)
}

var instrTable = []instrPattern{
	{
		Mnemonics: []string{"addi"},
		Regex:     reIType,
		Builder: func(m []string, _ string) (Instruction, error) {
			rd, rs1, imm := parseUint(m[1]), parseUint(m[2]), parseInt(m[3])
			if imm < ITypeImmMin || imm > ITypeImmMax {
				return 0, fmt.Errorf("immediate out of range for addi: %d", imm)
			}
			var instr Instruction
			instr.SetOpcode(OPCODE_I_TYPE)
			instr.SetRd(rd)
			instr.SetRs1(rs1)
			instr.SetFunct3(FUNCT3_ADDI)
			instr.SetImmI(int32(imm))
			return instr, nil
		},
	},
	{
		Mnemonics: []string{"add"},
		Regex:     reRType,
		Builder: func(m []string, _ string) (Instruction, error) {
			rd, rs1, rs2 := parseUint(m[1]), parseUint(m[2]), parseUint(m[3])
			var instr Instruction
			instr.SetOpcode(OPCODE_R_TYPE)
			instr.SetRd(rd)
			instr.SetRs1(rs1)
			instr.SetRs2(rs2)
			instr.SetFunct3(FUNCT3_ADD_SUB)
			instr.SetFunct7(FUNCT7_ADD)
			return instr, nil
		},
	},
	{
		Mnemonics: []string{"sub"},
		Regex:     reRType,
		Builder: func(m []string, _ string) (Instruction, error) {
			rd, rs1, rs2 := parseUint(m[1]), parseUint(m[2]), parseUint(m[3])
			var instr Instruction
			instr.SetOpcode(OPCODE_R_TYPE)
			instr.SetRd(rd)
			instr.SetRs1(rs1)
			instr.SetRs2(rs2)
			instr.SetFunct3(FUNCT3_ADD_SUB)
			instr.SetFunct7(FUNCT7_SUB)
			return instr, nil
		},
	},
	{
		Mnemonics: []string{"slt"},
		Regex:     reRType,
		Builder: func(m []string, _ string) (Instruction, error) {
			rd, rs1, rs2 := parseUint(m[1]), parseUint(m[2]), parseUint(m[3])
			var instr Instruction
			instr.SetOpcode(OPCODE_R_TYPE)
			instr.SetRd(rd)
			instr.SetRs1(rs1)
			instr.SetRs2(rs2)
			instr.SetFunct3(FUNCT3_SLT)
			instr.SetFunct7(0)
			return instr, nil
		},
	},
	{
		Mnemonics: []string{"beq", "bne", "blt"},
		Regex:     reBType,
		Builder: func(m []string, mnemonic string) (Instruction, error) {
			rs1, rs2, imm := parseUint(m[1]), parseUint(m[2]), parseInt(m[3])
			if imm < BTypeImmMin || imm > BTypeImmMax {
				return 0, fmt.Errorf("immediate out of range for %s: %d", mnemonic, imm)
			}
			var instr Instruction
			instr.SetOpcode(OPCODE_BRANCH)
			instr.SetRs1(rs1)
			instr.SetRs2(rs2)
			switch mnemonic {
			case "beq":
				instr.SetFunct3(FUNCT3_BEQ)
			case "bne":
				instr.SetFunct3(FUNCT3_BNE)
			case "blt":
				instr.SetFunct3(FUNCT3_SLT)
			}
			instr.SetImmB(int32(imm))
			return instr, nil
		},
	},
	{
		Mnemonics: []string{"jal"},
		Regex:     reJal,
		Builder: func(m []string, _ string) (Instruction, error) {
			rd, imm := parseUint(m[1]), parseInt(m[2])
			if imm < JTypeImmMin || imm > JTypeImmMax {
				return 0, fmt.Errorf("immediate out of range for jal: %d", imm)
			}
			var instr Instruction
			instr.SetOpcode(OPCODE_JAL)
			instr.SetRd(rd)
			instr.SetImmJ(int32(imm))
			return instr, nil
		},
	},
	{
		Mnemonics: []string{"jalr"},
		Regex:     reJalrLwSw,
		Builder: func(m []string, _ string) (Instruction, error) {
			rd, imm, rs1 := parseUint(m[1]), parseInt(m[2]), parseUint(m[3])
			if imm < ITypeImmMin || imm > ITypeImmMax {
				return 0, fmt.Errorf("immediate out of range for jalr: %d", imm)
			}
			var instr Instruction
			instr.SetOpcode(OPCODE_JALR)
			instr.SetRd(rd)
			instr.SetRs1(rs1)
			instr.SetFunct3(FUNCT3_JALR)
			instr.SetImmI(int32(imm))
			return instr, nil
		},
	},
	{
		Mnemonics: []string{"lw"},
		Regex:     reJalrLwSw,
		Builder: func(m []string, _ string) (Instruction, error) {
			rd, imm, rs1 := parseUint(m[1]), parseInt(m[2]), parseUint(m[3])
			if imm < ITypeImmMin || imm > ITypeImmMax {
				return 0, fmt.Errorf("immediate out of range for lw: %d", imm)
			}
			var instr Instruction
			instr.SetOpcode(OPCODE_LOAD)
			instr.SetRd(rd)
			instr.SetRs1(rs1)
			instr.SetFunct3(FUNCT3_LW)
			instr.SetImmI(int32(imm))
			return instr, nil
		},
	},
	{
		Mnemonics: []string{"slli"},
		Regex:     reSlli,
		Builder: func(m []string, _ string) (Instruction, error) {
			rd, rs1, shamt := parseUint(m[1]), parseUint(m[2]), parseUint(m[3])
			if shamt > SLLIShamtMax {
				return 0, fmt.Errorf("shift amount out of range for slli: %d", shamt)
			}
			var instr Instruction
			instr.SetOpcode(OPCODE_I_TYPE)
			instr.SetRd(rd)
			instr.SetRs1(rs1)
			instr.SetFunct3(FUNCT3_SLLI)
			instr.SetImmI(int32(shamt))
			return instr, nil
		},
	},
	{
		Mnemonics: []string{"sw"},
		Regex:     reJalrLwSw,
		Builder: func(m []string, _ string) (Instruction, error) {
			rs2, imm, rs1 := parseUint(m[1]), parseInt(m[2]), parseUint(m[3])
			if imm < ITypeImmMin || imm > ITypeImmMax {
				return 0, fmt.Errorf("immediate out of range for sw: %d", imm)
			}
			var instr Instruction
			instr.SetOpcode(OPCODE_STORE)
			instr.SetRs1(rs1)
			instr.SetRs2(rs2)
			instr.SetFunct3(FUNCT3_SW)
			instr.SetImmS(int32(imm))
			return instr, nil
		},
	},
}

// --- Fast mnemonic lookup ---
var instrLookup = func() map[string]*instrPattern {
	m := make(map[string]*instrPattern)
	for i := range instrTable {
		for _, mn := range instrTable[i].Mnemonics {
			m[mn] = &instrTable[i]
		}
	}
	return m
}()

// --- Main parse function ---
func ParseInstruction(line string) (Instruction, error) {
	line = removeCommentAndTrim(line)
	if line == "" {
		return 0, fmt.Errorf("empty line")
	}

	parts := strings.Fields(line)
	if len(parts) < 1 {
		return 0, fmt.Errorf("invalid instruction: %q", line)
	}
	mnemonic := parts[0]
	operands := strings.Join(parts[1:], "")
	operands = removeAllWhitespace(operands)

	pattern, ok := instrLookup[mnemonic]
	if !ok {
		return 0, fmt.Errorf("unsupported instruction: %q", mnemonic)
	}
	matches, err := parseOperands(operands, pattern.Regex, mnemonic)
	if err != nil {
		return 0, err
	}
	return pattern.Builder(matches, mnemonic)
}
