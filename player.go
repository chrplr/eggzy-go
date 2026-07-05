package main

import (
	"fmt"
	"math"
	"strconv"
)

// Player is the egg character: runs, jumps, wall-jumps, dashes and stomps enemies.
type Player struct {
	GravityActor
	controls Controls
	flame    Actor

	velX    float64
	facingX float64
	hurt    bool

	dashTimer          int
	dashAnimationTimer int
	dashAllowed        bool

	grabbedWall          float64
	previousGrabbedWall  float64
	coyoteTime           int
	fallTimer            int
	wallJumpCoyoteTime   int
	cachedJumpInputTimer int
	enemyStompedTimer    int
	changeDirectionTimer int

	lastDashSprite   string
	stompedLastFrame bool

	startPos   [2]float64
	replayData []ReplayFrame
}

func NewPlayer(controls Controls) *Player {
	p := &Player{
		controls:       controls,
		facingX:        1,
		dashTimer:      DashTimerTrailCutoff,
		lastDashSprite: "dash_horizontal_0_0",
	}
	p.Actor = newActor("blank", 0, 0, AnchorPlayer)
	p.gravityEnabled = true
	p.fallState = FallFalling
	p.self = p
	p.flame = newActor("flame_stand_0", 0, 0, AnchorFlame)
	return p
}

func (p *Player) collidableWidth(as *Assets) float64  { return PlayerWidth }
func (p *Player) collidableHeight(as *Assets) float64 { return PlayerHeight }

func (p *Player) newLevel(startPos [2]float64) {
	p.startPos = startPos
}

func (p *Player) reset(g *Game) {
	p.X, p.Y = p.startPos[0], p.startPos[1]
	p.velX = 0
	p.velY = 0
	p.facingX = 1
	p.hurt = false
	p.dashTimer = DashTimerTrailCutoff
	p.gravityEnabled = true
	p.grabbedWall = 0
	p.coyoteTime = 0
	p.wallJumpCoyoteTime = 0
	p.cachedJumpInputTimer = 0
	p.enemyStompedTimer = 0

	// Clear any enemies at/near the spawn, as if we'd stomped them.
	for _, enemy := range g.enemies {
		if math.Hypot(p.X-enemy.X, p.Y-enemy.Y) < 150 {
			enemy.destroy(g)
			g.playSound("enemy_death", 5)
		}
	}
}

func (p *Player) hitTest(g *Game, e *Enemy) bool {
	return p.getRectAt(g.assets, p.X, p.Y).colliderect(e.getRect(g.assets)) && !p.hurt
}

func (p *Player) getCollidingEnemies(g *Game) []*Enemy {
	var out []*Enemy
	for _, e := range g.enemies {
		if !e.dying && p.hitTest(g, e) {
			out = append(out, e)
		}
	}
	return out
}

