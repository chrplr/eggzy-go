package main

// Glue between the game and the pgzgo harness. The images/sounds/music embeds
// live here (the tilemaps embed stays in tilemap.go, near the map loader);
// input.go's Controls read the harness keyboard/gamepad through these wrappers.

import (
	"embed"

	"github.com/Zyko0/go-sdl3/sdl"
	"github.com/chrplr/pgzgo"
)

// Audio is the game's name for the harness mixer. (Assets is a struct in
// assets.go, since it extends the Screen with tileset and text helpers.)
type Audio = pgzgo.Audio

//go:embed images
var imagesFS embed.FS

//go:embed sounds music
var audioFS embed.FS

// app is the running harness; the input wrappers read its per-frame snapshots.
var app *pgzgo.App

// Keyboard bindings (held and rising-edge).
func keyDown(sc sdl.Scancode) bool        { return app.Keyboard.Held(sc) }
func keyJustPressed(sc sdl.Scancode) bool { return app.Keyboard.Pressed(sc) }

// Gamepad bindings used by Controls.
func padLeft() bool           { return app.Gamepad.Left() }
func padRight() bool          { return app.Gamepad.Right() }
func padUp() bool             { return app.Gamepad.Up() }
func padDown() bool           { return app.Gamepad.Down() }
func padAxisX() float64       { return app.Gamepad.AxisX() }
func padAxisY() float64       { return app.Gamepad.AxisY() }
func padButton0() bool        { return app.Gamepad.Button0() }
func padButton1() bool        { return app.Gamepad.Button1() }
func padButton0Pressed() bool { return app.Gamepad.Button0Pressed() }
func padButton1Pressed() bool { return app.Gamepad.Button1Pressed() }
