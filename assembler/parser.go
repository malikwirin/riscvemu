package assembler

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// TODO: add preparser to bring different instruction formats to a common form also to handle comments and labels
// ParseInstruction parses a single RISC-V assembler instruction (one line, e.g. "addi x1, x0, 5")
// and returns the corresponding encoded Instruction.
func ParseInstruction(line string) (Instruction, error) {
	line = strings.TrimSpace(line)
	if line == "" {
		return 0, fmt.Errorf("empty line")
	}

	// ectracting the mnemonic and operands
	parts := strings.Fields(line)
	if len(parts) < 1 {
		return 0, fmt.Errorf("invalid instruction: %q", line)
	}
	mnemonic := parts[0]
	operands := strings.Join(parts[1:], "")
	operands = strings.ReplaceAll(operands, " ", "")

	switch mnemonic {
	case "addi":
		// Format: addi rd, rs1, imm
		re := regexp.MustCompile(`^x(\d+),x(\d+),(-?\d+)$`)
		matches := re.FindStringSubmatch(operands)
		if matches == nil {
			return 0, fmt.Errorf("invalid addi operands: %q", operands)
		}
		rd, _ := strconv.ParseUint(matches[1], 10, 32)
		rs1, _ := strconv.ParseUint(matches[2], 10, 32)
		imm, _ := strconv.ParseInt(matches[3], 10, 32)
		if imm < -2048 || imm > 2047 {
			return 0, fmt.Errorf("immediate out of range for addi: %d", imm)
		}

		var instr Instruction
		instr.SetOpcode(OPCODE_I_TYPE)
		instr.SetRd(uint32(rd))
		instr.SetRs1(uint32(rs1))
		instr.SetFunct3(FUNCT3_ADDI)
		// Immediate: bits 20-31
		instr = instr | Instruction((uint32(imm)&0xFFF)<<20)
		return instr, nil
	case "add":
		// Format: add rd, rs1, rs2
		re := regexp.MustCompile(`^x(\d+),x(\d+),x(\d+)$`)
		matches := re.FindStringSubmatch(operands)
		if matches == nil {
			return 0, fmt.Errorf("invalid add operands: %q", operands)
		}
		rd, _ := strconv.ParseUint(matches[1], 10, 32)
		rs1, _ := strconv.ParseUint(matches[2], 10, 32)
		rs2, _ := strconv.ParseUint(matches[3], 10, 32)

		var instr Instruction
		instr.SetOpcode(OPCODE_R_TYPE)
		instr.SetRd(uint32(rd))
		instr.SetRs1(uint32(rs1))
		instr.SetRs2(uint32(rs2))
		instr.SetFunct3(FUNCT3_ADD_SUB)
		instr.SetFunct7(FUNCT7_ADD)
		return instr, nil
	case "beq":
		// Format: beq rs1, rs2, imm
		// Example: beq x1, x2, 32
		re := regexp.MustCompile(`^x(\d+),x(\d+),(-?\d+)$`)
		matches := re.FindStringSubmatch(operands)
		if matches == nil {
			return 0, fmt.Errorf("invalid beq operands: %q", operands)
		}
		rs1, _ := strconv.ParseUint(matches[1], 10, 32)
		rs2, _ := strconv.ParseUint(matches[2], 10, 32)
		imm, _ := strconv.ParseInt(matches[3], 10, 32)
		if imm < -4096 || imm > 4095 {
			return 0, fmt.Errorf("immediate out of range for beq: %d", imm)
		}

		var instr Instruction
		instr.SetOpcode(OPCODE_BRANCH)
		instr.SetRs1(uint32(rs1))
		instr.SetRs2(uint32(rs2))
		instr.SetFunct3(FUNCT3_BEQ)
		// B-type immediate encoding
		// imm[12]    -> bit 31
		// imm[10:5]  -> bits 30:25
		// imm[4:1]   -> bits 11:8
		// imm[11]    -> bit 7
		// The immediate is shifted right by 1 (lowest bit always 0)
		immU := uint32(imm) & 0x1FFF
		instr = instr | Instruction(((immU>>12)&0x1)<<31) // bit 12
		instr = instr | Instruction(((immU>>5)&0x3F)<<25) // bits 10:5
		instr = instr | Instruction(((immU>>1)&0xF)<<8)   // bits 4:1
		instr = instr | Instruction(((immU>>11)&0x1)<<7)  // bit 11
		return instr, nil
	case "bne":
		// Format: bne rs1, rs2, imm
		re := regexp.MustCompile(`^x(\d+),x(\d+),(-?\d+)$`)
		matches := re.FindStringSubmatch(operands)
		if matches == nil {
			return 0, fmt.Errorf("invalid bne operands: %q", operands)
		}
		rs1, _ := strconv.ParseUint(matches[1], 10, 32)
		rs2, _ := strconv.ParseUint(matches[2], 10, 32)
		imm, _ := strconv.ParseInt(matches[3], 10, 32)
		if imm < -4096 || imm > 4095 {
			return 0, fmt.Errorf("immediate out of range for bne: %d", imm)
		}

		var instr Instruction
		instr.SetOpcode(OPCODE_BRANCH)
		instr.SetRs1(uint32(rs1))
		instr.SetRs2(uint32(rs2))
		instr.SetFunct3(FUNCT3_BNE)
		// B-type immediate encoding
		immU := uint32(imm) & 0x1FFF
		instr = instr | Instruction(((immU>>12)&0x1)<<31) // bit 12
		instr = instr | Instruction(((immU>>5)&0x3F)<<25) // bits 10:5
		instr = instr | Instruction(((immU>>1)&0xF)<<8)   // bits 4:1
		instr = instr | Instruction(((immU>>11)&0x1)<<7)  // bit 11
		return instr, nil
	case "jal":
		// Format: jal rd, imm
		// Example: jal x1, 2048
		re := regexp.MustCompile(`^x(\d+),(-?\d+)$`)
		matches := re.FindStringSubmatch(operands)
		if matches == nil {
			return 0, fmt.Errorf("invalid jal operands: %q", operands)
		}
		rd, _ := strconv.ParseUint(matches[1], 10, 32)
		imm, _ := strconv.ParseInt(matches[2], 10, 32)
		if imm < -(1<<20) || imm > (1<<20)-1 {
			return 0, fmt.Errorf("immediate out of range for jal: %d", imm)
		}

		var instr Instruction
		instr.SetOpcode(OPCODE_JAL)
		instr.SetRd(uint32(rd))
		// J-type immediate encoding:
		// instr[31]    = imm[20]
		// instr[30:21] = imm[10:1]
		// instr[20]    = imm[11]
		// instr[19:12] = imm[19:12]
		immU := uint32(imm)
		instr = instr | Instruction(((immU>>20)&0x1)<<31)  // bit 20 -> 31
		instr = instr | Instruction(((immU>>1)&0x3FF)<<21) // bits 10:1 -> 30:21
		instr = instr | Instruction(((immU>>11)&0x1)<<20)  // bit 11 -> 20
		instr = instr | Instruction(((immU>>12)&0xFF)<<12) // bits 19:12 -> 19:12
		return instr, nil
	case "lw":
		// Format: lw rd, imm(rs1)
		// Example: lw x5, 16(x6)
		re := regexp.MustCompile(`^x(\d+),(-?\d+)\(x(\d+)\)$`)
		matches := re.FindStringSubmatch(operands)
		if matches == nil {
			return 0, fmt.Errorf("invalid lw operands: %q", operands)
		}
		rd, _ := strconv.ParseUint(matches[1], 10, 32)
		imm, _ := strconv.ParseInt(matches[2], 10, 32)
		rs1, _ := strconv.ParseUint(matches[3], 10, 32)
		if imm < -2048 || imm > 2047 {
			return 0, fmt.Errorf("immediate out of range for lw: %d", imm)
		}

		var instr Instruction
		instr.SetOpcode(OPCODE_LOAD)
		instr.SetRd(uint32(rd))
		instr.SetRs1(uint32(rs1))
		instr.SetFunct3(FUNCT3_LW)
		// I-type immediate: bits 20-31
		instr = instr | Instruction((uint32(imm)&0xFFF)<<20)
		return instr, nil
	case "sw":
		// Format: sw rs2, imm(rs1)
		// Example: sw x7, 12(x8)
		re := regexp.MustCompile(`^x(\d+),(-?\d+)\(x(\d+)\)$`)
		matches := re.FindStringSubmatch(operands)
		if matches == nil {
			return 0, fmt.Errorf("invalid sw operands: %q", operands)
		}
		rs2, _ := strconv.ParseUint(matches[1], 10, 32)
		imm, _ := strconv.ParseInt(matches[2], 10, 32)
		rs1, _ := strconv.ParseUint(matches[3], 10, 32)
		if imm < -2048 || imm > 2047 {
			return 0, fmt.Errorf("immediate out of range for sw: %d", imm)
		}

		var instr Instruction
		instr.SetOpcode(OPCODE_STORE)
		instr.SetRs1(uint32(rs1))
		instr.SetRs2(uint32(rs2))
		instr.SetFunct3(FUNCT3_SW)
		// S-type immediate is split: imm[11:5] in bits 25-31, imm[4:0] in bits 7-11
		immU := uint32(imm) & 0xFFF
		instr = instr | Instruction((immU>>5)<<25)  // bits 25-31
		instr = instr | Instruction((immU&0x1F)<<7) // bits 7-11
		return instr, nil
	default:
		return 0, fmt.Errorf("unsupported instruction: %q", mnemonic)
	}
}
