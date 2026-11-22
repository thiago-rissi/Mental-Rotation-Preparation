// Copyright (2026) Thiago Lizier Rissi

package main

// Stimulus rendering: glyph textures (filled test character and outline cue),
// the circular aperture and the orientation arrow.
//
// No goxpyriment stimulus takes a rotation, so glyphs are pre-rendered once to
// textures and drawn with SDL_RenderTextureRotated, which both mirrors and
// rotates. Two details matter:
//
//   - The texture is cropped to the glyph's INK bounding box (the src rect),
//     and the glyph is rotated about the centre of that box. A TTF surface
//     carries the font's ascent/descent padding, so its own centre is not the
//     visual centre of the character, and rotating about it would move the
//     character around inside the aperture from one orientation to the next.
//   - SDL applies the flip to the texture in its own frame and then rotates
//     the destination quad, so FLIP_HORIZONTAL with angle θ shows
//     rotate(θ)·mirror(glyph): mirror first, then rotate, as the paradigm
//     requires. -sheet shows this for every character and orientation.

import (
	"fmt"
	"image"
	"math"

	"github.com/Zyko0/go-sdl3/sdl"
	"github.com/Zyko0/go-sdl3/ttf"
	"github.com/chrplr/goxpyriment/apparatus"
	"github.com/chrplr/goxpyriment/control"
	"github.com/chrplr/goxpyriment/stimuli"
)

// Sizes, as fractions of the screen. The paper's character subtended ~1.5° and
// its aperture ~4°; viewing distance is not controlled here, so the display is
// scaled to the screen instead.
const (
	GlyphHeightFrac = 0.08 // ink height of the tallest character / screen height
	ApertureRatio   = 2.7  // aperture diameter / glyph height
	ArrowRatio      = 0.80 // arrow half-length / aperture radius
	RingWidthRatio  = 0.025
	ShaftWidthRatio = 0.06    // arrow shaft width / glyph height
	HeadLenRatio    = 0.28    // arrowhead length / arrow half-length
	HeadHalfWRatio  = 0.16    // arrowhead half-width / arrow half-length
	OutlineRatio    = 1. / 28 // outline band half-width / glyph height

	inkThreshold = 8 // alpha above which a pixel counts as ink
)

// Palette (light theme). Every colour the program uses is defined here. The
// background and foreground go to NewExperimentFromFlags; the framework then
// clears with exp.BackgroundColor and all text is drawn in exp.ForegroundColor.
var (
	bgColor       = control.RGB(0xf4, 0xf5, 0xf7)
	fgColor       = control.RGB(0x1c, 0x1c, 0x2e)
	glyphColor    = control.RGB(0x1c, 0x1c, 0x2e) // test character and outline cue
	ringColor     = control.RGB(0xc4, 0xc7, 0xd0) // aperture
	arrowColor    = control.RGB(0x3b, 0x4e, 0xc8)
	feedbackColor = control.RGB(0xb3, 0x26, 0x1e) // "Error" / "Too slow"
	// Used for the cue only if the outline turns out not to be hollow: a
	// filled glyph halfway between the background and the character.
	fallbackCueColor = control.RGB(0x88, 0x89, 0x93)
)

// ─── Pure helpers (unit-tested) ────────────────────────────────────────────

// alphaPlane extracts the alpha channel of RGBA32 pixels (byte order R,G,B,A).
func alphaPlane(pixels []byte, w, h, pitch int) []uint8 {
	a := make([]uint8, w*h)
	for y := 0; y < h; y++ {
		row := pixels[y*pitch:]
		for x := 0; x < w; x++ {
			a[y*w+x] = row[4*x+3]
		}
	}
	return a
}

// inkBBox returns the smallest rectangle holding every pixel whose alpha is
// above threshold, and false if there is none.
func inkBBox(alpha []uint8, w, h int, threshold uint8) (image.Rectangle, bool) {
	x0, y0, x1, y1 := w, h, -1, -1
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			if alpha[y*w+x] > threshold {
				x0, y0 = min(x0, x), min(y0, y)
				x1, y1 = max(x1, x), max(y1, y)
			}
		}
	}
	if x1 < 0 {
		return image.Rectangle{}, false
	}
	return image.Rect(x0, y0, x1+1, y1+1), true
}

