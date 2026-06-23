package cpu

import "testing"

// TestRSTagStringNoTag pins that the NoTag sentinel renders
// as a dash, the same convention the rest of the TUI uses
// for "no tag" cells.
func TestRSTagStringNoTag(t *testing.T) {
	if got := NoTag.String(); got != "-" {
		t.Errorf("NoTag.String() = %q, want %q", got, "-")
	}
}

// TestRSTagStringALUTag pins that a tag in the ALU
// namespace (below the lsuBase) renders as "ALU#N".
func TestRSTagStringALUTag(t *testing.T) {
	tag := RSTag(5)
	if got := tag.String(); got != "ALU#5" {
		t.Errorf("RSTag(5).String() = %q, want %q", got, "ALU#5")
	}
}

// TestRSTagStringLSUTag pins that a tag in the LSU
// namespace (at or above the lsuBase) renders as "LSU#N".
func TestRSTagStringLSUTag(t *testing.T) {
	tag := RSTag(1<<16 + 7)
	if got := tag.String(); got != "LSU#7" {
		t.Errorf("RSTag(1<<16+7).String() = %q, want %q", got, "LSU#7")
	}
}
