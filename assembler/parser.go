package assembler

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// TODO: add a preparser to normalize different instruction formats, handle comments, labels, and pseudo-instructions like 'j'.
// The preparser should:
// - Remove or handle comments (lines starting with # or ;, or after a # or ; on a line)
// - Normalize whitespace and commas
// - Expand pseudo-instructions (e.g. "j label" -> "jal x0, label")
// - Extract and store labels (e.g. "loop: addi x1, x1, 1")
// - Map labels to instruction addresses for later resolution

// parseOperands parses operands with a regex and returns matches or an error.
func parseOperands(operands string, re *regexp.Regexp, mnemonic string) ([]string, error) {
	matches := re.FindStringSubmatch(operands)
	if matches == nil {
		return nil, fmt.Errorf("invalid %s operands: %q", mnemonic, operands)
	}
	return matches, nil
}

// parseInt parses a string as int64.
func parseInt(s string) int64 {
	v, _ := strconv.ParseInt(s, 10, 32)
	return v
}

// parseUint parses a string as uint32.
func parseUint(s string) uint32 {
	v, _ := strconv.ParseUint(s, 10, 32)
	return uint32(v)
}

// ParseInstruction parses a single RISC-V assembler instruction (e.g. "addi x1, x0, 5")
// and returns the corresponding encoded Instruction.
func ParseInstruction(line string) (Instruction, error) {
	line = strings.TrimSpace(line)
	if line == "" {
		return 0, fmt.Errorf("empty line")
	}

	parts := strings.Fields(line)
	if len(parts) < 1 {
		return 0, fmt.Errorf("invalid instruction: %q", line)
	}
	mnemonic := parts[0]
	operands := strings.Join(parts[1:], "")
	operands = strings.ReplaceAll(operands, " ", "")

	switch mnemonic {
	case "addi":
		re := regexp.MustCompile(`^x(\d+),x(\d+),(-?\d+)$`)
		m, err := parseOperands(operands, re, mnemonic)
		if err != nil {
			return 0, err
		}
		rd, rs1, imm := parseUint(m[1]), parseUint(m[2]), parseInt(m[3])
		if imm < -2048 || imm > 2047 {
			return 0, fmt.Errorf("immediate out of range for addi: %d", imm)
		}
		var instr Instruction
		instr.SetOpcode(OPCODE_I_TYPE)
		instr.SetRd(rd)
		instr.SetRs1(rs1)
		instr.SetFunct3(FUNCT3_ADDI)
		instr.SetImmI(int32(imm))
		return instr, nil
	case "add":
		re := regexp.MustCompile(`^x(\d+),x(\d+),x(\d+)$`)
		m, err := parseOperands(operands, re, mnemonic)
		if err != nil {
			return 0, err
		}
		rd, rs1, rs2 := parseUint(m[1]), parseUint(m[2]), parseUint(m[3])
		var instr Instruction
		instr.SetOpcode(OPCODE_R_TYPE)
		instr.SetRd(rd)
		instr.SetRs1(rs1)
		instr.SetRs2(rs2)
		instr.SetFunct3(FUNCT3_ADD_SUB)
		instr.SetFunct7(FUNCT7_ADD)
		return instr, nil
	case "beq", "bne":
		re := regexp.MustCompile(`^x(\d+),x(\d+),(-?\d+)$`)
		m, err := parseOperands(operands, re, mnemonic)
		if err != nil {
			return 0, err
		}
		rs1, rs2, imm := parseUint(m[1]), parseUint(m[2]), parseInt(m[3])
		if imm < -4096 || imm > 4095 {
			return 0, fmt.Errorf("immediate out of range for %s: %d", mnemonic, imm)
		}
		var instr Instruction
		instr.SetOpcode(OPCODE_BRANCH)
		instr.SetRs1(rs1)
		instr.SetRs2(rs2)
		if mnemonic == "beq" {
			instr.SetFunct3(FUNCT3_BEQ)
		} else {
			instr.SetFunct3(FUNCT3_BNE)
		}
		instr.SetImmB(int32(imm))
		return instr, nil
	case "jal":
		re := regexp.MustCompile(`^x(\d+),(-?\d+)$`)
		m, err := parseOperands(operands, re, mnemonic)
		if err != nil {
			return 0, err
		}
		rd, imm := parseUint(m[1]), parseInt(m[2])
		if imm < -(1<<20) || imm > (1<<20)-1 {
			return 0, fmt.Errorf("immediate out of range for jal: %d", imm)
		}
		var instr Instruction
		instr.SetOpcode(OPCODE_JAL)
		instr.SetRd(rd)
		instr.SetImmJ(int32(imm))
		return instr, nil
	case "jalr":
		re := regexp.MustCompile(`^x(\d+),(-?\d+)\(x(\d+)\)$`)
		m, err := parseOperands(operands, re, mnemonic)
		if err != nil {
			return 0, err
		}
		rd, imm, rs1 := parseUint(m[1]), parseInt(m[2]), parseUint(m[3])
		if imm < -2048 || imm > 2047 {
			return 0, fmt.Errorf("immediate out of range for jalr: %d", imm)
		}
		var instr Instruction
		instr.SetOpcode(OPCODE_JALR)
		instr.SetRd(rd)
		instr.SetRs1(rs1)
		instr.SetFunct3(FUNCT3_JALR)
		instr.SetImmI(int32(imm))
		return instr, nil
	case "lw":
		re := regexp.MustCompile(`^x(\d+),(-?\d+)\(x(\d+)\)$`)
		m, err := parseOperands(operands, re, mnemonic)
		if err != nil {
			return 0, err
		}
		rd, imm, rs1 := parseUint(m[1]), parseInt(m[2]), parseUint(m[3])
		if imm < -2048 || imm > 2047 {
			return 0, fmt.Errorf("immediate out of range for lw: %d", imm)
		}
		var instr Instruction
		instr.SetOpcode(OPCODE_LOAD)
		instr.SetRd(rd)
		instr.SetRs1(rs1)
		instr.SetFunct3(FUNCT3_LW)
		instr.SetImmI(int32(imm))
		return instr, nil
	case "slli":
		re := regexp.MustCompile(`^x(\d+),x(\d+),(\d+)$`)
		m, err := parseOperands(operands, re, mnemonic)
		if err != nil {
			return 0, err
		}
		rd, rs1, shamt := parseUint(m[1]), parseUint(m[2]), parseUint(m[3])
		if shamt > 31 {
			return 0, fmt.Errorf("shift amount out of range for slli: %d", shamt)
		}
		var instr Instruction
		instr.SetOpcode(OPCODE_I_TYPE)
		instr.SetRd(rd)
		instr.SetRs1(rs1)
		instr.SetFunct3(FUNCT3_SLLI)
		instr.SetImmI(int32(shamt))
		return instr, nil
	case "slt":
		re := regexp.MustCompile(`^x(\d+),x(\d+),x(\d+)$`)
		m, err := parseOperands(operands, re, mnemonic)
		if err != nil {
			return 0, err
		}
		rd, rs1, rs2 := parseUint(m[1]), parseUint(m[2]), parseUint(m[3])
		var instr Instruction
		instr.SetOpcode(OPCODE_R_TYPE)
		instr.SetRd(rd)
		instr.SetRs1(rs1)
		instr.SetRs2(rs2)
		instr.SetFunct3(FUNCT3_SLT)
		instr.SetFunct7(0)
		return instr, nil
	case "sub":
		re := regexp.MustCompile(`^x(\d+),x(\d+),x(\d+)$`)
		m, err := parseOperands(operands, re, mnemonic)
		if err != nil {
			return 0, err
		}
		rd, rs1, rs2 := parseUint(m[1]), parseUint(m[2]), parseUint(m[3])
		var instr Instruction
		instr.SetOpcode(OPCODE_R_TYPE)
		instr.SetRd(rd)
		instr.SetRs1(rs1)
		instr.SetRs2(rs2)
		instr.SetFunct3(FUNCT3_ADD_SUB)
		instr.SetFunct7(FUNCT7_SUB)
		return instr, nil
	case "sw":
		re := regexp.MustCompile(`^x(\d+),(-?\d+)\(x(\d+)\)$`)
		m, err := parseOperands(operands, re, mnemonic)
		if err != nil {
			return 0, err
		}
		rs2, imm, rs1 := parseUint(m[1]), parseInt(m[2]), parseUint(m[3])
		if imm < -2048 || imm > 2047 {
			return 0, fmt.Errorf("immediate out of range for sw: %d", imm)
		}
		var instr Instruction
		instr.SetOpcode(OPCODE_STORE)
		instr.SetRs1(rs1)
		instr.SetRs2(rs2)
		instr.SetFunct3(FUNCT3_SW)
		instr.SetImmS(int32(imm))
		return instr, nil
	default:
		return 0, fmt.Errorf("unsupported instruction: %q", mnemonic)
	}
}