// mask is an alpha plane with its ink box.
type mask struct {
	a    []uint8
	w, h int
	box  image.Rectangle
}

func (m mask) at(x, y int) uint8 {
	if x < 0 || y < 0 || x >= m.w || y >= m.h {
		return 0
	}
	return m.a[y*m.w+x]
}

func (m mask) centre() (float64, float64) {
	return float64(m.box.Min.X+m.box.Max.X) / 2, float64(m.box.Min.Y+m.box.Max.Y) / 2
}

// hollowScore measures how much of the deep interior of a filled glyph is also
// covered by its outline version, after aligning the two on their ink centres.
// The deep interior is the set of filled pixels whose whole (2k+1)² square
// neighbourhood is filled, i.e. pixels more than k px from any edge. A hollow
// outline whose stroke reaches less than k px inside the contour covers none
// of it (score ≈ 0); a filled one covers all of it (score ≈ 1). n is the number
// of interior pixels the score is based on.
func hollowScore(fill, outline mask, k int) (score float64, n int) {
	fcx, fcy := fill.centre()
	ocx, ocy := outline.centre()
	dx, dy := int(math.Round(ocx-fcx)), int(math.Round(ocy-fcy))
	covered := 0
	for y := fill.box.Min.Y; y < fill.box.Max.Y; y++ {
		for x := fill.box.Min.X; x < fill.box.Max.X; x++ {
			if !deepInside(fill, x, y, k) {
				continue
			}
			n++
			if outline.at(x+dx, y+dy) > 128 {
				covered++
			}
		}
	}
	if n == 0 {
		return 0, 0
	}
	return float64(covered) / float64(n), n
}

func deepInside(m mask, x, y, k int) bool {
	for yy := y - k; yy <= y+k; yy++ {
		for xx := x - k; xx <= x+k; xx++ {
			if m.at(xx, yy) <= 128 {
				return false
			}
		}
	}
	return true
}

// circlePoints returns n points on a circle of radius r (y up).
func circlePoints(r float32, n int) []sdl.FPoint {
	pts := make([]sdl.FPoint, n)
	for i := range n {
		a := 2 * math.Pi * float64(i) / float64(n)
		pts[i] = sdl.FPoint{X: r * float32(math.Cos(a)), Y: r * float32(math.Sin(a))}
	}
	return pts
}

// arrowGeometry returns the shaft end points and the arrowhead triangle of an
// arrow through the origin pointing θ degrees clockwise from up, in
// centre-origin, y-up coordinates: the tip is at (L·sin θ, L·cos θ).
func arrowGeometry(theta float64, halfLen, headLen, headHalfW float32) (tail, base sdl.FPoint, head []sdl.FPoint) {
	rad := theta * math.Pi / 180
	ux, uy := float32(math.Sin(rad)), float32(math.Cos(rad)) // unit vector toward the tip
	px, py := uy, -ux                                        // perpendicular
	tip := sdl.FPoint{X: halfLen * ux, Y: halfLen * uy}
	tail = sdl.FPoint{X: -tip.X, Y: -tip.Y}
	base = sdl.FPoint{X: tip.X - headLen*ux, Y: tip.Y - headLen*uy}
	head = []sdl.FPoint{
		tip,
		{X: base.X + headHalfW*px, Y: base.Y + headHalfW*py},
		{X: base.X - headHalfW*px, Y: base.Y - headHalfW*py},
	}
	return tail, base, head
}

// ─── SDL side ──────────────────────────────────────────────────────────────

// glyph is a texture holding one character, with the ink box as src rect.
type glyph struct {
	tex *sdl.Texture
	src sdl.FRect
}

