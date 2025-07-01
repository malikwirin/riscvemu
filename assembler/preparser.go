package assembler

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// preprocessPseudoInstructions rewrites pseudoinstructions (like "j label") to real instructions.
func preprocessPseudoInstructions(line string) string {
	fields := strings.Fields(line)
	if len(fields) == 0 {
		return line
	}
	switch fields[0] {
	case "j":
		if len(fields) == 2 {
			return fmt.Sprintf("jal x0, %s", fields[1])
		}
	}
	return line
}

// ReplaceLabelOperandWithOffset replaces a label operand in a branch or jump instruction
// with the correct PC-relative offset using the provided label mapping.
// idx is the instruction index (not byte address).
func ReplaceLabelOperandWithOffset(line string, idx int, labelMap map[string]int) (string, error) {
	fields := strings.Fields(line)
	if len(fields) == 0 {
		return line, nil
	}
	mnemonic := fields[0]

	// Only process branch and jump instructions with a label as the last operand
	needsLabel := false
	var labelOperandIdx int
	switch mnemonic {
	case "beq", "bne", "blt": // remember to add more mnemonics as needed
		if len(fields) == 4 {
			needsLabel = true
			labelOperandIdx = 3
		}
	case "jal":
		if len(fields) == 3 {
			needsLabel = true
			labelOperandIdx = 2
		}
	}

	if !needsLabel {
		return line, nil
	}

	label := fields[labelOperandIdx]

	// If it's already a number, keep as is
	if _, err := strconv.Atoi(label); err == nil || strings.HasPrefix(label, "-") {
		return line, nil
	}

	// Lookup label address in bytes
	targetAddr, ok := labelMap[label]
	if !ok {
		return "", fmt.Errorf("unknown label: %q", label)
	}
	curAddr := idx * INSTRUCTION_SIZE
	offset := targetAddr - curAddr

	// For branches and jumps, replace label with offset (as string)
	fields[labelOperandIdx] = fmt.Sprintf("%d", offset)
	return strings.Join(fields, " "), nil
}

// AssembleFile reads an assembler source file and returns a slice of Instructions.
func AssembleFile(filename string) ([]Instruction, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var rawLines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		rawLines = append(rawLines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	labelMap, instructions := parseLabelsAndInstructions(rawLines)
	var program []Instruction
	for idx, line := range instructions {
		line, err = ReplaceLabelOperandWithOffset(preprocessPseudoInstructions(line), idx, labelMap)
		if err != nil {
			return nil, fmt.Errorf("%s:%d: %w", filename, idx+1, err)
		}
		instr, err := ParseInstruction(line)
		if err != nil {
			return nil, fmt.Errorf("%s:%d: %w", filename, idx+1, err)
		}
		program = append(program, instr)
	}
	return program, nil
}

// parseLabelsAndInstructions parses a list of assembler source lines.
// It returns a label map (label name -> address in bytes) and a slice of instruction lines (labels stripped).
func parseLabelsAndInstructions(lines []string) (map[string]int, []string) {
	labelMap := make(map[string]int)
	var instructions []string
	instrIndex := 0

	for _, rawLine := range lines {
		labels, instr := splitLabelsAndInstruction(rawLine)
		for _, label := range labels {
			labelMap[label] = instrIndex * INSTRUCTION_SIZE
		}
		if instr == "" {
			continue
		}
		instructions = append(instructions, instr)
		instrIndex++
	}

	return labelMap, instructions
}
