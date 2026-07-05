package main

import "fmt"

// Animation plays a sequence of sprites (e.g. gem pickup, losing a life). The
// format string uses a single %d for the frame number.
type Animation struct {
	Actor
	formatStr     string
	numFrames     int
	frameInterval int
	timer         int
	riseTime      int
}

func NewAnimation(x, y float64, formatStr string, numFrames, frameInterval int, anchor Anchor, initialDelay, riseTime int) *Animation {
	an := &Animation{
		formatStr:     formatStr,
		numFrames:     numFrames,
		frameInterval: frameInterval,
		timer:         -initialDelay,
		riseTime:      riseTime,
	}
	an.Actor = newActor("blank", x, y, anchor)
	an.updateImage()
	return an
}

// NewDashTrail creates the fading trail left behind by the player's dash.
func NewDashTrail(x, y float64, image string) *Animation {
	return NewAnimation(x, y, image+"_trail_%d", 6, 5, AnchorPlayer, 0, -1)
}

func (an *Animation) updateImage() {
	if an.timer < 0 {
		an.Image = "blank"
		return
	}
	frame := an.timer / an.frameInterval
	if frame > an.numFrames-1 {
		frame = an.numFrames - 1
	}
	an.Image = fmt.Sprintf(an.formatStr, frame)
}

func (an *Animation) Update(g *Game) {
	an.timer++
	an.updateImage()
	if an.riseTime > -1 && an.timer > an.riseTime {
		an.Y -= 1
	}
}

func (an *Animation) finished() bool {
	return an.timer/an.frameInterval >= an.numFrames
}

func (an *Animation) Draw(g *Game) { an.drawImage(g.assets) }
