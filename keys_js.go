// Copyright (2026) Thiago Lizier Rissi

//go:build js

package main

// In the browser, SDL_PushEvent is not implemented by the go-sdl3 fork, so the
// autopilot dispatches DOM keyboard events instead. SDL's Emscripten backend
// listens to those, so they reach the experiment exactly as a real key press
// does: converted to an SDL event, timestamped by SDL, read by GetKeyEventTS.

import (
	"strings"
	"syscall/js"

	"github.com/chrplr/goxpyriment/control"
)

// pressJS dispatches one synthetic DOM keyboard event on the canvas (it bubbles
// to document and window, wherever SDL listens). keyCode, charCode and which
// are legacy fields that the KeyboardEvent constructor ignores, so they are
// defined on the event object.
const pressJS = `(function (type, key, code, keyCode, charCode) {
  var ev = new KeyboardEvent(type, {key: key, code: code, bubbles: true, cancelable: true});
  Object.defineProperty(ev, 'keyCode', {get: function () { return keyCode; }});
  Object.defineProperty(ev, 'which', {get: function () { return keyCode; }});
  Object.defineProperty(ev, 'charCode', {get: function () { return charCode; }});
  var target = document.getElementById('canvas') || window;
  target.dispatchEvent(ev);
})`

var pressFn js.Value

type domKey struct {
	key, code string
	keyCode   int
}

var domKeys = map[control.Keycode]domKey{
	control.K_F:      {"f", "KeyF", 70},
	control.K_J:      {"j", "KeyJ", 74},
	control.K_SPACE:  {" ", "Space", 32},
	control.K_RETURN: {"Enter", "Enter", 13},
}

func press(k domKey, char rune) {
	if pressFn.IsUndefined() {
		pressFn = js.Global().Call("eval", pressJS)
	}
	pressFn.Invoke("keydown", k.key, k.code, k.keyCode, 0)
	if char != 0 {
		pressFn.Invoke("keypress", k.key, k.code, int(char), int(char))
	}
	pressFn.Invoke("keyup", k.key, k.code, k.keyCode, 0)
}

// pushKey presses key.
func pushKey(key control.Keycode) {
	if k, ok := domKeys[key]; ok {
		press(k, 0)
	}
}

// typeText types letters and spaces, one key press each.
func typeText(text string) {
	for _, r := range text {
		switch {
		case r == ' ':
			press(domKey{" ", "Space", 32}, ' ')
		case r >= 'a' && r <= 'z':
			up := strings.ToUpper(string(r))
			press(domKey{string(r), "Key" + up, int(up[0])}, r)
		}
	}
}
