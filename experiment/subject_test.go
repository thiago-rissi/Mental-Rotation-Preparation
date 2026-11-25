// Copyright (2026) Thiago Lizier Rissi

package main

import (
	"slices"
	"testing"
)

func TestHasSubjectFlag(t *testing.T) {
	for _, c := range []struct {
		args []string
		want bool
	}{
		{nil, false},
		{[]string{"-w", "-quick"}, false},
		{[]string{"-sheet"}, false},
		{[]string{"-s", "3"}, true},
		{[]string{"--s", "3"}, true},
		{[]string{"-w", "-s=3"}, true},
		{[]string{"--s=3"}, true},
		{[]string{"--", "-s", "3"}, false},
	} {
		if got := hasSubjectFlag(c.args); got != c.want {
			t.Errorf("hasSubjectFlag(%q) = %v, want %v", c.args, got, c.want)
		}
	}
}

func TestWithSubject(t *testing.T) {
	got, added := withSubject([]string{"prog", "-w"}, 123456)
	if !added || !slices.Equal(got, []string{"prog", "-s", "123456", "-w"}) {
		t.Errorf("no -s: got %q, added %v", got, added)
	}
	in := []string{"prog", "-s", "7", "-w"}
	got, added = withSubject(in, 123456)
	if added || !slices.Equal(got, in) {
		t.Errorf("-s given: got %q, added %v", got, added)
	}
}

func TestRandomSubject(t *testing.T) {
	for range 1000 {
		n := randomSubject()
		if n < randomSubjectMin || n > randomSubjectMax {
			t.Fatalf("randomSubject() = %d, outside %d..%d", n, randomSubjectMin, randomSubjectMax)
		}
		if _, err := BuildSession(DefaultConfig(), n, false); err != nil {
			t.Fatalf("BuildSession for random subject %d: %v", n, err)
		}
	}
}
