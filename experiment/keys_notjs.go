// Copyright (2026) Thiago Lizier Rissi

//go:build !js

package main

import (
	"unsafe"

	"github.com/Zyko0/go-sdl3/sdl"
	"github.com/chrplr/goxpyriment/control"
)

// pushKey puts a KEY_DOWN event for key in the SDL queue. SDL_PushEvent is
// thread-safe, so this may run on any goroutine.
func pushKey(key control.Keycode) {
	var ev sdl.Event
	ke := (*sdl.KeyboardEvent)(unsafe.Pointer(&ev))
	ke.Type = sdl.EVENT_KEY_DOWN
	ke.Timestamp = sdl.TicksNS()
	ke.Key = key
	ke.Down = true
	_ = sdl.PushEvent(&ev)
}

// typeText does nothing on the desktop: a TEXT_INPUT event carries a C string
// that cannot be pushed from Go safely, so the autopilot answers the final
// question with an empty answer there.
func typeText(string) {}
