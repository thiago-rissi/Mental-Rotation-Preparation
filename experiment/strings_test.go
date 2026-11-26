// Copyright (2026) Thiago Lizier Rissi

package main

import (
	"regexp"
	"strings"
	"testing"
	"unicode"
	"unicode/utf8"
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
Respond as quickly as you can, but try not to make mistakes.

The session takes about ten minutes.

Press the space bar to continue.`},
		"block types": {tx.BlockTypes, `The experiment is divided into blocks of two kinds.

In some blocks, the circle stays empty for a moment before the character appears.

In the other blocks, you first see the outline of the coming character.
An arrow then shows where the top of the character will point, and the character appears as soon as the arrow is gone.

Press the space bar to continue.`},
		"practice": {tx.Practice(6), `You will begin with six practice trials.
Place your left index finger on F and your right index finger on J.

Press the space bar when you are ready.`},
		"after practice": {tx.AfterPractice(6, 24), `The practice is over.
The experiment itself has six blocks of 24 trials, and you can rest between blocks.

Press the space bar to begin.`},
		"block N": {tx.BlockIntro(conditionCatalog["N"], 1, 6, m), `Block 1 of 6

In this block, the circle stays empty for a moment before the character appears.

Press J for normal and F for mirrored.

Press the space bar to begin.`},
		"block B after a break": {tx.BlockDone(2) + "\n\n" + tx.BlockIntro(conditionCatalog["B400"], 3, 6, KeyMapping{NormalKey: "f", BackwardKey: "j"}), `End of block 2.
You may rest for a moment before continuing.

Block 3 of 6

In this block, you will see
outline -> arrow -> character.

Press F for normal and J for mirrored.

Press the space bar to begin.`},
		"end": {tx.End(`C:\data\x.csv`), `The experiment is over.
Thank you for taking part.

Your responses were saved in
C:\data\x.csv

This task is adapted from Cooper and Shepard (1973), "The time required to prepare for a rotated stimulus", Memory & Cognition, 1(3), 246-250.

Press any key to close the window.`},
		"end (browser)": {tx.EndBrowser, `The experiment is over.
Thank you for taking part.

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

// TestOneSentencePerLine checks that no line of any screen, in either
// language, holds more than one sentence: a sentence that ends (. ? or !)
// must be the last thing on its line.
func TestOneSentencePerLine(t *testing.T) {
	end := regexp.MustCompile(`[.?!]\s+\S`)
	for _, lang := range []string{"en", "fr"} {
		tx, err := TextsFor(lang)
		if err != nil {
			t.Fatal(err)
		}
		for _, screen := range allScreens(tx) {
			for _, line := range strings.Split(screen, "\n") {
				if end.MatchString(line) {
					t.Errorf("%s: more than one sentence on the line %q", lang, line)
				}
			}
		}
	}
}

// TestKeysHighlighted checks two rules on every screen, in both languages:
// the response keys F and J are always boxed, and the key that follows
// "press" ("appuyez sur") is always highlighted (bold or boxed).
func TestKeysHighlighted(t *testing.T) {
	for _, lang := range []string{"en", "fr"} {
		tx, err := TextsFor(lang)
		if err != nil {
			t.Fatal(err)
		}
		for _, s := range markedScreens(tx) {
			for _, problem := range keyProblems(s) {
				t.Errorf("%s: %s in %q", lang, problem, plainText(s))
			}
		}
	}
}

// TestKeyProblemsDetects makes sure keyProblems does find what it looks for.
func TestKeyProblemsDetects(t *testing.T) {
	for text, want := range map[string]int{
		"Press " + boxed("J") + " for normal and " + boxed("F") + " for mirrored.":            0,
		"Press the " + bold("space bar") + ". Appuyez sur la " + bold("barre d'espace") + ".": 0,
		"Place your finger on F.":           1,
		"Press J for normal.":               2,
		"Press the space bar to continue.":  1,
		"Appuyez sur une touche.":           1,
		"Jump and Fly are words, not keys.": 0,
	} {
		if got := keyProblems(text); len(got) != want {
			t.Errorf("keyProblems(%q) = %q, want %d problem(s)", plainText(text), got, want)
		}
	}
}

// keyProblems lists the F and J that are not boxed, and the keys after
// "press" or "appuyez sur" that are neither bold nor boxed.
func keyProblems(s string) []string {
	var problems []string
	inBox := false
	var word []rune
	endWord := func() {
		if w := string(word); !inBox && (w == "F" || w == "J") {
			problems = append(problems, "key "+w+" is not boxed")
		}
		word = word[:0]
	}
	for _, r := range s + " " {
		switch {
		case r == boxedOn:
			endWord()
			inBox = true
		case r == boxedOff:
			word = word[:0]
			inBox = false
		case unicode.IsLetter(r):
			word = append(word, r)
		default:
			endWord()
		}
	}
	lower := strings.ToLower(s)
	for _, verb := range []string{"press ", "appuyez sur "} {
		for from := 0; ; {
			i := strings.Index(lower[from:], verb)
			if i < 0 {
				break
			}
			i += from
			rest := s[i+len(verb):]
			for _, article := range []string{"the ", "la ", "le ", "l'"} {
				rest = strings.TrimPrefix(rest, article)
			}
			if r, _ := utf8.DecodeRuneInString(rest); r != boldOn && r != boxedOn {
				problems = append(problems, "key after \""+strings.TrimSpace(verb)+"\" is not highlighted")
			}
			from = i + len(verb)
		}
	}
	return problems
}

// markedScreens renders every participant-facing string of a language, with
// the bold and boxed marks kept.
func markedScreens(tx Texts) []string {
	m := MappingFor(2)
	all := []string{tx.Welcome(m), tx.BlockTypes, tx.Practice(6), tx.AfterPractice(6, 24),
		tx.BlockDone(1), tx.End("path"), tx.EndBrowser, tx.Error, tx.TooSlow, tx.DebriefPrompt, tx.DebriefHint}
	for _, name := range []string{"N", "B400", "B1000", "C"} {
		all = append(all, tx.BlockIntro(conditionCatalog[name], 1, 6, m))
	}
	return all
}

// allScreens renders every participant-facing string of a language.
func allScreens(tx Texts) []string {
	all := markedScreens(tx)
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
