package assembler

import (
	"bufio"
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
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";") {
			// Skip empty lines and comments
			continue
		}
		instr, err := ParseInstruction(line)
		if err != nil {
			return nil, err
		}
		program = append(program, instr)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return program, nil
}
