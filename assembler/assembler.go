package assembler

import (
    "bufio"
    "os"
    "strings"
)

type Instruction struct {
    Op      string
    Args    []string
    RawLine string
}

func ParseAssemblyFile(filename string) ([]Instruction, error) {
    file, err := os.Open(filename)
    if err != nil {
        return nil, err
    }
    defer file.Close()

    var instructions []Instruction
    scanner := bufio.NewScanner(file)
    for scanner.Scan() {
        line := strings.TrimSpace(scanner.Text())
        if line == "" || strings.HasPrefix(line, "#") {
            continue
        }
        fields := strings.Fields(line)
        op := fields[0]
        args := strings.Split(strings.Join(fields[1:], ""), ",")
        for i, a := range args {
            args[i] = strings.TrimSpace(a)
        }
        instructions = append(instructions, Instruction{Op: op, Args: args, RawLine: line})
    }
    return instructions, scanner.Err()
}
