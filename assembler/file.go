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
		line := scanner.Text()
		lineNum++
		// Remove comments (everything after # or ;)
		if idx := strings.IndexAny(line, "#;"); idx != -1 {
			line = line[:idx]
		}
		line = strings.TrimSpace(line)
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
