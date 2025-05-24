package cli

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/malikwirin/riscvemu/arch"
	"github.com/chzyer/readline"
)

func runREPLWithInput(input string, machine *arch.Machine) string {
	rl, _ := readline.NewEx(&readline.Config{
		Prompt:          "> ",
		Stdin:           strings.NewReader(input),
		Stdout:          os.Stdout,
		HistoryFile:     "",
		DisableAutoSave: true,
	})
	repl := &REPL{
		machine: machine,
		rl:      rl,
	}
	// capture output
	var buf bytes.Buffer
	stdout := os.Stdout
	os.Stdout = &buf
	repl.Start()
	os.Stdout = stdout
	return buf.String()
}

func TestREPL_Commands(t *testing.T) {
	m := arch.NewMachine(64)
	input := "help\nfoobar\nstep\nquit\n"
	output := runREPLWithInput(input, m)

	if !strings.Contains(output, "Available commands") {
		t.Errorf("missing help output, got: %q", output)
	}
	if !strings.Contains(output, "Unknown command") {
		t.Errorf("missing unknown command output, got: %q", output)
	}
	if !strings.Contains(output, "Executed 1 step") {
		t.Errorf("missing step output, got: %q", output)
	}
	if !strings.Contains(output, "Goodbye!") {
		t.Errorf("missing quit output, got: %q", output)
	}
}

func TestREPL_ExitAlias(t *testing.T) {
	m := arch.NewMachine(64)
	input := "exit\n"
	output := runREPLWithInput(input, m)
	if !strings.Contains(output, "Goodbye!") {
		t.Errorf("exit should quit the REPL, got: %q", output)
	}
}

func TestREPL_EmptyInput(t *testing.T) {
	m := arch.NewMachine(64)
	input := "\n\nquit\n"
	output := runREPLWithInput(input, m)
	// Only goodbye should appear, nothing else
	if !strings.Contains(output, "Goodbye!") {
		t.Errorf("missing goodbye for empty input test, got: %q", output)
	}
}
