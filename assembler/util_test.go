package assembler

import (
	"os"
	"reflect"
	"testing"
)

// Test removeCommentAndTrim strips comments and trims whitespace.
func TestRemoveCommentAndTrim(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{"   addi x1, x0, 5   # comment", "addi x1, x0, 5"},
		{"label: add x2, x1, x0 ; inline comment", "label: add x2, x1, x0"},
		{"   # only comment", ""},
		{"sw x3, 0(x1)", "sw x3, 0(x1)"},
		{"", ""},
		{"   ", ""},
	}
	for _, c := range cases {
		got := removeCommentAndTrim(c.in)
		if got != c.want {
			t.Errorf("removeCommentAndTrim(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

// Test splitLabelsAndInstruction splits multiple labels and returns the instruction part.
func TestSplitLabelsAndInstruction(t *testing.T) {
	cases := []struct {
		in        string
		wantLabs  []string
		wantInstr string
	}{
		{"foo: addi x1, x0, 1", []string{"foo"}, "addi x1, x0, 1"},
		{"foo: bar: add x2, x1, x0", []string{"foo", "bar"}, "add x2, x1, x0"},
		{"label:", []string{"label"}, ""},
		{"   addi x2, x3, 4", nil, "addi x2, x3, 4"},
		{"# just a comment", nil, ""},
		{"foo: bar: # comment", []string{"foo", "bar"}, ""},
	}
	for _, c := range cases {
		gotLabs, gotInstr := splitLabelsAndInstruction(c.in)
		if !reflect.DeepEqual(gotLabs, c.wantLabs) || gotInstr != c.wantInstr {
			t.Errorf("splitLabelsAndInstruction(%q) = %v, %q; want %v, %q",
				c.in, gotLabs, gotInstr, c.wantLabs, c.wantInstr)
		}
	}
}

// Test linesFromFile reads all lines from a file and returns them as []string.
func TestLinesFromFile(t *testing.T) {
	content := "addi x1, x0, 1\nadd x2, x1, x0\n# comment line\n"
	tmp, err := os.CreateTemp("", "utiltest-*.asm")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmp.Name())
	if _, err := tmp.WriteString(content); err != nil {
		t.Fatalf("failed to write to temp file: %v", err)
	}
	tmp.Close()

	lines, err := linesFromFile(tmp.Name())
	if err != nil {
		t.Fatalf("linesFromFile failed: %v", err)
	}
	want := []string{"addi x1, x0, 1", "add x2, x1, x0", "# comment line"}
	if !reflect.DeepEqual(lines, want) {
		t.Errorf("linesFromFile: got %v, want %v", lines, want)
	}
}

// Test removeAllWhitespace removes all whitespace characters from a string.
func TestRemoveAllWhitespace(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{" a b c  d\t e\nf", "abcdef"},
		{"\t  addi   x1,   x0,   5  ", "addix1,x0,5"},
		{"", ""},
		{"   ", ""},
		{"no_whitespace", "no_whitespace"},
	}
	for _, c := range cases {
		got := removeAllWhitespace(c.in)
		if got != c.want {
			t.Errorf("removeAllWhitespace(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}
