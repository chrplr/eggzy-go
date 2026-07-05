package main

import (
	"fmt"
	"path"
	"strconv"
	"strings"

	"github.com/Zyko0/go-sdl3/sdl"
)

type Game struct {
	player       *Player
	ghostPlayers []*GhostPlayer

	timer           int
	timeRemaining   float64
	timePickupBonus float64
	gainedTimeTimer int

	levelIndex int
	levelText  string

	grid              [][]int
	collisionTiles    map[int]bool
	tilesetTexture    *sdl.Texture
	tilesetGridW      int
	backgroundImage   string
	backgroundYOffset float64
	biome             Biome

	blockRects []Rect
	doors      []*Door
	gems       []*Gem
	enemies    []*Enemy
	animations []*Animation
	exitOpen   bool

	assets     *Assets
	audio      *Audio
}

func NewGame(player *Player, replays [][]ReplayFrame, assets *Assets, audio *Audio) *Game {
	g := &Game{
		player: player,
		assets: assets,
		audio:  audio,
	}
	gemNewGame()
	for _, r := range replays {
		g.ghostPlayers = append(g.ghostPlayers, NewGhostPlayer(r))
	}
	g.timeRemaining = InitialTimeRemaining * 60
	g.timePickupBonus = InitialPickupTimeBonus
	g.levelIndex = InitialLevelCycle*len(LevelSequence) - 1
	g.nextLevel()
	return g
}

func (g *Game) nextLevel() {
	g.levelIndex++

	// Each time we loop back to the first level, reduce the pickup time bonus.
	if g.levelIndex != 0 && g.levelIndex%len(LevelSequence) == 0 {
		if g.timePickupBonus > 1 {
			g.timePickupBonus -= 1
		} else if g.timePickupBonus == 1 {
			g.timePickupBonus = 0.5
		}
	}

	g.blockRects = nil
	g.doors = nil
	g.gems = nil
	g.enemies = nil
	g.animations = nil
	g.levelText = ""

	levelFilename := LevelSequence[g.levelIndex%len(LevelSequence)]
	playerStart := g.loadLevel(levelFilename)

	g.exitOpen = false

	if g.player != nil {
		g.player.newLevel(playerStart)
	}

	g.generateBlockRects()

	if g.player != nil {
		g.player.reset(g)
	}

	g.playSound("new_wave", 1)
}

