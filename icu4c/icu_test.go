package icu4c

import "testing"

func TestVersion(t *testing.T) {
	v := Version()
	t.Logf("linked ICU version: %s", v)
	if v != "78.3" {
		t.Fatalf("ICU version = %q, want 78.3", v)
	}
}
