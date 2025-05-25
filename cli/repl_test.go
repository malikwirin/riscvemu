package cli

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/malikwirin/riscvemu/arch"
	"github.com/chzyer/readline"
)

type readCloser struct {
	*strings.Reader
}
func (r *readCloser) Close() error { return nil }

func runREPLWithInput(input string, machine *arch.Machine) string {
	stdin := &readCloser{strings.NewReader(input)}
	rl, _ := readline.NewEx(&readline.Config{
		Prompt:          "> ",
		Stdin:           stdin,
		Stdout:          os.Stdout,
		HistoryFile:     "",
	})
	repl := &REPL{
		machine: machine,
		rl:      rl,
	}
	// capture output
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	repl.Start()

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
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