// draw draws the glyph with its ink centre at (cx, cy) (centre-origin, y up),
// scaled by scale, rotated angle degrees clockwise, mirrored first if backward.
func (g *glyph) draw(s *apparatus.Screen, cx, cy, scale float32, angle float64, backward bool) error {
	sx, sy := s.CenterToSDL(cx, cy)
	w, h := g.src.W*scale, g.src.H*scale
	dst := sdl.FRect{X: sx - w/2, Y: sy - h/2, W: w, H: h}
	flip := sdl.FLIP_NONE
	if backward {
		flip = sdl.FLIP_HORIZONTAL
	}
	// center = nil: rotate about the centre of dst, i.e. the ink centre.
	if err := s.Renderer.RenderTextureRotated(g.tex, &g.src, &dst, angle, nil, flip); err != nil {
		return fmt.Errorf("drawing glyph: %w", err)
	}
	return nil
}

// rasterize renders one character with SDL_ttf and returns its alpha mask.
// Only the alpha is kept: textures are built from it with the colour wanted
// (glyphTexture), and the outline cue is derived from it (outlineMask).
func rasterize(font *ttf.Font, ch rune) (mask, error) {
	raw, err := font.RenderGlyphBlended(ch, control.White)
	if err != nil {
		return mask{}, fmt.Errorf("rendering %q: %w", ch, err)
	}
	surf, err := raw.Convert(sdl.PIXELFORMAT_RGBA32)
	raw.Destroy()
	if err != nil {
		return mask{}, fmt.Errorf("converting %q: %w", ch, err)
	}
	defer surf.Destroy()
	w, h := int(surf.W), int(surf.H)
	m := mask{a: alphaPlane(surf.Pixels(), w, h, int(surf.Pitch)), w: w, h: h}
	box, ok := inkBBox(m.a, w, h, inkThreshold)
	if !ok {
		return mask{}, fmt.Errorf("character %q renders no ink", ch)
	}
	m.box = box
	return m, nil
}

// diskOffsets lists the integer offsets within distance r of the origin.
func diskOffsets(r int) []image.Point {
	var pts []image.Point
	for dy := -r; dy <= r; dy++ {
		for dx := -r; dx <= r; dx++ {
			if dx*dx+dy*dy <= r*r {
				pts = append(pts, image.Pt(dx, dy))
			}
		}
	}
	return pts
}

// outlineMask returns the hollow outline of a glyph: the band of pixels within
// r px of its edge, inside or outside, computed as the grey-level
// morphological gradient (dilation minus erosion by a disk of radius r) of its
// alpha. The band is 2r wide and keeps the antialiasing of the source. The
// result is padded by r px on each side, so its ink centre is the glyph's.
//
// This replaces SDL_ttf's own outline mode (TTF_SetFontOutline), whose binding
// is not implemented in the browser build: computing the outline here gives
// the same cue on the desktop and in the browser.
func outlineMask(m mask, r int) mask {
	disk := diskOffsets(r)
	w, h := m.w+2*r, m.h+2*r
	out := mask{a: make([]uint8, w*h), w: w, h: h}
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			sx, sy := x-r, y-r // source coordinates
			lo, hi := uint8(255), uint8(0)
			for _, d := range disk {
				v := m.at(sx+d.X, sy+d.Y)
				lo, hi = min(lo, v), max(hi, v)
			}
			out.a[y*w+x] = hi - lo
		}
	}
	out.box, _ = inkBBox(out.a, w, h, inkThreshold)
	return out
}

// glyphTexture makes a texture of the given colour whose alpha is the mask,
// with the mask's ink box as source rectangle.
func glyphTexture(s *apparatus.Screen, m mask, color sdl.Color) (*glyph, error) {
	px := make([]byte, 4*m.w*m.h)
	for i, a := range m.a {
		px[4*i], px[4*i+1], px[4*i+2] = color.R, color.G, color.B
		px[4*i+3] = uint8(uint16(a) * uint16(color.A) / 255)
	}
	tex, err := s.TextureFromRGBA(m.w, m.h, px, 4*m.w)
	if err != nil {
		return nil, err
	}
	if err := tex.SetScaleMode(sdl.SCALEMODE_LINEAR); err != nil {
		tex.Destroy()
		return nil, fmt.Errorf("scale mode: %w", err)
	}
	src := sdl.FRect{X: float32(m.box.Min.X), Y: float32(m.box.Min.Y),
		W: float32(m.box.Dx()), H: float32(m.box.Dy())}
	return &glyph{tex: tex, src: src}, nil
}