func (g *Game) loadLevel(filename string) [2]float64 {
	playerStart := [2]float64{0, 0}
	levelCycle := g.levelIndex / len(LevelSequence)

	m, err := loadTMX(path.Join("tilemaps", filename))
	if err != nil {
		return playerStart
	}

	g.backgroundImage = propValue(m.Properties, "Background", "")
	if v, err := strconv.Atoi(propValue(m.Properties, "Background Offset Y", "0")); err == nil {
		g.backgroundYOffset = float64(v)
	} else {
		g.backgroundYOffset = 0
	}

	biomeName := propValue(m.Properties, "biome", "")
	biome := BiomeForest
	if strings.EqualFold(biomeName, "castle") {
		biome = BiomeCastle
	}
	g.biome = biome

	g.levelText = "LEVEL " + strconv.Itoa(g.levelIndex+1)
	if tutorial, ok := propLookup(m.Properties, "TutorialText"); ok && g.player != nil && levelCycle == 0 && len(tutorial) > 0 {
		dashName := g.player.controls.buttonName("dash")
		jumpName := g.player.controls.buttonName("jump")
		g.levelText = strings.ReplaceAll(strings.ReplaceAll(tutorial, "{DASH}", dashName), "{JUMP}", jumpName)
	}

	// Tile grid: each ID is offset by -1 so empty tiles become -1.
	mapWidth, mapHeight := m.Layer.Width, m.Layer.Height
	tiles := parseCSV(m.Layer.Data)
	g.grid = make([][]int, mapHeight)
	for row := 0; row < mapHeight; row++ {
		g.grid[row] = make([]int, mapWidth)
		for col := 0; col < mapWidth; col++ {
			idx := row*mapWidth + col
			if idx < len(tiles) {
				g.grid[row][col] = tiles[idx] - 1
			} else {
				g.grid[row][col] = -1
			}
		}
	}

	// Object layer: player start, gems, enemies and doors.
	for _, obj := range m.ObjectGroup.Objects {
		ox := float64(int(obj.X))
		oy := float64(int(obj.Y))
		name := obj.Name
		n := len(name)
		switch {
		case name == "PlayerStart":
			playerStart = [2]float64{ox, oy}
		case name == "Gem":
			g.gems = append(g.gems, NewGem(ox, oy))
		case strings.Contains(name, "Enemy") && n >= 3:
			// e.g. "EnemyR21": [-3]=facing, [-2]=type, [-1]=cycle
			enemyLevelCycle := int(name[n-1] - '0')
			appearanceCount := (levelCycle - enemyLevelCycle) + 1
			if appearanceCount >= 1 {
				facing := 1.0
				if name[n-3] == 'L' {
					facing = -1
				}
				enemyType := int(name[n-2] - '0')
				g.enemies = append(g.enemies, NewEnemy(ox, oy, enemyType, biome, facing, appearanceCount))
			}
		case strings.Contains(name, "Door"):
			variant := propValue(obj.Properties, "Variant", "0")
			doorBiome := propValue(obj.Properties, "Biome", biomeName)
			entrance := strings.Contains(name, "Entrance")
			g.doors = append(g.doors, NewDoor(ox, oy, doorBiome, variant, entrance))
		}
	}

	// Tileset: which tiles are collidable, and the tileset image.
	tilesetPath := path.Join("tilemaps", m.Tileset.Source)
	g.collisionTiles = make(map[int]bool)
	if tsx, err := loadTSX(tilesetPath); err == nil {
		for _, t := range tsx.Tiles {
			g.collisionTiles[t.ID] = true
		}
		imgPath := path.Join("tilemaps", tsx.Image.Source)
		g.tilesetTexture = g.assets.TilesetTexture(imgPath)
		if tsx.Image.Width > 0 {
			g.tilesetGridW = tsx.Image.Width / GridBlockSize
		} else if g.tilesetTexture != nil {
			g.tilesetGridW = int(g.tilesetTexture.W) / GridBlockSize
		}
	}

	return playerStart
}

func (g *Game) generateBlockRects() {
	g.blockRects = nil
	var current *Rect
	add := func() {
		g.blockRects = append(g.blockRects, *current)
		current = nil
	}

	// Merge horizontal runs of collidable tiles.
	for gy := 0; gy < len(g.grid); gy++ {
		row := g.grid[gy]
		for gx := 0; gx < len(row); gx++ {
			if g.collisionTiles[row[gx]] {
				px := float64(gx * GridBlockSize)
				py := float64(gy * GridBlockSize)
				if current == nil {
					current = &Rect{px, py, GridBlockSize, GridBlockSize}
				} else {
					current.W += GridBlockSize
				}
			} else if current != nil {
				add()
			}
		}
		if current != nil {
			add()
		}
	}

	// Consolidate vertically: merge equal-width rects stacked directly below.
	anyFound := true
	for anyFound {
		anyFound = false
		for i := range g.blockRects {
			cur := g.blockRects[i]
			belowIdx := -1
			for j := range g.blockRects {
				b := g.blockRects[j]
				if b.X == cur.X && b.W == cur.W && b.Y == cur.Y+cur.H {
					belowIdx = j
					break
				}
			}
			if belowIdx != -1 {
				g.blockRects[i].H += g.blockRects[belowIdx].H
				g.blockRects = append(g.blockRects[:belowIdx], g.blockRects[belowIdx+1:]...)
				anyFound = true
				break
			}
		}
	}

	// Extend rects touching the top of the level upward, so you can't stand on
	// trees that poke off the top of the screen.
	for i := range g.blockRects {
		if g.blockRects[i].Y == 0 {
			h := g.blockRects[i].H
			g.blockRects[i].Y = LevelYBoundary
			g.blockRects[i].H = h + (-LevelYBoundary)
		}
	}
}

