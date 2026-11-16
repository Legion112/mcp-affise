package app

import "testing"

func TestVersion(t *testing.T) {
	if version == "" {
		t.Fatal("empty version")
	}
}
