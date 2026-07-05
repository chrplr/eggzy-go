package main

import (
	"fmt"

	"github.com/Zyko0/go-sdl3/sdl"
	"github.com/chrplr/pgzgo"
)

// Assets embeds the pgzgo Screen — so the image cache and the common helpers
// (Blit, BlitTile, FillRect, Fill, Size, Destroy) are promoted directly — and
// adds eggzy's two specifics: tileset-texture loading from the embedded tilemaps
// filesystem, and the sprite-font text used across menus and the HUD.
type Assets struct {
	*pgzgo.Screen
}

// TilesetTexture loads a tileset PNG from the embedded tilemaps filesystem
// (e.g. "tilemaps/tiles_forest.png"), cached by path.
func (a *Assets) TilesetTexture(path string) *sdl.Texture {
	return a.LoadTexture(tilemapsFS, path)
}

// TextAlign selects horizontal text alignment. Its values match pgzgo.Align.
type TextAlign int

const (
	AlignLeft TextAlign = iota
	AlignCentre
	AlignRight
)

// eggzyFont builds a sprite font: glyphs are named "<font>NNN" by zero-padded
// codepoint, the controller-button images stand in for '%' and '#', and a space
// advances 22px.
func eggzyFont(font string) pgzgo.Font {
	return pgzgo.Font{
		Space: 22,
		Name: func(r rune) string {
			switch r {
			case '%':
				return "xb_a"
			case '#':
				return "xb_b"
			}
			return fmt.Sprintf("%s%03d", font, int(r))
		},
	}
}

// DrawText draws text using a sprite font with the given alignment.
func (a *Assets) DrawText(text string, x, y float64, align TextAlign, font string) {
	a.Screen.DrawText(text, x, y, pgzgo.Align(align), eggzyFont(font))
}
