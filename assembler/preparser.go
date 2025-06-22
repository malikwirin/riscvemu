package assembler

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// AssembleFile reads an assembler source file and returns a slice of Instructions.
func AssembleFile(filename string) ([]Instruction, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var program []Instruction
	scanner := bufio.NewScanner(file)
	lineNum := 0
	for scanner.Scan() {
		line := removeCommentAndTrim(scanner.Text())
		lineNum++
		if line == "" {
			continue
		}
		// Handle label lines
		for {
			idx := strings.Index(line, ":")
			if idx == -1 {
				break
			}
			rest := strings.TrimSpace(line[idx+1:])
			if rest == "" {
				line = ""
				break
			}
			line = rest
		}
		if line == "" {
			continue
		}
		instr, err := ParseInstruction(line)
		if err != nil {
			return nil, fmt.Errorf("%s:%d: %w", filename, lineNum, err)
		}
		program = append(program, instr)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
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
		if instr == "" {
			continue
		}
		for _, label := range labels {
			labelMap[label] = instrIndex * INSTRUCTION_SIZE
		}
		instructions = append(instructions, instr)
		instrIndex++
	}

	return labelMap, instructions
}
