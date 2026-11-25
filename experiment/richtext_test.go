// Copyright (2026) Thiago Lizier Rissi

package main

import (
	"reflect"
	"testing"
)

func TestPlainText(t *testing.T) {
	s := "Press " + boxed("J") + " or the " + bold("space bar") + "."
	if got := plainText(s); got != "Press J or the space bar." {
		t.Errorf("plainText = %q", got)
	}
}

func TestParseMarkup(t *testing.T) {
	got := parseMarkup("Press " + boxed("J") + ", then the " + bold("space bar") + ".\n\nEnd")
	want := [][]word{
		{
			{{text: "Press"}},
			{{text: "J", boxed: true}, {text: ","}},
			{{text: "then"}},
			{{text: "the"}},
			{{text: "space", bold: true}},
			{{text: "bar", bold: true}, {text: "."}},
		},
		nil, // the empty line between the paragraphs
		{{{text: "End"}}},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("parseMarkup:\n got %+v\nwant %+v", got, want)
	}
	// A boxed text is one word, even with a space in it.
	got = parseMarkup(boxed("Page Up"))
	want = [][]word{{{{text: "Page Up", boxed: true}}}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("boxed with a space:\n got %+v\nwant %+v", got, want)
	}
}

func TestWrapLines(t *testing.T) {
	// Every character is 1 wide, a space too.
	width := func(w word) float32 {
		n := 0
		for _, s := range w {
			n += len([]rune(s.text))
		}
		return float32(n)
	}
	lines, ok := wrapLines(parseMarkup("aaa bbb ccc\n\ndd"), width, 1, 7)
	if !ok {
		t.Fatal("wrapLines: every word fits, but ok is false")
	}
	var got []string
	for _, l := range lines {
		s := ""
		for i, w := range l {
			if i > 0 {
				s += " "
			}
			for _, seg := range w {
				s += seg.text
			}
		}
		got = append(got, s)
	}
	want := []string{"aaa bbb", "ccc", "", "dd"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("wrapLines = %q, want %q", got, want)
	}
	if _, ok := wrapLines(parseMarkup("aaaaaaaa"), width, 1, 7); ok {
		t.Error("a word wider than the line was accepted")
	}
}
