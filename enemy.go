package main

import "strconv"

// Enemy is a gravity-affected (or flying) foe. Four types per biome.
type Enemy struct {
	GravityActor
	directionX            float64
	directionY            float64
	etype                 int
	biome                 Biome
	health                int
	speed                 float64
	useDirectionalSprites bool
	dying                 bool
	stompedTimer          int
}

func NewEnemy(x, y float64, etype int, biome Biome, directionX float64, appearanceCount int) *Enemy {
	flying := enemyTypesFlying[biome][etype]
	e := &Enemy{
		directionX: directionX,
		etype:      etype,
		biome:      biome,
		health:     enemyTypesHealth[etype],
		speed:      float64(enemyTypesSpeed[etype]),
	}
	e.Actor = newActor("blank", x, y, enemyTypesAnchorPoints[biome][etype])
	e.gravityEnabled = !flying
	e.fallState = FallFalling
	e.self = e

	// Flying enemies on their third appearance move diagonally.
	if appearanceCount >= 3 && flying {
		e.directionY = 1
	}
	e.useDirectionalSprites = (biome == BiomeCastle && etype >= 2) || (biome == BiomeForest && etype < 2)
	return e
}

func (e *Enemy) collidableWidth(as *Assets) float64 {
	return enemyTypesWidthOverrides[e.biome][e.etype]
}
func (e *Enemy) collidableHeight(as *Assets) float64 {
	return enemyTypesHeightOverrides[e.biome][e.etype]
}

func (e *Enemy) Update(g *Game) {
	e.gravUpdate(g, !e.dying)

	if !e.dying {
		e.stompedTimer--

		// Don't move horizontally while falling (flying enemies always "fall" per
		// GravityActor, but should still move).
		if !e.gravityEnabled || e.fallState != FallFalling {
			if e.move(g, e.directionX, 0, e.speed) {
				e.directionX = -e.directionX
			}
			if e.directionY != 0 && e.move(g, 0, e.directionY, e.speed) {
				e.directionY = -e.directionY
			}
		}
	}

	image := enemySpriteNames[e.biome][e.etype]
	if e.useDirectionalSprites {
		dirIdx := "0"
		if e.directionX > 0 {
			dirIdx = "1"
		}
		image += "_" + dirIdx
	}
	image += "_" + strconv.Itoa((g.timer/4)%8)
	if e.stompedTimer > 0 || e.dying {
		image += "_hit"
	}
	e.Image = image
}

func (e *Enemy) stomped(g *Game) {
	// Don't lose health repeatedly if stomped on consecutive frames.
	if e.stompedTimer <= 0 {
		e.health--
		if e.health <= 0 {
			e.destroy(g)
			g.playSound("enemy_death", 5)
		} else {
			g.playSound("enemy_take_damage", 5)
		}
	}
	e.stompedTimer = 2
}

func (e *Enemy) destroy(g *Game) {
	e.dying = true
	e.gravityEnabled = true

	// Explosion first so it appears under the gain-time animation.
	explosionSprite := "air_explosion"
	if e.etype > 1 {
		explosionSprite = "explosion"
	}
	g.animations = append(g.animations, NewAnimation(e.X, e.Y, explosionSprite+"_%d", 12, 4, AnchorCentreBottom, 0, -1))

	g.gainTime(StompEnemyTimeBonus, e.centerx(g.assets), e.centery(g.assets))
}

func (e *Enemy) Draw(g *Game) { e.drawImage(g.assets) }
