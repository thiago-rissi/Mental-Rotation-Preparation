// Copyright (2026) Thiago Lizier Rissi

package main

// Participants do not need a number of their own. When -s is not given, the
// program draws one at random and passes it on as if it had been typed, so
// that goxpyriment does not open its setup dialog to ask for it. The number
// sets the condition order and the response keys exactly as a typed one does
// (see ConditionOrder and MappingFor), so the counterbalancing is random: every
// order and key mapping is equally likely, but a given group of participants
// is not guaranteed to be balanced.
//
// The browser version does the same in webapp/index.html, which writes a
// random number into the page address (?s=...) before the program starts.

import (
	"math/rand/v2"
	"strconv"
	"strings"
)

// Random participant numbers have six digits, so they are never confused with
// numbers given by hand for tests (-s 1).
const (
	randomSubjectMin = 100000
	randomSubjectMax = 999999
)

func randomSubject() int {
	return randomSubjectMin + rand.IntN(randomSubjectMax-randomSubjectMin+1)
}

// hasSubjectFlag reports whether the command-line arguments (program name
// excluded) set -s, in any of the forms the flag package accepts.
func hasSubjectFlag(args []string) bool {
	for _, a := range args {
		if a == "--" {
			return false // the flags end here
		}
		if !strings.HasPrefix(a, "-") {
			continue
		}
		name := strings.TrimPrefix(strings.TrimPrefix(a, "-"), "-")
		if name == "s" || strings.HasPrefix(name, "s=") {
			return true
		}
	}
	return false
}

// withSubject returns args (program name first) with "-s n" added right after
// the program name when they do not already set -s, and whether it added it.
func withSubject(args []string, n int) ([]string, bool) {
	if len(args) == 0 || hasSubjectFlag(args[1:]) {
		return args, false
	}
	out := make([]string, 0, len(args)+2)
	out = append(out, args[0], "-s", strconv.Itoa(n))
	out = append(out, args[1:]...)
	return out, true
}
