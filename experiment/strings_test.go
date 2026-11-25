// Copyright (2026) Thiago Lizier Rissi

package main

import (
	"strings"
	"testing"
	"unicode"
)

// The English screens must match the specified wording exactly. Subject 1 has
// J for normal and F for mirrored; subject 2 the reverse.
func TestEnglishTextsExact(t *testing.T) {
	tx, err := TextsFor("en")
	if err != nil {
		t.Fatal(err)
	}
	m := MappingFor(1)
	cases := map[string][2]string{
		"welcome": {tx.Welcome(m), `Rotated Characters

In this experiment you will see letters and digits, one at a time, inside a circle.

Each character may be printed normally or as a mirror image.

Press J if it is normal and F if it is mirrored.

Respond as quickly as you can, but try not to make mistakes. The session takes about ten minutes.

Press the space bar to continue.`},
		"block types": {tx.BlockTypes, `The experiment is divided into blocks of two kinds.

In some blocks, the circle stays empty for a moment before the character appears.

In the other blocks, you first see the outline of the coming character, upright and in its normal form. An arrow then shows where the top of the character will point, and the character appears as soon as the arrow is gone.

Press the space bar to continue.`},
		"practice": {tx.Practice(6), `You will begin with six practice trials. Place your left index finger on F and your right index finger on J.

Press the space bar when you are ready.`},
		"after practice": {tx.AfterPractice(6, 24), `The practice is over. The experiment itself has six blocks of 24 trials, and you can rest between blocks.

Press the space bar to begin.`},
		"block N": {tx.BlockIntro(conditionCatalog["N"], 1, 6, m), `Block 1 of 6

In this block, the circle stays empty for a moment before the character appears.

Press J for normal and F for mirrored.

Press the space bar to begin.`},
		"block B after a break": {tx.BlockDone(2) + "\n\n" + tx.BlockIntro(conditionCatalog["B400"], 3, 6, KeyMapping{NormalKey: "f", BackwardKey: "j"}), `End of block 2. You may rest for a moment before continuing.

Block 3 of 6

In this block, you will see the outline of the character, then the arrow, then the character.

Press F for normal and J for mirrored.

Press the space bar to begin.`},
		"end": {tx.End(`C:\data\x.csv`), `The experiment is over. Thank you for taking part.

Your responses were saved in
C:\data\x.csv

This task is adapted from Cooper and Shepard (1973), "The time required to prepare for a rotated stimulus", Memory & Cognition, 1(3), 246-250.

Press any key to close the window.`},
		"end (browser)": {tx.EndBrowser, `The experiment is over. Thank you for taking part.

This task is adapted from Cooper and Shepard (1973), "The time required to prepare for a rotated stimulus", Memory & Cognition, 1(3), 246-250.

Press any key to finish.`},
	}
	for name, c := range cases {
		if got := plainText(c[0]); got != c[1] {
			t.Errorf("%s screen differs from the specification:\n--- got\n%s\n--- want\n%s", name, got, c[1])
		}
	}
	// The bold and boxed parts (see richtext.go).
	w := tx.Welcome(m)
	for _, part := range []string{boxed("J"), boxed("F"), bold("space bar")} {
		if !strings.Contains(w, part) {
			t.Errorf("welcome screen: %q is not marked", plainText(part))
		}
	}
	if tx.Error != "Error" || tx.TooSlow != "Too slow" {
		t.Errorf("feedback words: %q / %q", tx.Error, tx.TooSlow)
	}
}

// allScreens renders every participant-facing string of a language.
func allScreens(tx Texts) []string {
	m := MappingFor(2)
	all := []string{tx.Welcome(m), tx.BlockTypes, tx.Practice(6), tx.AfterPractice(6, 24),
		tx.BlockDone(1), tx.End("path"), tx.EndBrowser, tx.Error, tx.TooSlow, tx.DebriefPrompt, tx.DebriefHint}
	for _, name := range []string{"N", "B400", "B1000", "C"} {
		all = append(all, tx.BlockIntro(conditionCatalog[name], 1, 6, m))
	}
	for i := range all {
		all[i] = plainText(all[i])
	}
	return all
}

// No colons, "=" signs, em dashes, emoji or leftover placeholders, in either
// language.
func TestTextStyle(t *testing.T) {
	for _, lang := range []string{"en", "fr"} {
		tx, err := TextsFor(lang)
		if err != nil {
			t.Fatal(err)
		}
		for _, s := range allScreens(tx) {
			if s == "" || strings.ContainsAny(s, ":=—–{}%") {
				t.Errorf("%s: forbidden character in %q", lang, s)
			}
			for _, r := range s {
				if r > 0xFF && !unicode.IsLetter(r) {
					t.Errorf("%s: unexpected symbol %q in %q", lang, r, s)
				}
			}
		}
	}
	if _, err := TextsFor("de"); err == nil {
		t.Error("unknown language accepted")
	}
}