func (p *Player) Update(g *Game) {
	wasLanded := p.landed()
	p.gravUpdate(g, !p.hurt)

	if wasLanded && !p.landed() {
		// Walked off a platform - grant coyote time.
		p.coyoteTime = CoyoteTime
		p.fallTimer = 0
	}

	if p.top(g.assets) >= Height {
		p.reset(g)
	}

	// Enemy collisions - stomp or die.
	stompedAny := false
	for _, enemy := range p.getCollidingEnemies(g) {
		enemyRect := enemy.getRect(g.assets)
		factor := 0.2
		if p.velY > 0 {
			factor = 0.5
		}
		threshold := enemyRect.Top() + (enemyRect.Bottom()-enemyRect.Top())*factor
		if p.Y < threshold || p.stompedLastFrame {
			enemy.stomped(g)
			stompedAny = true
			p.velY = -6
			p.enemyStompedTimer = 3
			p.dashAllowed = true
		} else {
			p.hurt = true
			p.velY = -12
			p.fallState = FallFalling
			p.fallTimer = 0
			p.dashTimer = DashTimerTrailCutoff
			g.playSound("player_death", 1)
			g.animations = append(g.animations, NewAnimation(p.X, p.Y, "loselife_%d", 8, 4, AnchorCentre, 0, -1))
			break
		}
	}
	p.stompedLastFrame = stompedAny

	if p.landed() {
		p.dashAllowed = true
	}

	p.dashTimer--
	p.dashAnimationTimer++
	p.cachedJumpInputTimer--
	p.coyoteTime--
	p.wallJumpCoyoteTime--

	if p.dashTimer > DashTimerTrailCutoff && p.dashTimer%DashTrailInterval == 0 {
		g.animations = append(g.animations, NewDashTrail(p.X, p.Y, p.lastDashSprite))
	}

	dx := 0.0
	jumpPressed := p.controls.buttonPressed(0)

	jump := func() {
		p.velY = JumpVelY
		p.fallState = FallJumping
		p.coyoteTime = 0
		p.cachedJumpInputTimer = 0
		p.lowerGravityTimer = 5
		p.fallTimer = 0
		g.playSound("jump", 1)
	}
	wallJump := func(wallDirection float64) {
		p.velY = JumpVelY
		p.fallState = FallWallJumping
		p.velX = -wallDirection * WallJumpXVel
		p.facingX = -wallDirection
		p.grabbedWall = 0
		p.previousGrabbedWall = 0
		p.wallJumpCoyoteTime = 0
		p.cachedJumpInputTimer = 0
		p.fallTimer = 0
		g.playSound("jump", 1)
	}

	if p.hurt {
		p.gravityEnabled = true
		if p.top(g.assets) >= Height {
			p.hurt = false
		}
	} else if p.dashTimer > 0 {
		// For the first few frames of a dash the player doesn't move.
		if p.dashTimer < DashTime {
			if p.dashTimer%DashTrailInterval == 0 {
				g.animations = append(g.animations, NewDashTrail(p.X, p.Y, p.lastDashSprite))
			}
			// Vertical then horizontal component applied separately.
			p.move(g, 0, sign(p.velY), absf(p.velY))
			if p.move(g, sign(p.velX), 0, absf(p.velX)) && p.velY >= 0 {
				// Hit a wall while not travelling up - end the dash and grab.
				p.dashTimer = 0
				p.grabbedWall = p.facingX
			}
		}
	} else {
		dx = p.controls.getX()

		if p.grabbedWall != 0 {
			// Wall slide
			p.gravityEnabled = false
			if jumpPressed || p.cachedJumpInputTimer > 0 {
				wallJump(p.grabbedWall)
			} else if dx == -p.grabbedWall {
				p.previousGrabbedWall = p.grabbedWall
				p.wallJumpCoyoteTime = WallJumpCoyoteTime
				p.grabbedWall = 0
			} else {
				rect := p.getRectAt(g.assets, p.X+p.grabbedWall, p.Y)
				if p.move(g, 0, 1, 1) || !g.positionBlocked(rect) {
					p.grabbedWall = 0
				}
			}
		} else {
			if jumpPressed && p.wallJumpCoyoteTime > 0 {
				wallJump(p.previousGrabbedWall)
			} else {
				// Normal movement
				p.gravityEnabled = true
				if dx == 0 {
					p.velX = moveTowards(p.velX, 0, 1)
				} else {
					p.facingX = dx
					p.velX = moveTowards(p.velX, PlayerMaxXRunSpeed*dx, 1)
				}

				// Grab a wall if we hit one while moving down.
				if p.velX != 0 && p.move(g, sign(p.velX), 0, absf(p.velX)) && p.velY > 0 {
					p.grabbedWall = sign(p.velX)
					p.velX = 0
				}

				if (jumpPressed || p.cachedJumpInputTimer > 0) && (p.landed() || p.coyoteTime > 0) {
					jump()
				} else if jumpPressed && !p.landed() {
					p.cachedJumpInputTimer = CacheJumpInputTime
				} else if !p.landed() && p.velY < 0 && !p.controls.buttonDown(0) && p.dashTimer < -10 && p.enemyStompedTimer <= 0 {
					// Cut the jump short if the jump button is released while rising.
					p.velY = math.Min(p.velY+1, 0)
				}

				if p.dashAllowed && p.controls.buttonPressed(1) {
					dy := p.controls.getY()
					if dx != 0 || dy != 0 {
						length := math.Hypot(dx, dy)
						p.velX = float64(int(dx / length * DashSpeed))
						p.velY = float64(int(dy / length * DashSpeed))
						p.gravityEnabled = false
						p.dashAllowed = false
						p.dashTimer = DashTime + DashPauseTime
						p.dashAnimationTimer = 0
						p.fallState = FallFalling
						p.wallJumpCoyoteTime = 0
						g.playSound("jump_long", 5)
					}
				}
			}
		}
	}

	if sign(dx) != sign(p.velX) && p.dashTimer <= 0 {
		p.changeDirectionTimer = 5
	} else {
		p.changeDirectionTimer--
	}

	p.determineSprite(g, dx)

	if !p.landed() && p.dashTimer <= 0 {
		p.fallTimer++
	}

	p.replayData = append(p.replayData, ReplayFrame{X: p.X, Y: p.Y, Level: g.levelIndex, Sprite: p.Image})
}

