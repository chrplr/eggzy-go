package main

import "math"

// Rect is an axis-aligned rectangle with pygame-Rect-style collision tests.
type Rect struct {
	X, Y, W, H float64
}

func (r Rect) Left() float64   { return r.X }
func (r Rect) Top() float64    { return r.Y }
func (r Rect) Right() float64  { return r.X + r.W }
func (r Rect) Bottom() float64 { return r.Y + r.H }

// colliderect reports whether two rectangles overlap (touching edges do not count).
func (r Rect) colliderect(o Rect) bool {
	return r.X < o.X+o.W && r.X+r.W > o.X && r.Y < o.Y+o.H && r.Y+r.H > o.Y
}

// collidepoint reports whether (px, py) lies inside the rectangle.
func (r Rect) collidepoint(px, py float64) bool {
	return px >= r.X && px < r.X+r.W && py >= r.Y && py < r.Y+r.H
}

func moveTowards(n, target, speed float64) float64 {
	if n < target {
		return math.Min(n+speed, target)
	}
	return math.Max(n-speed, target)
}

func sign(x float64) float64 {
	if x == 0 {
		return 0
	}
	if x < 0 {
		return -1
	}
	return 1
}

func absf(x float64) float64 { return math.Abs(x) }
