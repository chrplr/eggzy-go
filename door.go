package main

import "fmt"

// Door blocks the level exit until every gem is collected, then plays an open
// animation.
type Door struct {
	Actor
	biome     string
	variant   string
	opening   bool
	lastFrame int
	frame     int
}

func NewDoor(x, y float64, biome, variant string, alreadyOpen bool) *Door {
	last := 13
	if biome == "castle" {
		last = 15
	}
	frame := 0
	if alreadyOpen {
		frame = last
	}
	d := &Door{biome: biome, variant: variant, opening: alreadyOpen, lastFrame: last, frame: frame}
	d.Actor = newActor(d.spriteName(), x, y, AnchorTopLeft)
	return d
}

func (d *Door) spriteName() string {
	return fmt.Sprintf("door_%s_%s_%d", d.biome, d.variant, d.frame)
}

func (d *Door) Update(g *Game) {
	if d.opening && d.frame < d.lastFrame && g.timer%3 == 0 {
		d.frame++
		d.Image = d.spriteName()
	}
}

func (d *Door) open()             { d.opening = true }
func (d *Door) isFullyOpen() bool { return d.frame == d.lastFrame }

func (d *Door) Draw(g *Game) { d.drawImage(g.assets) }
