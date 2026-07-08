package cpu

import "testing"

// TestRSTagString pins the three render cases for a rename
// tag: the NoTag sentinel renders as "-", a tag below the
// lsuBase renders as "ALU#N", and a tag at or above the
// lsuBase renders as "LSU#N".
func TestRSTagString(t *testing.T) {
	cases := []struct {
		name string
		tag  RSTag
		want string
	}{
		{"no tag", NoTag, "-"},
		{"ALU tag", RSTag(5), "ALU#5"},
		{"LSU tag", RSTag(1<<16 + 7), "LSU#7"},
	}
	for _, tc := range cases {
		if got := tc.tag.String(); got != tc.want {
			t.Errorf("%s: %s.String() = %q, want %q", tc.name, tc.tag, got, tc.want)
		}
	}
}
