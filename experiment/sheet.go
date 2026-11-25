// Copyright (2026) Thiago Lizier Rissi

package main

// -sheet: a stimulus check sheet. Every character, normal and backward, at one
// orientation, plus the identity cue, the condition-C cue (alone and overlaid
// on the test character) and the arrow, each in its own aperture with a
// crosshair through the aperture centre, so that rotation, mirroring and
// centring can be checked by eye.
//
//	LEFT/RIGHT  orientation -/+ 30°
//	UP/DOWN     character of the bottom row
//	P           save a screenshot (PNG) in the data directory
//	SPACE, ESC  quit

import (
	"fmt"
	"log"
	"path/filepath"

	"github.com/Zyko0/go-sdl3/sdl"
	"github.com/chrplr/goxpyriment/apparatus"
	"github.com/chrplr/goxpyriment/control"
	"github.com/chrplr/goxpyriment/stimuli"
)

func runSheet(exp *control.Experiment, cfg Config, ap *autopilot) error {
	st, _, err := setupDisplay(exp, cfg)
	if err != nil {
		return err
	}
	defer st.Destroy()

	w, h := exp.DrawArea()
	font, err := control.FontFromMemory(fontData, max(12, h/50))
	if err != nil {
		return fmt.Errorf("label font: %w", err)
	}
	defer font.Close()

	nCols := max(6, len(cfg.Characters))
	cellW := 0.96 * w / float32(nCols)
	top := h/2 - 0.09*h
	rowH := (top + h/2) / 3
	scale := min(0.85, 0.9*cellW/(2*st.ApR), 0.72*rowH/(2*st.ApR))
	r := st.ApR * scale
	colX := func(i int) float32 { return -0.48*w + cellW*(float32(i)+0.5) }
	rowY := func(i int) float32 { return top - (float32(i)+0.45)*rowH }

	cross := []*stimuli.PolyLine{
		stimuli.NewPolyLine([]sdl.FPoint{{X: -r, Y: 0}, {X: r, Y: 0}}, false, 1, ringColor),
		stimuli.NewPolyLine([]sdl.FPoint{{X: 0, Y: -r}, {X: 0, Y: r}}, false, 1, ringColor),
	}
	cell := func(s *apparatus.Screen, x, y float32) error {
		if err := st.DrawAperture(s, x, y, scale); err != nil {
			return err
		}
		for _, c := range cross {
			c.SetPosition(sdl.FPoint{X: x, Y: y})
			if err := c.Draw(s); err != nil {
				return err
			}
		}
		return nil
	}

	var labels textLines
	defer func() { labels.unload() }()
	newLabel := func(s string, x, y float32) *stimuli.TextLine {
		l := labels.add(stimuli.NewTextLine(s, x, y, exp.ForegroundColor))
		l.Font = font
		return l
	}

	orient, chIdx := 0, 2%len(cfg.Characters)
	build := func() stimuli.DrawFunc {
		labels.unload()
		labels = labels[:0]
		ch := cfg.Characters[chIdx]
		newLabel(fmt.Sprintf("orientation %d° clockwise (departure %d°)", orient, DepartureOf(orient)),
			0, h/2-0.025*h)
		newLabel("LEFT/RIGHT: ±30°   UP/DOWN: bottom-row character   P: screenshot   SPACE: quit",
			0, h/2-0.025*h-float32(font.Height())*1.1)
		below := func(y float32) float32 { return y - r - float32(font.Height())*0.7 }
		for i, c := range cfg.Characters {
			newLabel(c+" normal", colX(i), below(rowY(0)))
			newLabel(c+" backward", colX(i), below(rowY(1)))
		}
		bottom := []string{
			ch + " identity cue", ch + fmt.Sprintf(" C cue %d°", orient), "C cue + test",
			fmt.Sprintf("arrow %d°", orient), "arrow + test", "arrow + backward",
		}
		for i, s := range bottom {
			newLabel(s, colX(i), below(rowY(2)))
		}
		return func(s *apparatus.Screen) error {
			for _, l := range labels {
				if err := l.Draw(s); err != nil {
					return err
				}
			}
			for i, c := range cfg.Characters {
				for row, backward := range []bool{false, true} {
					x, y := colX(i), rowY(row)
					if err := cell(s, x, y); err != nil {
						return err
					}
					if err := st.DrawTest(s, x, y, scale, c, orient, backward); err != nil {
						return err
					}
				}
			}
			y := rowY(2)
			steps := []func(x float32) error{
				func(x float32) error { return st.DrawCue(s, x, y, scale, ch, 0) },
				func(x float32) error { return st.DrawCue(s, x, y, scale, ch, orient) },
				func(x float32) error {
					if err := st.DrawTest(s, x, y, scale, ch, orient, false); err != nil {
						return err
					}
					return st.DrawCue(s, x, y, scale, ch, orient)
				},
				func(x float32) error { return st.DrawArrow(s, x, y, scale, orient) },
				func(x float32) error {
					if err := st.DrawArrow(s, x, y, scale, orient); err != nil {
						return err
					}
					return st.DrawTest(s, x, y, scale, ch, orient, false)
				},
				func(x float32) error {
					if err := st.DrawArrow(s, x, y, scale, orient); err != nil {
						return err
					}
					return st.DrawTest(s, x, y, scale, ch, orient, true)
				},
			}
			for i, f := range steps {
				if err := cell(s, colX(i), y); err != nil {
					return err
				}
				if err := f(colX(i)); err != nil {
					return err
				}
			}
			return nil
		}
	}

	shot := func() string {
		return filepath.Join(exp.Data.Directory, fmt.Sprintf("sheet_%s_%03d.png", cfg.Characters[chIdx], orient))
	}

	// Autopilot: one screenshot per experimental orientation, then quit.
	if ap != nil {
		for _, o := range Orientations(cfg.Departures) {
			orient = o
			exp.Screen.ScreenshotNextFrame(shot())
			if err := exp.Show(build()); err != nil {
				return err
			}
			log.Printf("sheet screenshot: %s", shot())
		}
		return nil
	}

	for {
		if err := exp.Show(build()); err != nil {
			return err
		}
		key, err := exp.Keyboard.WaitKeys([]control.Keycode{
			control.K_LEFT, control.K_RIGHT, control.K_UP, control.K_DOWN, control.K_P, control.K_SPACE}, -1)
		if err != nil {
			return err
		}
		switch key {
		case control.K_LEFT:
			orient = (orient + 330) % 360
		case control.K_RIGHT:
			orient = (orient + 30) % 360
		case control.K_UP:
			chIdx = (chIdx + 1) % len(cfg.Characters)
		case control.K_DOWN:
			chIdx = (chIdx + len(cfg.Characters) - 1) % len(cfg.Characters)
		case control.K_P:
			exp.Screen.ScreenshotNextFrame(shot())
			log.Printf("screenshot: %s", shot())
		case control.K_SPACE:
			return nil
		}
	}
}

// textLines collects text stimuli so they can be unloaded together.
type textLines []*stimuli.TextLine

func (ts *textLines) add(t *stimuli.TextLine) *stimuli.TextLine { *ts = append(*ts, t); return t }

func (ts textLines) unload() {
	for _, t := range ts {
		_ = t.Unload()
	}
}