func (g *Game) Update() {
	g.timer++
	g.gainedTimeTimer--

	if g.timeRemaining > 0 {
		g.timeRemaining -= 1
	}

	// Update objects in the original's order.
	if g.player != nil {
		g.player.Update(g)
	}
	for _, d := range g.doors {
		d.Update(g)
	}
	for _, a := range g.animations {
		a.Update(g)
	}
	for _, gm := range g.gems {
		gm.Update(g)
	}
	for _, e := range g.enemies {
		e.Update(g)
	}
	for _, gp := range g.ghostPlayers {
		gp.Update(g)
	}

	// Remove expired enemies, animations and collected gems.
	var enemies []*Enemy
	for _, e := range g.enemies {
		if e.top(g.assets) < Height {
			enemies = append(enemies, e)
		}
	}
	g.enemies = enemies

	var anims []*Animation
	for _, a := range g.animations {
		if !a.finished() {
			anims = append(anims, a)
		}
	}
	g.animations = anims

	var gems []*Gem
	for _, gm := range g.gems {
		if !gm.collected {
			gems = append(gems, gm)
		}
	}
	g.gems = gems

	// Exit door / level completion.
	if g.player != nil {
		if g.exitOpen {
			if g.player.centerx(g.assets) >= Width {
				g.nextLevel()
			}
		} else if len(g.gems) == 0 {
			g.exitOpen = true
			for _, d := range g.doors {
				d.open()
			}
		}
	}
}

func (g *Game) Draw() {
	// Background
	g.assets.Blit(g.backgroundImage, 0, g.backgroundYOffset)

	// Level tiles
	for rowY := 0; rowY < len(g.grid); rowY++ {
		x := 0.0
		for _, tile := range g.grid[rowY] {
			if tile >= 0 && g.tilesetGridW > 0 {
				gx := tile % g.tilesetGridW
				gy := tile / g.tilesetGridW
				g.assets.BlitTile(g.tilesetTexture,
					float64(gx*GridBlockSize), float64(gy*GridBlockSize), GridBlockSize, GridBlockSize,
					x, float64(rowY*GridBlockSize))
			}
			x += GridBlockSize
		}
	}

	// Objects, in draw order
	for _, gp := range g.ghostPlayers {
		gp.Draw(g)
	}
	for _, d := range g.doors {
		d.Draw(g)
	}
	for _, a := range g.animations {
		a.Draw(g)
	}
	if g.player != nil {
		g.player.Draw(g)
	}
	for _, gm := range g.gems {
		gm.Draw(g)
	}
	for _, e := range g.enemies {
		e.Draw(g)
	}

	g.drawUI()
}

func (g *Game) drawUI() {
	// Level text bar
	g.assets.FillRect(0, 500, Width, 50, 0, 54, 255)
	g.assets.Blit("text_area_frame", 0, 500)
	g.assets.DrawText(g.levelText, Width/2, 508, AlignCentre, "font")

	// Time remaining
	g.assets.Blit("status_back", Width/2-297/2, 0)
	font := "font"
	if g.gainedTimeTimer >= 0 {
		font = "fontbr"
	}
	g.assets.DrawText(fmt.Sprintf("%.1f", g.timeRemaining/60), Width/2, 10, AlignCentre, font)
}

func (g *Game) gainTime(t, x, y float64) {
	g.timeRemaining += t * 60
	timeAddedID := "half"
	if t != 0.5 {
		timeAddedID = strconv.Itoa(int(t))
	}
	format := "timer_plus_" + timeAddedID + "_%d"
	g.animations = append(g.animations, NewAnimation(x, y, format, 14, 4, AnchorCentre, 5, 34))
	g.animations = append(g.animations, NewAnimation(x, y, "pickup_%d", 8, 4, AnchorCentre, 0, -1))
	g.gainedTimeTimer = 20
}

func (g *Game) positionBlocked(rect Rect) bool {
	for _, br := range g.blockRects {
		if rect.colliderect(br) {
			return true
		}
	}
	for _, d := range g.doors {
		if !d.isFullyOpen() && d.spriteRect(g.assets).colliderect(rect) {
			return true
		}
	}
	// Can't go off the left edge or above the vertical boundary (but the right
	// edge is allowed, so the player can exit through the door).
	if rect.Left() <= 0 || rect.Top() < LevelYBoundary {
		return true
	}
	return false
}

// playSound plays a game sound, but only when there is a player (not on the menu).
func (g *Game) playSound(name string, count int) {
	if g.player == nil {
		return
	}
	g.audio.PlaySound(name, count)
}
