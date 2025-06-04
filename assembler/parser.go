package assembler

import (
    "fmt"
    "regexp"
    "strconv"
    "strings"
)

// TODO: add preparser to bring different instruction formats to a common form
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
        instr = instr | Instruction((immU>>5)<<25) // bits 25-31
        instr = instr | Instruction((immU&0x1F)<<7) // bits 7-11
        return instr, nil
    default:
        return 0, fmt.Errorf("unsupported instruction: %q", mnemonic)
    }
}