// Stimuli holds every pre-rendered stimulus.
type Stimuli struct {
	GlyphH    float32 // ink height of the tallest character, px
	ApR       float32 // aperture radius, px
	FontPt    float32
	OutlinePx int // half-width of the outline band
	// CueHollow is false when the outline did not come out hollow (strokes too
	// thin for the band), in which case the cue is a grey filled glyph.
	CueHollow   bool
	HollowScore float64 // worst (highest) score over the characters
	HollowN     int     // fewest interior pixels the score was based on

	fill, cue map[string]*glyph
	rings     map[float32]*stimuli.PolyLine
	arrows    map[arrowKey]*arrowStim
}

type arrowKey struct {
	scale  float32
	orient int
}

type arrowStim struct {
	shaft *stimuli.PolyLine
	head  *stimuli.Shape
}

// probePt is the size at which glyphs are measured and the outline is checked.
const probePt = 200

// NewStimuli renders the characters for a screen of logical height screenH.
func NewStimuli(s *apparatus.Screen, fontData []byte, chars []string, screenH float32) (*Stimuli, error) {
	font, err := control.FontFromMemory(fontData, probePt)
	if err != nil {
		return nil, fmt.Errorf("opening font: %w", err)
	}
	defer font.Close()

	runes := make([]rune, len(chars))
	for i, c := range chars {
		runes[i] = []rune(c)[0]
	}

	// 1. Measure at the probe size, and check there (where strokes are thick
	// enough for the interior test to mean something) that the outline is
	// hollow at the same stroke-to-glyph ratio as on screen.
	probe := map[rune]mask{}
	maxInk := 0
	for _, ch := range runes {
		m, err := rasterize(font, ch)
		if err != nil {
			return nil, err
		}
		probe[ch] = m
		maxInk = max(maxInk, m.box.Dy())
	}
	target := GlyphHeightFrac * screenH
	pt := float32(probePt) * target / float32(maxInk)
	outlinePx := max(1, int(math.Round(float64(target*OutlineRatio))))
	probeOutline := max(2, int(math.Round(float64(outlinePx)*probePt/float64(pt))))

	st := &Stimuli{
		FontPt:    pt,
		OutlinePx: outlinePx,
		ApR:       ApertureRatio * target / 2,
		fill:      map[string]*glyph{},
		cue:       map[string]*glyph{},
		rings:     map[float32]*stimuli.PolyLine{},
		arrows:    map[arrowKey]*arrowStim{},
		HollowN:   math.MaxInt,
	}
	for _, ch := range runes {
		score, n := hollowScore(probe[ch], outlineMask(probe[ch], probeOutline), probeOutline+1)
		st.HollowScore = max(st.HollowScore, score)
		st.HollowN = min(st.HollowN, n)
	}
	st.CueHollow = st.HollowN > 0 && st.HollowScore < 0.2

	// 2. Render at the display size.
	if err := font.SetSize(pt); err != nil {
		return nil, fmt.Errorf("setting font size %.1f: %w", pt, err)
	}
	for i, ch := range runes {
		m, err := rasterize(font, ch)
		if err != nil {
			return nil, err
		}
		if st.fill[chars[i]], err = glyphTexture(s, m, glyphColor); err != nil {
			return nil, fmt.Errorf("texture for %q: %w", ch, err)
		}
		st.GlyphH = max(st.GlyphH, float32(m.box.Dy()))
		cue, color := outlineMask(m, outlinePx), glyphColor
		if !st.CueHollow {
			cue, color = m, fallbackCueColor
		}
		if st.cue[chars[i]], err = glyphTexture(s, cue, color); err != nil {
			return nil, fmt.Errorf("cue texture for %q: %w", ch, err)
		}
	}
	return st, nil
}

// Destroy frees the textures.
func (st *Stimuli) Destroy() {
	for _, m := range []map[string]*glyph{st.fill, st.cue} {
		for _, g := range m {
			g.tex.Destroy()
		}
	}
}

