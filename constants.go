package main

const (
	Width  = 825
	Height = 550

	GridBlockSize  = 25
	LevelYBoundary = -100

	InitialLevelCycle = 0

	InitialTimeRemaining   = 15
	InitialPickupTimeBonus = 2.0
	StompEnemyTimeBonus    = 3.0

	// Player movement
	CoyoteTime           = 6
	JumpVelY             = -10
	WallJumpXVel         = 8
	WallJumpCoyoteTime   = 15
	CacheJumpInputTime   = 5
	PlayerWidth          = 20
	PlayerHeight         = 40
	PlayerMaxXRunSpeed   = 5
	DashTime             = 18
	DashSpeed            = 10
	DashPauseTime        = 5
	DashTrailInterval    = 3
	DashTimerTrailCutoff = -10

	GravityMaxFallSpeed = 7

	MaxReplays     = 10
	ReplayFilename = "eggzy-replays"

	targetFPS   = 60
	frameMillis = 1000 / targetFPS
)

// LevelSequence is the order in which the Tiled maps are played.
var LevelSequence = []string{
	"starter1.tmx", "starter2.tmx", "starter3.tmx", "starter4.tmx",
	"forest1.tmx", "forest2.tmx", "forest3.tmx", "forest4.tmx", "forest9.tmx",
	"castle1.tmx", "castle2.tmx", "castle3.tmx", "castle4.tmx",
	"castle5.tmx", "castle6.tmx", "castle7.tmx", "castle8.tmx",
	"forest5.tmx", "forest6.tmx", "forest7.tmx", "forest8.tmx",
}

// Biome selects which enemy/door variants a level uses.
type Biome int

const (
	BiomeForest Biome = 0
	BiomeCastle Biome = 1
)

// Enemy tables, indexed [biome][type]. There are four enemy types per biome.
var (
	enemySpriteNames = [2][4]string{
		BiomeForest: {"fly", "mghost", "triffid", "bigbloom"},
		BiomeCastle: {"robot0", "robot1", "robot2", "robot3"},
	}
	enemyTypesFlying = [2][4]bool{
		BiomeForest: {true, true, false, true},
		BiomeCastle: {true, true, false, false},
	}
	enemyTypesWidthOverrides = [2][4]float64{
		BiomeForest: {30, 50, 50, 50},
		BiomeCastle: {30, 50, 48, 50},
	}
	enemyTypesHeightOverrides = [2][4]float64{
		BiomeForest: {30, 65, 70, 90},
		BiomeCastle: {40, 40, 60, 120},
	}
	enemyTypesAnchorPoints = [2][4]Anchor{
		BiomeForest: {anchorAbsY(60), AnchorCentreBottom, AnchorCentreBottom, AnchorCentreBottom},
		BiomeCastle: {anchorAbsY(40), anchorAbsY(40), anchorAbsY(95), AnchorCentreBottom},
	}
	enemyTypesHealth = [4]int{1, 3, 1, 3}
	enemyTypesSpeed  = [4]int{2, 1, 2, 1}
)

// SpecialFontSymbols substitute for controller button glyphs in text.
var specialFontSymbols = map[string]rune{"xb_a": '%', "xb_b": '#'}
