// Copyright (2026) Thiago Lizier Rissi

package main

// -autopilot (?autopilot in the browser): for testing the program without a
// person at the keyboard. Key presses enter where a real one would, so they
// travel the same path to GetKeyEventTS (see pushKey in keys_notjs.go and
// keys_js.go). Responses are random (about 10% errors, 3% timeouts). The run
// is flagged in the -info.txt.

import (
	"fmt"
	"math/rand/v2"
	"path/filepath"
	"runtime"
	"time"

	"github.com/chrplr/goxpyriment/control"
)

type autopilot struct {
	rng   *rand.Rand
	shots int             // screenshots taken so far
	once  map[string]bool // names already captured by snapshotOnce
}

// snapshotOnce is snapshot for the first occurrence of name only.
func (a *autopilot) snapshotOnce(exp *control.Experiment, name string) {
	if a.once == nil {
		a.once = map[string]bool{}
	}
	if !a.once[name] {
		a.once[name] = true
		a.snapshot(exp, name)
	}
}

// maxShots bounds the screenshots an autopilot run saves (instruction screens
// and the displays of the first trial of each block).
const maxShots = 40

// snapshot saves the next frame as a PNG in the data directory (desktop only:
// the browser has no filesystem to save it to).
func (a *autopilot) snapshot(exp *control.Experiment, name string) {
	if a.shots >= maxShots || runtime.GOOS == "js" {
		return
	}
	a.shots++
	exp.Screen.ScreenshotNextFrame(filepath.Join(exp.Data.Directory,
		fmt.Sprintf("autopilot_%02d_%s.png", a.shots, name)))
}

func newAutopilot() *autopilot {
	return &autopilot{rng: rand.New(rand.NewPCG(uint64(time.Now().UnixNano()), 1))}
}

// pressAfter presses key after d, from another goroutine.
func (a *autopilot) pressAfter(key control.Keycode, d time.Duration) {
	go func() {
		time.Sleep(d)
		pushKey(key)
	}()
}

// typeThenEnter types text (where the platform allows it, see typeText) and
// then presses ENTER, from another goroutine.
func (a *autopilot) typeThenEnter(text string) {
	go func() {
		time.Sleep(300 * time.Millisecond)
		typeText(text)
		time.Sleep(300 * time.Millisecond)
		pushKey(control.K_RETURN)
	}()
}

// respond schedules a response to a trial whose test is on screen now: RT grows
// with the angular departure, as a participant's would.
func (a *autopilot) respond(t Trial, m KeyMapping) {
	if a.rng.Float64() < 0.03 {
		return // timeout
	}
	key := m.NormalKey
	if t.Version == Backward {
		key = m.BackwardKey
	}
	if a.rng.Float64() < 0.10 {
		if key == m.NormalKey {
			key = m.BackwardKey
		} else {
			key = m.NormalKey
		}
	}
	code := control.K_F
	if key == "j" {
		code = control.K_J
	}
	rt := 450*time.Millisecond + time.Duration(t.Departure)*2*time.Millisecond +
		time.Duration(a.rng.IntN(200))*time.Millisecond
	a.pressAfter(code, rt)
}
