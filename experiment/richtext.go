// Copyright (2026) Thiago Lizier Rissi

package main

// Instruction screens with bold words and boxed keys.
//
// In strings.go, bold("space bar") draws "space bar" in bold, and boxed("J")
// draws J inside a box, like a key of the keyboard. goxpyriment's own text box
// renders a whole screen as one plain block, so these screens are laid out
// here instead: words are wrapped to the screen width, each line is centred,
// "\n" starts a new line and "\n\n" leaves an empty line, as before.
//
// Bold is drawn by printing the word twice, one or two pixels apart: SDL_ttf's
// bold style is not available in the browser build, and this looks the same
// on the desktop and in the browser.

import (
	"fmt"
	"math"
	"strings"

	"github.com/Zyko0/go-sdl3/sdl"
	"github.com/Zyko0/go-sdl3/ttf"
	"github.com/chrplr/goxpyriment/apparatus"
	"github.com/chrplr/goxpyriment/control"
)

// Markers placed in the text by bold and boxed. They are control characters,
// which never occur in the texts themselves.
const (
	boldOn   = '\x01'
	boldOff  = '\x02'
	boxedOn  = '\x03'
	boxedOff = '\x04'
)

// bold marks s to be drawn in bold.
func bold(s string) string { return string(boldOn) + s + string(boldOff) }

// boxed marks s to be drawn inside a box, like a key.
func boxed(s string) string { return string(boxedOn) + s + string(boxedOff) }

// plainText removes the bold and boxed marks.
func plainText(s string) string {
	return strings.Map(func(r rune) rune {
		switch r {
		case boldOn, boldOff, boxedOn, boxedOff:
			return -1
		}
		return r
	}, s)
}

// ─── Layout (pure, unit-tested) ───────────────────────────────────────────

// segment is a piece of a word with a single style.
type segment struct {
	text        string
	bold, boxed bool
}

// word is what wrapping never splits: the characters between two spaces,
// possibly in several styles ("J," with J boxed is one word). A boxed text is
// never split, even if it contains spaces.
type word []segment

// parseMarkup splits marked-up text into source lines (one per "\n"), each a
// list of words. Runs of spaces count as one, and an empty line has no words.
func parseMarkup(text string) [][]word {
	var lines [][]word
	var line []word
	var cur word
	var buf strings.Builder
	isBold, isBoxed := false, false

	flushSeg := func() {
		if buf.Len() > 0 {
			cur = append(cur, segment{text: buf.String(), bold: isBold, boxed: isBoxed})
			buf.Reset()
		}
	}
	flushWord := func() {
		flushSeg()
		if len(cur) > 0 {
			line = append(line, cur)
			cur = nil
		}
	}
	for _, r := range text {
		switch {
		case r == boldOn || r == boldOff:
			flushSeg()
			isBold = r == boldOn
		case r == boxedOn || r == boxedOff:
			flushSeg()
			isBoxed = r == boxedOn
		case r == '\n':
			flushWord()
			lines = append(lines, line)
			line = nil
		case r == ' ' && !isBoxed:
			flushWord()
		default:
			buf.WriteRune(r)
		}
	}
	flushWord()
	return append(lines, line)
}

// wrapLines breaks each source line into screen lines no wider than maxW,
// given the width of each word and of a space. ok is false when a single word
// is wider than maxW (the text then needs a smaller font).
func wrapLines(src [][]word, width func(word) float32, space, maxW float32) (out [][]word, ok bool) {
	ok = true
	for _, line := range src {
		var cur []word
		var curW float32
		for _, w := range line {
			ww := width(w)
			if ww > maxW {
				ok = false
			}
			if len(cur) > 0 && curW+space+ww > maxW {
				out = append(out, cur)
				cur, curW = nil, 0
			}
			if len(cur) > 0 {
				curW += space
			}
			cur = append(cur, w)
			curW += ww
		}
		out = append(out, cur) // an empty source line stays an empty line
	}
	return out, ok
}

// ─── Rendering ─────────────────────────────────────────────────────────────

// Proportions, relative to the height of the font.
const (
	lineSpacing    = 1.2  // distance between lines
	boxPadX        = 0.35 // space between a boxed text and the sides of its box
	boxPadY        = 0.05 // space above and below it
	boxLineRatio   = 1. / 16
	boldShiftRatio = 1. / 22
	richTextWidth  = 0.92 // of the drawing area, as goxpyriment's text boxes
	richTextHeight = 0.92
	minTextPt      = 12
	// The font may shrink down to this fraction of the default size so that
	// each line of the text stays on one screen line (as goxpyriment's own
	// text boxes do).
	keepBreaksFloor = 0.75
)

// richItem is one rendered segment, at a position in centre-origin, y-up
// coordinates (the centre of its text).
type richItem struct {
	tex          *sdl.Texture
	cx, cy, w, h float32
	bold, boxed  bool
}

// richText is a screen of marked-up text, ready to draw.
type richText struct {
	items      []richItem
	padX, padY float32
	boxLine    float32
	boldShift  float32
	color      sdl.Color
	// OffsetY moves the whole text up (positive) or down (negative) from the
	// centre of the screen, in pixels.
	OffsetY float32
}