func (p *Player) determineSprite(g *Game, dx float64) {
	p.Image = "blank"
	p.flame.Image = "blank"
	p.flame.anchor = AnchorFlame

	if p.hurt && g.timer%2 != 1 {
		return
	}

	dirIndex := "0"
	if p.facingX < 0 {
		dirIndex = "1"
	}

	switch {
	case p.hurt:
		frame := min(p.fallTimer/8, 5)
		p.Image = "die_" + strconv.Itoa(frame)

	case p.grabbedWall != 0 && p.velY >= 0:
		p.Image = "climb_" + dirIndex + "_1"
		p.flame.Image = "flame_climb_" + dirIndex + "_1"

	case !p.landed():
		switch {
		case p.fallState == FallJumping:
			frame := min(p.fallTimer/3, 5)
			p.Image = "jump_" + dirIndex + "_" + strconv.Itoa(frame)
			p.flame.Image = "flame_jump_" + dirIndex + "_" + strconv.Itoa(frame+1)
		case p.fallState == FallWallJumping:
			frame := min(p.fallTimer/8, 2)
			flameFrame := min(p.fallTimer/4, 6)
			p.Image = "wall_jump_" + dirIndex + "_" + strconv.Itoa(frame)
			p.flame.Image = "flame_wall_jump_" + dirIndex + "_" + strconv.Itoa(flameFrame)
		case p.dashTimer > 0:
			if p.dashAnimationTimer < 4 {
				flameFrame := p.dashAnimationTimer / 2
				p.Image = "dash_start_" + dirIndex
				p.lastDashSprite = p.Image
				p.flame.Image = "flame_dash_start_" + dirIndex + "_" + strconv.Itoa(flameFrame)
				p.flame.anchor = AnchorFlame
			} else {
				timer := p.dashAnimationTimer - 4
				frame := min(timer/3, 2)
				flameFrame := min(timer/3, 7)
				sprite := "dash_"
				if p.velY < 0 {
					sprite += "up_"
				} else if p.velY > 0 {
					sprite += "down_"
				}
				if p.velX != 0 {
					sprite += "horizontal_"
				}
				p.Image = fmt.Sprintf("%s%s_%d", sprite, dirIndex, frame)
				p.lastDashSprite = p.Image
				p.flame.Image = fmt.Sprintf("flame_%s%s_%d", sprite, dirIndex, flameFrame)
				p.flame.anchor = AnchorFlameDash
			}
		default:
			frame := min(p.fallTimer/8, 1)
			p.Image = "fall_" + dirIndex + "_" + strconv.Itoa(frame)
			p.flame.Image = "flame_wall_jump_" + dirIndex + "_" + strconv.Itoa(frame+4)
		}

	case dx == 0:
		p.Image = "stand_front"
		p.flame.Image = "flame_stand_" + strconv.Itoa((g.timer/4)%8)

	case p.changeDirectionTimer > 0:
		p.Image = "change_dir_" + dirIndex + "_0"
		p.flame.Image = "flame_change_dir_" + dirIndex + "_" + strconv.Itoa((g.timer/4)%3)

	default:
		frame := (g.timer / 4) % 8
		p.Image = "run_" + dirIndex + "_" + strconv.Itoa(frame)
		p.flame.Image = "flame_run_" + dirIndex + "_" + strconv.Itoa((g.timer/4)%8)
	}
}

func (p *Player) Draw(g *Game) {
	p.drawImage(g.assets)
	p.flame.X, p.flame.Y = p.X, p.Y
	p.flame.drawImage(g.assets)
}