func (st *Stimuli) ring(scale float32) *stimuli.PolyLine {
	if p, ok := st.rings[scale]; ok {
		return p
	}
	w := max(2, float32(math.Round(float64(st.GlyphH*RingWidthRatio*scale))))
	p := stimuli.NewPolyLine(circlePoints(st.ApR*scale, 180), true, w, ringColor)
	st.rings[scale] = p
	return p
}

func (st *Stimuli) arrow(scale float32, orient int) *arrowStim {
	k := arrowKey{scale, orient}
	if a, ok := st.arrows[k]; ok {
		return a
	}
	half := ArrowRatio * st.ApR * scale
	tail, base, head := arrowGeometry(float64(orient), half, HeadLenRatio*half, HeadHalfWRatio*half)
	w := max(3, st.GlyphH*ShaftWidthRatio*scale)
	a := &arrowStim{
		shaft: stimuli.NewPolyLine([]sdl.FPoint{tail, base}, false, w, arrowColor),
		head:  stimuli.NewShape(head, arrowColor),
	}
	st.arrows[k] = a
	return a
}

// DrawAperture draws the ring centred on (cx, cy).
func (st *Stimuli) DrawAperture(s *apparatus.Screen, cx, cy, scale float32) error {
	p := st.ring(scale)
	p.SetPosition(sdl.FPoint{X: cx, Y: cy})
	return p.Draw(s)
}

// DrawArrow draws the orientation arrow pointing orient degrees clockwise from up.
func (st *Stimuli) DrawArrow(s *apparatus.Screen, cx, cy, scale float32, orient int) error {
	a := st.arrow(scale, orient)
	pos := sdl.FPoint{X: cx, Y: cy}
	a.shaft.SetPosition(pos)
	a.head.SetPosition(pos)
	if err := a.shaft.Draw(s); err != nil {
		return err
	}
	return a.head.Draw(s)
}

// DrawTest draws the filled test character.
func (st *Stimuli) DrawTest(s *apparatus.Screen, cx, cy, scale float32, ch string, orient int, backward bool) error {
	g, ok := st.fill[ch]
	if !ok {
		return fmt.Errorf("no glyph for %q", ch)
	}
	return g.draw(s, cx, cy, scale, float64(orient), backward)
}

// DrawCue draws the outline of the NORMAL character (identity cue when orient
// is 0, combined cue of condition C otherwise).
func (st *Stimuli) DrawCue(s *apparatus.Screen, cx, cy, scale float32, ch string, orient int) error {
	g, ok := st.cue[ch]
	if !ok {
		return fmt.Errorf("no cue glyph for %q", ch)
	}
	return g.draw(s, cx, cy, scale, float64(orient), false)
}

// ─── Full-size trial scenes ───────────────────────────────────────────────

func (st *Stimuli) ApertureScene() stimuli.DrawFunc {
	return func(s *apparatus.Screen) error { return st.DrawAperture(s, 0, 0, 1) }
}

func (st *Stimuli) TestScene(t Trial) stimuli.DrawFunc {
	return func(s *apparatus.Screen) error {
		if err := st.DrawAperture(s, 0, 0, 1); err != nil {
			return err
		}
		return st.DrawTest(s, 0, 0, 1, t.Character, t.Orientation, t.Version == Backward)
	}
}

func (st *Stimuli) CueScene(ch string, orient int) stimuli.DrawFunc {
	return func(s *apparatus.Screen) error {
		if err := st.DrawAperture(s, 0, 0, 1); err != nil {
			return err
		}
		return st.DrawCue(s, 0, 0, 1, ch, orient)
	}
}

func (st *Stimuli) ArrowScene(orient int) stimuli.DrawFunc {
	return func(s *apparatus.Screen) error {
		if err := st.DrawAperture(s, 0, 0, 1); err != nil {
			return err
		}
		return st.DrawArrow(s, 0, 0, 1, orient)
	}
}

// MessageScene draws a text line inside the aperture.
func (st *Stimuli) MessageScene(msg *stimuli.TextLine) stimuli.DrawFunc {
	return func(s *apparatus.Screen) error {
		if err := st.DrawAperture(s, 0, 0, 1); err != nil {
			return err
		}
		return msg.Draw(s)
	}
}