// newRichText lays out and renders text with the experiment's font, at the
// default text size or smaller if that is needed to fit the screen.
func newRichText(exp *control.Experiment, text string) (*richText, error) {
	areaW, areaH := exp.DrawArea()
	maxW, maxH := richTextWidth*areaW, richTextHeight*areaH
	src := parseMarkup(text)

	var font *ttf.Font
	var lines [][]word
	var h, space float32
	var width func(word) float32
	// try lays the text out at one size. It keeps the font open when the
	// layout is accepted by accept, and closes it otherwise.
	try := func(size float32, accept func(fits, unwrapped bool) bool) (bool, error) {
		f, err := control.FontFromMemory(fontData, size)
		if err != nil {
			return false, fmt.Errorf("text font at %.0f pt: %w", size, err)
		}
		fh := float32(f.Height())
		sw, _, _ := f.StringSize(" ")
		fw := func(w word) float32 {
			var total float32
			for _, s := range w {
				tw, _, _ := f.StringSize(s.text)
				total += float32(tw)
				if s.bold {
					total += max(1, float32(math.Round(float64(fh*boldShiftRatio))))
				}
				if s.boxed {
					total += 2 * boxPadX * fh
				}
			}
			return total
		}
		ls, ok := wrapLines(src, fw, float32(sw), maxW)
		fits := ok && float32(len(ls))*lineSpacing*fh <= maxH
		if !accept(fits, len(ls) == len(src)) {
			f.Close()
			return false, nil
		}
		font, lines, h, space, width = f, ls, fh, float32(sw), fw
		return true, nil
	}
	// First choice: the largest size, down to keepBreaksFloor of the default,
	// at which every line written in the text fits on one screen line, so
	// that a sentence is never split over two lines. Otherwise: the largest
	// size at which the text fits the screen, wrapping the long lines.
	largest := float32(math.Round(float64(exp.DefaultFontSize)))
	floor := max(minTextPt, float32(math.Ceil(float64(keepBreaksFloor*largest))))
	done := false
	for size := largest; size >= floor && !done; size-- {
		var err error
		if done, err = try(size, func(fits, unwrapped bool) bool { return fits && unwrapped }); err != nil {
			return nil, err
		}
	}
	for size := largest; !done; size-- {
		var err error
		if done, err = try(size, func(fits, _ bool) bool { return fits || size <= minTextPt }); err != nil {
			return nil, err
		}
	}
	defer font.Close()

	rt := &richText{
		padX:      boxPadX * h,
		padY:      boxPadY * h,
		boxLine:   max(1, float32(math.Round(float64(h*boxLineRatio)))),
		boldShift: max(1, float32(math.Round(float64(h*boldShiftRatio)))),
		color:     exp.ForegroundColor,
	}
	lineH := lineSpacing * h
	top := float32(len(lines)) * lineH / 2
	renderer := exp.Screen.Renderer
	for i, line := range lines {
		cy := top - (float32(i)+0.5)*lineH
		var lineW float32
		for j, w := range line {
			if j > 0 {
				lineW += space
			}
			lineW += width(w)
		}
		x := -lineW / 2
		for j, w := range line {
			if j > 0 {
				x += space
			}
			for _, s := range w {
				surf, err := font.RenderTextBlended(s.text, exp.ForegroundColor)
				if err != nil {
					rt.Unload()
					return nil, fmt.Errorf("rendering %q: %w", s.text, err)
				}
				tex, err := renderer.CreateTextureFromSurface(surf)
				sw, sh := float32(surf.W), float32(surf.H)
				surf.Destroy()
				if err != nil {
					rt.Unload()
					return nil, fmt.Errorf("texture for %q: %w", s.text, err)
				}
				segW := sw
				if s.bold {
					segW += rt.boldShift
				}
				if s.boxed {
					x += rt.padX
				}
				rt.items = append(rt.items, richItem{tex: tex, cx: x + sw/2, cy: cy, w: sw, h: sh, bold: s.bold, boxed: s.boxed})
				x += segW
				if s.boxed {
					x += rt.padX
				}
			}
		}
	}
	return rt, nil
}

// Draw draws the text; it is a stimuli.DrawFunc.
func (rt *richText) Draw(s *apparatus.Screen) error {
	r := s.Renderer
	for _, it := range rt.items {
		sx, sy := s.CenterToSDL(it.cx, it.cy+rt.OffsetY)
		dst := sdl.FRect{X: sx - it.w/2, Y: sy - it.h/2, W: it.w, H: it.h}
		if it.boxed {
			box := sdl.FRect{X: dst.X - rt.padX, Y: dst.Y - rt.padY, W: dst.W + 2*rt.padX, H: dst.H + 2*rt.padY}
			c := keyBoxColor
			if err := r.SetDrawColor(c.R, c.G, c.B, c.A); err != nil {
				return err
			}
			if err := r.RenderFillRect(&box); err != nil {
				return err
			}
			c = rt.color
			if err := r.SetDrawColor(c.R, c.G, c.B, c.A); err != nil {
				return err
			}
			for k := float32(0); k < rt.boxLine; k++ {
				edge := sdl.FRect{X: box.X + k, Y: box.Y + k, W: box.W - 2*k, H: box.H - 2*k}
				if err := r.RenderRect(&edge); err != nil {
					return err
				}
			}
		}
		if err := r.RenderTexture(it.tex, nil, &dst); err != nil {
			return err
		}
		if it.bold {
			dst.X += rt.boldShift
			if err := r.RenderTexture(it.tex, nil, &dst); err != nil {
				return err
			}
		}
	}
	return nil
}

// Unload frees the textures.
func (rt *richText) Unload() {
	for _, it := range rt.items {
		it.tex.Destroy()
	}
	rt.items = nil
}
