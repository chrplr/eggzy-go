package main

import "math"

// Anchor kinds for each axis.
const (
	akCenter = 0
	akBottom = 1
	akAbs    = 2
)

// Anchor describes how a sprite is pinned to its (X, Y) position, mirroring
// Pygame Zero's anchor tuples like ("center", "bottom") or ("center", 60).
type Anchor struct {
	xKind, yKind int
	xVal, yVal   float64
}

var (
	AnchorCentre       = Anchor{akCenter, akCenter, 0, 0}
	AnchorCentreBottom = Anchor{akCenter, akBottom, 0, 0}
	AnchorPlayer       = Anchor{akCenter, akAbs, 0, 60}
	AnchorFlame        = Anchor{akCenter, akAbs, 0, 78}
	AnchorFlameDash    = Anchor{akCenter, akAbs, 0, 130}
	AnchorTopLeft      = Anchor{akAbs, akAbs, 0, 0}
)

func anchorAbsY(v float64) Anchor { return Anchor{akCenter, akAbs, 0, v} }

func (an Anchor) offset(w, h float64) (float64, float64) {
	ax := w / 2
	if an.xKind == akAbs {
		ax = an.xVal
	}
	var ay float64
	switch an.yKind {
	case akCenter:
		ay = h / 2
	case akBottom:
		ay = h
	case akAbs:
		ay = an.yVal
	}
	return ax, ay
}

// Actor is a positioned sprite with an anchor, the base for all game objects.
type Actor struct {
	X, Y   float64
	Image  string
	anchor Anchor
}

func newActor(image string, x, y float64, anchor Anchor) Actor {
	return Actor{X: x, Y: y, Image: image, anchor: anchor}
}

func (a *Actor) anchorOffset(as *Assets) (float64, float64) {
	w, h := as.Size(a.Image)
	return a.anchor.offset(w, h)
}

// drawImage blits the sprite at its anchored screen position.
func (a *Actor) drawImage(as *Assets) {
	ax, ay := a.anchorOffset(as)
	as.Blit(a.Image, a.X-ax, a.Y-ay)
}

// spriteRect returns the full sprite bounding rectangle.
func (a *Actor) spriteRect(as *Assets) Rect {
	ax, ay := a.anchorOffset(as)
	w, h := as.Size(a.Image)
	return Rect{a.X - ax, a.Y - ay, w, h}
}

func (a *Actor) top(as *Assets) float64 {
	_, ay := a.anchorOffset(as)
	return a.Y - ay
}

func (a *Actor) centerx(as *Assets) float64 {
	r := a.spriteRect(as)
	return r.X + r.W/2
}

func (a *Actor) centery(as *Assets) float64 {
	r := a.spriteRect(as)
	return r.Y + r.H/2
}

func (a *Actor) collidepoint(as *Assets, px, py float64) bool {
	return a.spriteRect(as).collidepoint(px, py)
}

// collidable lets CollideActor query the concrete type's collision box size,
// which for the player and enemies is smaller than the sprite.
type collidable interface {
	collidableWidth(as *Assets) float64
	collidableHeight(as *Assets) float64
}

// CollideActor moves through the level one pixel at a time, stopping at walls.
type CollideActor struct {
	Actor
	self collidable
}

// move steps up to speed pixels along (dx, dy); returns true if a wall blocked it.
func (c *CollideActor) move(g *Game, dx, dy, speed float64) bool {
	newX, newY := c.X, c.Y
	steps := int(speed)
	for i := 0; i < steps; i++ {
		newX += dx
		newY += dy
		if g.positionBlocked(c.getRectAt(g.assets, newX, newY)) {
			return true
		}
		c.X, c.Y = newX, newY
	}
	return false
}

// getRectAt returns the collision rectangle as if the actor were at (cx, cy),
// where cx is the horizontal centre and cy the bottom.
func (c *CollideActor) getRectAt(as *Assets, cx, cy float64) Rect {
	w := c.self.collidableWidth(as)
	h := c.self.collidableHeight(as)
	return Rect{cx - math.Floor(w/2), cy - h, w, h}
}

func (c *CollideActor) getRect(as *Assets) Rect {
	return c.getRectAt(as, c.X, c.Y)
}

// FallState tracks the vertical movement mode of a gravity actor.
type FallState int

const (
	FallLanded FallState = iota
	FallFalling
	FallJumping
	FallWallJumping
)

// GravityActor is a CollideActor subject to gravity (player and enemies).
type GravityActor struct {
	CollideActor
	gravityEnabled    bool
	velY              float64
	fallState         FallState
	lowerGravityTimer int
}

func (ga *GravityActor) gravUpdate(g *Game, detect bool) {
	if !ga.gravityEnabled {
		return
	}

	ga.lowerGravityTimer--

	mod := 2
	if ga.lowerGravityTimer > 0 {
		mod = 3
	}
	if g.timer%mod == 0 {
		ga.velY = math.Min(ga.velY+1, GravityMaxFallSpeed)
	}

	if detect && ga.velY != 0 {
		if ga.fallState == FallLanded {
			ga.fallState = FallFalling
		}
		if ga.move(g, 0, sign(ga.velY), absf(ga.velY)) {
			// Landed or hit head on ceiling
			if ga.velY > 0 {
				ga.velY = 0
				ga.fallState = FallLanded
			}
		}
	} else {
		ga.Y += ga.velY
	}
}

func (ga *GravityActor) landed() bool {
	return ga.fallState == FallLanded
}
