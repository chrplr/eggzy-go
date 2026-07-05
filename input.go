package main

import "github.com/Zyko0/go-sdl3/sdl"

// Keyboard/gamepad snapshotting lives in the pgzgo harness; keyDown,
// keyJustPressed and the pad* helpers are thin wrappers over it (see harness.go).

// Controls models the player's movement axes and two action buttons
// (button 0 = jump, button 1 = dash).
type Controls struct{}

// gamepadDeadZone matches the Python JoystickControls dead-zone for eggzy.
const gamepadDeadZone = 0.6

// The axis getters return a digital -1/0/1, reading the keyboard or, failing
// that, the gamepad (d-pad first, then the left stick past the dead-zone).
func (c Controls) getX() float64 {
	if keyDown(sdl.SCANCODE_LEFT) {
		return -1
	} else if keyDown(sdl.SCANCODE_RIGHT) {
		return 1
	}
	if padLeft() {
		return -1
	} else if padRight() {
		return 1
	}
	if ax := padAxisX(); ax <= -gamepadDeadZone {
		return -1
	} else if ax >= gamepadDeadZone {
		return 1
	}
	return 0
}

func (c Controls) getY() float64 {
	if keyDown(sdl.SCANCODE_UP) {
		return -1
	} else if keyDown(sdl.SCANCODE_DOWN) {
		return 1
	}
	if padUp() {
		return -1
	} else if padDown() {
		return 1
	}
	if ay := padAxisY(); ay <= -gamepadDeadZone {
		return -1
	} else if ay >= gamepadDeadZone {
		return 1
	}
	return 0
}

func (c Controls) buttonDown(button int) bool {
	if button == 0 {
		return keyDown(sdl.SCANCODE_SPACE) || keyDown(sdl.SCANCODE_RETURN) || keyDown(sdl.SCANCODE_KP_ENTER) || padButton0()
	}
	return keyDown(sdl.SCANCODE_Z) || padButton1()
}

func (c Controls) buttonPressed(button int) bool {
	if button == 0 {
		return keyJustPressed(sdl.SCANCODE_SPACE) || keyJustPressed(sdl.SCANCODE_RETURN) || keyJustPressed(sdl.SCANCODE_KP_ENTER) || padButton0Pressed()
	}
	return keyJustPressed(sdl.SCANCODE_Z) || padButton1Pressed()
}

// buttonName returns the label used in tutorial text for the "dash"/"jump" action.
func (c Controls) buttonName(button string) string {
	switch button {
	case "dash":
		return "Z"
	case "jump":
		return "SPACE"
	}
	return "?"
}
