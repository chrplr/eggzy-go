package main

import "strconv"

// gemNextType cycles gem types 1..4 across the game (a class variable in Python).
var gemNextType = 1

func gemNewGame() { gemNextType = 1 }

// Gem is a pickup that grants bonus time.
type Gem struct {
	Actor
	gtype     int
	collected bool
}

func NewGem(x, y float64) *Gem {
	t := gemNextType
	gemNextType++
	if gemNextType >= 5 {
		gemNextType = 1
	}
	return &Gem{Actor: newActor("blank", x, y, AnchorCentreBottom), gtype: t}
}

func (gm *Gem) Update(g *Game) {
	cx, cy := gm.centerx(g.assets), gm.centery(g.assets)
	if g.player != nil && g.player.collidepoint(g.assets, cx, cy) {
		g.gainTime(g.timePickupBonus, cx, cy)
		g.playSound("collect", 1)
		gm.collected = true
	}
	animFrame := (g.timer / 6) % 4
	gm.Image = "gem" + strconv.Itoa(gm.gtype) + "_" + strconv.Itoa(animFrame)
}

func (gm *Gem) Draw(g *Game) { gm.drawImage(g.assets) }
