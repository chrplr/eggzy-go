package main

// GhostPlayer replays a previous run's recorded frames as a translucent "ghost".
type GhostPlayer struct {
	Actor
	replayData  []ReplayFrame
	replayFrame int
	level       int
}

func NewGhostPlayer(data []ReplayFrame) *GhostPlayer {
	gp := &GhostPlayer{replayData: data}
	x, y := 0.0, 0.0
	if len(data) > 0 {
		x, y = data[0].X, data[0].Y
	}
	gp.Actor = newActor("blank", x, y, AnchorPlayer)
	return gp
}

func (gp *GhostPlayer) Update(g *Game) {
	gp.replayFrame++
	if gp.replayFrame < len(gp.replayData) {
		f := gp.replayData[gp.replayFrame]
		gp.X, gp.Y, gp.level = f.X, f.Y, f.Level
		if f.Sprite == "blank" {
			gp.Image = "blank"
		} else {
			gp.Image = "ghost_" + f.Sprite
		}
	}
}

func (gp *GhostPlayer) Draw(g *Game) {
	// Only draw if the ghost is on the same level as the real player.
	if gp.level == g.levelIndex {
		gp.drawImage(g.assets)
	}
}
