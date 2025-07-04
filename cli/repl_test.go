package cli

import (
	"strings"
	"testing"

	"github.com/chzyer/readline"
	"github.com/malikwirin/riscvemu/arch"
	"github.com/malikwirin/riscvemu/assembler"
	"github.com/stretchr/testify/assert"
)

// readCloser wraps a strings.Reader to implement io.ReadCloser for readline.
type readCloser struct {
	*strings.Reader
}

func (r *readCloser) Close() error { return nil }

// runREPLWithInput runs the REPL with the given input and returns the captured output.
func runREPLWithInput(input string, machine *arch.Machine) string {
	stdin := &readCloser{strings.NewReader(input)}
	rl, _ := readline.NewEx(&readline.Config{
		Prompt:      "> ",
		Stdin:       stdin,
		Stdout:      nil, // not used, we capture output below
		HistoryFile: "",
	})
	repl := &REPL{
		machine: machine,
		rl:      rl,
	}
	return captureOutput(func() {
		repl.Start()
	})
}

func TestREPL_Commands(t *testing.T) {
	m := arch.NewMachine(64)
	instr, _ := assembler.ParseInstruction("addi x0, x0, 0")
	m.Memory.Data[0] = byte(instr)
	m.Memory.Data[1] = byte(instr >> 8)
	m.Memory.Data[2] = byte(instr >> 16)
	m.Memory.Data[3] = byte(instr >> 24)
	input := "help\nfoobar\nstep\nquit\n"
	output := runREPLWithInput(input, m)

	assert.Contains(t, output, "Available commands", "missing help output")
	assert.Contains(t, output, "Unknown command", "missing unknown command output")
	assert.Contains(t, output, "Executed 1 step", "missing step output")
	assert.Contains(t, output, "Goodbye!", "missing quit output")
}

func TestREPL_ExitAlias(t *testing.T) {
	m := arch.NewMachine(64)
	input := "exit\n"
	output := runREPLWithInput(input, m)
	assert.Contains(t, output, "Goodbye!", "exit should quit the REPL")
}

func TestREPL_EmptyInput(t *testing.T) {
	m := arch.NewMachine(64)
	input := "\n\nquit\n"
	output := runREPLWithInput(input, m)
	assert.Contains(t, output, "Goodbye!", "missing goodbye for empty input test")
}
