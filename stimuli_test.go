// Copyright (2026) Thiago Lizier Rissi

package main

import (
	"image"
	"math"
	"testing"
)

// disc returns a w x h alpha plane with a filled disc (hollow: a ring of the
// given stroke) centred at (cx, cy).
func disc(w, h int, cx, cy, r, stroke float64, hollow bool) mask {
	a := make([]uint8, w*h)
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			d := math.Hypot(float64(x)+0.5-cx, float64(y)+0.5-cy)
			in := d <= r
			if hollow {
				in = math.Abs(d-r) <= stroke
			}
			if in {
				a[y*w+x] = 255
			}
		}
	}
	box, _ := inkBBox(a, w, h, inkThreshold)
	return mask{a: a, w: w, h: h, box: box}
}

func TestInkBBox(t *testing.T) {
	a := make([]uint8, 10*8)
	a[2*10+3] = 200 // (3,2)
	a[5*10+7] = 9   // (7,5), just above the threshold
	a[6*10+1] = 8   // (1,6), not ink
	box, ok := inkBBox(a, 10, 8, inkThreshold)
	if !ok || box != image.Rect(3, 2, 8, 6) {
		t.Fatalf("box %v ok %v, want (3,2)-(8,6)", box, ok)
	}
	if _, ok := inkBBox(make([]uint8, 4), 2, 2, inkThreshold); ok {
		t.Fatal("empty plane reported ink")
	}
}

func TestAlphaPlane(t *testing.T) {
	// 2x2 RGBA with a 12-byte pitch (4 bytes of row padding).
	px := []byte{
		1, 2, 3, 10, 1, 2, 3, 20, 0, 0, 0, 0,
		1, 2, 3, 30, 1, 2, 3, 40, 0, 0, 0, 0,
	}
	got := alphaPlane(px, 2, 2, 12)
	want := []uint8{10, 20, 30, 40}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("alpha %v, want %v", got, want)
		}
	}
}

func TestHollowScore(t *testing.T) {
	fill := disc(80, 90, 40, 50, 30, 0, false)
	// Outline surfaces are larger and offset; alignment is on ink centres.
	ring := disc(100, 100, 50, 52, 30, 3, true)
	solid := disc(100, 100, 50, 52, 33, 0, false)
	if s, n := hollowScore(fill, ring, 4); n == 0 || s > 0.01 {
		t.Errorf("hollow outline: score %.3f over %d px, want ~0", s, n)
	}
	if s, n := hollowScore(fill, solid, 4); n == 0 || s < 0.99 {
		t.Errorf("filled outline: score %.3f over %d px, want ~1", s, n)
	}
}

// outlineMask turns a filled disc into a hollow ring of width 2r centred on its
// edge, with the same ink centre.
func TestOutlineMask(t *testing.T) {
	const r = 3
	fill := disc(80, 80, 40, 40, 25, 0, false)
	out := outlineMask(fill, r)
	if out.w != fill.w+2*r || out.h != fill.h+2*r {
		t.Fatalf("outline is %dx%d, want %dx%d", out.w, out.h, fill.w+2*r, fill.h+2*r)
	}
	if s, n := hollowScore(fill, out, r+1); n == 0 || s > 0.01 {
		t.Errorf("outline not hollow: interior covered %.3f over %d px", s, n)
	}
	// The ring reaches about r px beyond the disc on each side.
	if d := out.box.Dx() - fill.box.Dx(); d < 2*r-1 || d > 2*r+1 {
		t.Errorf("outline box %v vs fill box %v: grew by %d px, want about %d", out.box, fill.box, d, 2*r)
	}
	fcx, fcy := fill.centre()
	ocx, ocy := out.centre()
	if math.Abs(ocx-r-fcx) > 0.5 || math.Abs(ocy-r-fcy) > 0.5 {
		t.Errorf("ink centre moved: fill (%.1f,%.1f), outline (%.1f,%.1f) minus padding %d", fcx, fcy, ocx, ocy, r)
	}
	// On the edge the ring is opaque; at the centre it is empty.
	if out.at(40+r, 40+r) != 0 {
		t.Errorf("centre of the ring is not empty")
	}
	if out.at(40+25+r, 40+r) < 200 {
		t.Errorf("edge of the ring is not opaque: %d", out.at(40+25+r, 40+r))
	}
}

func TestArrowGeometry(t *testing.T) {
	near := func(a, b float32) bool { return math.Abs(float64(a-b)) < 1e-3 }
	cases := []struct {
		theta  float64
		tx, ty float32
	}{
		{0, 0, 100},    // up
		{90, 100, 0},   // right (clockwise from up)
		{180, 0, -100}, // down
		{270, -100, 0}, // left
	}
	for _, c := range cases {
		tail, base, head := arrowGeometry(c.theta, 100, 28, 16)
		tip := head[0]
		if !near(tip.X, c.tx) || !near(tip.Y, c.ty) {
			t.Errorf("θ=%v: tip (%v,%v), want (%v,%v)", c.theta, tip.X, tip.Y, c.tx, c.ty)
		}
		if !near(tail.X, -c.tx) || !near(tail.Y, -c.ty) {
			t.Errorf("θ=%v: tail (%v,%v) not opposite the tip", c.theta, tail.X, tail.Y)
		}
		if !near(base.X, c.tx*0.72) || !near(base.Y, c.ty*0.72) {
			t.Errorf("θ=%v: head base (%v,%v)", c.theta, base.X, base.Y)
		}
	}
}

func TestDepartureOf(t *testing.T) {
	for o, want := range map[int]int{0: 0, 60: 60, 120: 120, 180: 180, 240: 120, 300: 60, 360: 0, -60: 60} {
		if got := DepartureOf(o); got != want {
			t.Errorf("DepartureOf(%d) = %d, want %d", o, got, want)
		}
	}
	got := Orientations([]int{0, 60, 120, 180})
	want := []int{0, 60, 120, 180, 240, 300}
	if len(got) != len(want) {
		t.Fatalf("Orientations = %v", got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("Orientations = %v, want %v", got, want)
		}
	}
}
