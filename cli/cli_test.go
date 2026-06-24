package cli

import (
	"testing"

	"codeberg.org/malik/riscvemu/arch/cpu"
)

// TestSplitSubcommand pins the subcommand detection: the
// first token is recognised when it matches a known
// subcommand (tui, repl, run, help), and the rest is
// returned unchanged. Anything else falls through to the
// CLI dispatch.
func TestSplitSubcommand(t *testing.T) {
	cases := []struct {
		name     string
		args     []string
		wantSub  string
		wantRest []string
	}{
		{"empty", nil, "", nil},
		{"tui subcommand", []string{"tui"}, "tui", nil},
		{"repl subcommand", []string{"repl"}, "repl", nil},
		{"run subcommand", []string{"run", "foo.s"}, "run", []string{"foo.s"}},
		{"help subcommand", []string{"help"}, "help", nil},
		{"positional falls through", []string{"foo.s"}, "", []string{"foo.s"}},
		{"positional with cycles falls through", []string{"foo.s", "100"}, "", []string{"foo.s", "100"}},
		{"subcommand with flags", []string{"tui", "-alu-rs", "8"}, "tui", []string{"-alu-rs", "8"}},
		{"flags before subcommand stays positional", []string{"-alu-rs", "8", "tui"}, "", []string{"-alu-rs", "8", "tui"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			gotSub, gotRest := splitSubcommand(tc.args)
			if gotSub != tc.wantSub {
				t.Errorf("sub = %q, want %q", gotSub, tc.wantSub)
			}
			if !equalStrings(gotRest, tc.wantRest) {
				t.Errorf("rest = %v, want %v", gotRest, tc.wantRest)
			}
		})
	}
}

// TestParseRunArgs pins the CLI-side flag parser: pipeline
// flags stay in their -name value form, the first
// positional is the program file, the second is the cycle
// count, and a third positional is rejected.
func TestParseRunArgs(t *testing.T) {
	t.Run("program only", func(t *testing.T) {
		_, _, prog, cycles, err := ParseRunArgs([]string{"foo.s"})
		if err != nil {
			t.Fatalf("ParseRunArgs: %v", err)
		}
		if prog != "foo.s" {
			t.Errorf("program = %q, want %q", prog, "foo.s")
		}
		if cycles != 0 {
			t.Errorf("cycles = %d, want 0", cycles)
		}
	})

	t.Run("program with cycle count", func(t *testing.T) {
		_, _, prog, cycles, err := ParseRunArgs([]string{"foo.s", "42"})
		if err != nil {
			t.Fatalf("ParseRunArgs: %v", err)
		}
		if prog != "foo.s" || cycles != 42 {
			t.Errorf("got prog=%q cycles=%d, want foo.s/42", prog, cycles)
		}
	})

	t.Run("flags before program", func(t *testing.T) {
		cfg, _, prog, _, err := ParseRunArgs([]string{"-alu-rs", "8", "foo.s"})
		if err != nil {
			t.Fatalf("ParseRunArgs: %v", err)
		}
		if cfg.ALURSCount != 8 {
			t.Errorf("ALURSCount = %d, want 8", cfg.ALURSCount)
		}
		if prog != "foo.s" {
			t.Errorf("program = %q, want foo.s", prog)
		}
	})

	t.Run("flags after program", func(t *testing.T) {
		_, _, prog, _, err := ParseRunArgs([]string{"foo.s", "-mem", "2048"})
		if err != nil {
			t.Fatalf("ParseRunArgs: %v", err)
		}
		if prog != "foo.s" {
			t.Errorf("program = %q, want foo.s", prog)
		}
	})

	t.Run("rejects extra positional", func(t *testing.T) {
		_, _, _, _, err := ParseRunArgs([]string{"foo.s", "100", "extra"})
		if err == nil {
			t.Error("want error for third positional argument, got nil")
		}
	})

	t.Run("rejects non-numeric cycle count", func(t *testing.T) {
		_, _, _, _, err := ParseRunArgs([]string{"foo.s", "abc"})
		if err == nil {
			t.Error("want error for non-numeric cycle count, got nil")
		}
	})

	t.Run("rejects unknown flag", func(t *testing.T) {
		_, _, _, _, err := ParseRunArgs([]string{"-unknown", "1", "foo.s"})
		if err == nil {
			t.Error("want error for unknown flag, got nil")
		}
	})
}

// TestRunUnknownSubcommand pins the dispatch error path
// for the positional-fallback: a non-flag token is
// treated as a program file, and a second non-numeric
// token is rejected as a malformed cycle count. The
// canonical "unknown subcommand" path applies only when
// a non-empty subcommand string is set that does not
// match any case; that path is exercised in
// TestSplitSubcommand above.
func TestRunUnknownSubcommand(t *testing.T) {
	err := Run([]string{"bogus", "foo.s"})
	if err == nil {
		t.Fatal("Run(bogus): want error, got nil")
	}
	if !contains(err.Error(), "cycle count") {
		t.Errorf("error = %q, want 'cycle count' (from the positional CLI path)", err)
	}
}

// TestRunHelp pins that the help subcommand returns nil
// (the caller exits cleanly) and that Run("") errors out
// because the run subcommand needs a program file.
func TestRunHelp(t *testing.T) {
	if err := Run([]string{"help"}); err != nil {
		t.Errorf("Run(help): %v", err)
	}
	if err := Run([]string{"-help"}); err != nil {
		t.Errorf("Run(-help): %v", err)
	}
	if err := Run([]string{"--help"}); err != nil {
		t.Errorf("Run(--help): %v", err)
	}
}

// TestRunEmptyPositional pins that a bare empty string is
// rejected by the pipeline parser. The "" token looks
// like a positional program file but ConfigFromFlags
// reports it as a malformed flag first.
func TestRunEmptyPositional(t *testing.T) {
	err := Run([]string{""})
	if err == nil {
		t.Fatal("Run(\"\"): want error, got nil")
	}
}

// TestRunParsesConfigFromFlags checks that pipeline-flag
// overrides work for the tui subcommand by parsing
// without entering the TUI (HelpRequested intercept).
func TestRunParsesConfigFromFlags(t *testing.T) {
	// -help is a sentinel that the real TUI/REPL would
	// normally short-circuit on, but here it lets us pin
	// the flag-parse path without starting an actual
	// frontend.
	if err := Run([]string{"tui", "-alu-rs", "8", "-help"}); err != nil {
		t.Errorf("Run(tui -alu-rs 8 -help): %v", err)
	}
}

// equalStrings compares two []string for element-wise
// equality. nil and empty slice are considered equal so
// the test cases do not have to distinguish them.
func equalStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// contains is a tiny indirection so the file does not
// need to import "strings" for one assertion.
func contains(s, sub string) bool {
	return len(s) >= len(sub) && indexOf(s, sub) >= 0
}

func indexOf(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}

// ensure cpu import is used (helpers in this file do not
// need it directly, but keeping the import signals that
// the package-level types from cpu are the canonical
// Config carrier for ParseRunArgs and RunScript).
var _ = cpu.SpecConfig
