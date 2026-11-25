// Copyright (2026) Thiago Lizier Rissi

package main

// A free-text answer box. goxpyriment's stimuli.TextInput is a single line
// with no wrapping, and it creates new text textures on every redraw without
// releasing them; an answer of a few sentences needs wrapping, so this is a
// small local loop on the SDL event queue (same approach as the framework's
// examples/Typing-Speed).

import (
	"time"

	"github.com/chrplr/goxpyriment/apparatus"
	"github.com/chrplr/goxpyriment/control"
	"github.com/chrplr/goxpyriment/stimuli"
)

const maxAnswerRunes = 2000

// askText shows prompt and an answer box and returns what was typed when
// ENTER is pressed. ESC or closing the window returns the text so far together
// with control.EndLoop.
func (r *runner) askText(prompt, hint string) (string, error) {
	exp := r.exp
	if err := drainEvents(); err != nil { // forget keys pressed before the question
		return "", err
	}
	win := exp.Screen.Window
	if err := win.StartTextInput(); err != nil {
		return "", err
	}
	defer func() { _ = win.StopTextInput() }()

	w, h := exp.DrawArea()
	question := stimuli.NewTextBox(prompt, int32(0.9*w), control.Point(0, 0.25*h), exp.ForegroundColor)
	defer question.Unload()
	hintLine := stimuli.NewTextLine(hint, 0, -0.42*h, exp.ForegroundColor)
	defer hintLine.Unload()

	var runes []rune
	var answer *stimuli.TextBox
	defer func() {
		if answer != nil {
			_ = answer.Unload()
		}
	}()

	if r.auto != nil {
		r.auto.typeThenEnter("autopilot answer")
	}

	if r.auto != nil {
		r.auto.snapshot(exp, "debrief")
	}
	redraw := true
	for {
		if redraw {
			if answer != nil {
				if err := answer.Unload(); err != nil {
					return string(runes), err
				}
			}
			answer = stimuli.NewTextBox(string(runes)+"_", int32(0.85*w), control.Point(0, -0.08*h), exp.ForegroundColor)
			scene := stimuli.DrawFunc(func(s *apparatus.Screen) error {
				if err := question.Draw(s); err != nil {
					return err
				}
				if err := answer.Draw(s); err != nil {
					return err
				}
				return hintLine.Draw(s)
			})
			if err := exp.Show(scene); err != nil {
				return string(runes), err
			}
			redraw = false
		}

		var ev control.Event
		got := false
		for control.PollEvent(&ev) {
			got = true
			switch ev.Type {
			case control.EVENT_QUIT:
				return string(runes), control.EndLoop
			case control.EVENT_KEY_DOWN:
				switch ev.KeyboardEvent().Key {
				case control.K_ESCAPE:
					return string(runes), control.EndLoop
				case control.K_RETURN, control.K_KP_ENTER:
					return string(runes), nil
				case control.K_BACKSPACE:
					if len(runes) > 0 {
						runes = runes[:len(runes)-1]
						redraw = true
					}
				}
			case control.EVENT_TEXT_INPUT:
				if len(runes) < maxAnswerRunes {
					runes = append(runes, []rune(ev.TextInputEvent().Text)...)
					redraw = true
				}
			}
		}
		if !got {
			time.Sleep(10 * time.Millisecond)
		}
	}
}
