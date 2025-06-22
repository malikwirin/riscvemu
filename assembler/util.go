package assembler

import (
	"bufio"
	"os"
	"strings"
	"unicode"
)

// removeAllWhitespace removes all whitespace characters from a string.
func removeAllWhitespace(s string) string {
	var b strings.Builder
	for _, r := range s {
		if !unicode.IsSpace(r) {
			b.WriteRune(r)
		}
	}
	return b.String()
}

// removeCommentAndTrim removes comments (everything after # or ;) and trims whitespace.
func removeCommentAndTrim(line string) string {
	if idx := strings.IndexAny(line, "#;"); idx != -1 {
		line = line[:idx]
	}
	return strings.TrimSpace(line)
}

// splitLabelsAndInstruction parses a line into labels and the instruction part.
// E.g. "foo: bar: addi x1, x0, 1" -> []{"foo", "bar"}, "addi x1, x0, 1"
func splitLabelsAndInstruction(line string) (labels []string, instr string) {
	line = removeCommentAndTrim(line)
	for {
		idx := strings.Index(line, ":")
		if idx == -1 {
			break
		}
		label := strings.TrimSpace(line[:idx])
		if label != "" {
			labels = append(labels, label)
		}
		line = strings.TrimSpace(line[idx+1:])
	}
	return labels, line
}

// linesFromFile reads all lines from a file and returns them as []string.
func linesFromFile(filename string) ([]string, error) {
	var lines []string
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return lines, nil
}
