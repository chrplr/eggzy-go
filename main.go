package main

import (
	"flag"
	"fmt"
	"sort"

	"github.com/chrplr/pgzgo"
)

type State int

const (
	StateTitle State = iota
	StateControls
	StatePlay
	StateGameOver
)

var (
	state              State
	game               *Game
	highScore          int
	gameOverStateTimer int
	allReplays         [][]ReplayFrame
	totalFrames        int

	assets     *Assets
	audio      *Audio
	replayPath string
)

func update() {
	totalFrames++

	controls := Controls{}
	jumpPressed := controls.buttonPressed(0)

	switch state {
	case StateTitle:
		if jumpPressed {
			state = StateControls
		}

	case StateControls:
		if jumpPressed {
			state = StatePlay
			game = NewGame(NewPlayer(controls), allReplays, assets, audio)
			audio.PlayMusic("ingame_theme", 0.2)
		}

	case StatePlay:
		if game.timeRemaining <= 0 {
			game.playSound("gameover", 1)
			state = StateGameOver
			gameOverStateTimer = 0

			allReplays = append(allReplays, game.player.replayData)
			if len(allReplays) > MaxReplays {
				sort.SliceStable(allReplays, func(i, j int) bool {
					return len(allReplays[i]) > len(allReplays[j])
				})
				allReplays = allReplays[:MaxReplays]
			}
			saveReplays(replayPath, allReplays)
		} else {
			game.Update()
		}

	case StateGameOver:
		gameOverStateTimer++
		if gameOverStateTimer > 60 && jumpPressed {
			if game.timer > highScore {
				highScore = game.timer
			}
			state = StateTitle
			audio.PlayMusic("title_theme", 0.3)
		}
	}
}

func draw() {
	switch state {
	case StateTitle:
		assets.Blit("title", 0, 0)
		assets.Blit("press_to_start", 0, 0)
		animFrame := (totalFrames / 6) % 11
		assets.Blit("start"+itoa(animFrame), Width/2-150, 360)

	case StateControls:
		assets.Fill(0, 0, 0)
		assets.Blit("controls", 0, 0)

	case StatePlay:
		game.Draw()

	case StateGameOver:
		assets.Fill(0, 54, 255)
		animFrame := (totalFrames / 5) % 14
		assets.Blit("gameover"+itoa(animFrame), Width/2-625/2, 100)

		seconds := game.timer / 60
		if seconds >= 60 {
			assets.Blit("survived_for_mins_seconds", 0, 0)
			assets.DrawText(itoa(seconds/60), 180, 270, AlignRight, "fontlrg")
			assets.DrawText(itoa(seconds%60), 470, 270, AlignCentre, "fontlrg")
		} else {
			assets.Blit("survived_for_seconds", 0, 0)
			assets.DrawText(itoa(seconds), 300, 310, AlignRight, "fontlrg")
		}

		if game.timer > highScore {
			nr := (totalFrames / 5) % 8
			assets.Blit("newrecord"+itoa(nr), Width/2-575/2, 380)
		}
	}
}

func itoa(i int) string { return fmt.Sprintf("%d", i) }

func main() {
	replays := flag.String("replays", "eggzy-replays.txt", "path to the replay save file")
	selftest := flag.Bool("selftest", false, "load every level and step the physics, then exit (headless test)")
	flag.Parse()

	replayPath = *replays

	a, err := pgzgo.New(pgzgo.Config{
		Title:  "Eggzy",
		Width:  Width,
		Height: Height,
		Images: imagesFS,
		Audio:  audioFS,
	})
	if err != nil {
		panic(err)
	}
	defer a.Close()

	app = a
	assets = &Assets{Screen: a.Screen}
	audio = a.Audio

	allReplays, highScore = loadReplays(replayPath)

	if *selftest {
		g := NewGame(NewPlayer(Controls{}), nil, assets, audio)
		for lvl := 0; lvl < len(LevelSequence); lvl++ {
			for i := 0; i < 120; i++ {
				g.Update()
			}
			fmt.Printf("level %d (%s): %dx%d grid, %d blockRects, %d gems, %d enemies, %d doors\n",
				lvl, LevelSequence[lvl], len(g.grid), func() int {
					if len(g.grid) > 0 {
						return len(g.grid[0])
					}
					return 0
				}(), len(g.blockRects), len(g.gems), len(g.enemies), len(g.doors))
			g.nextLevel()
		}
		fmt.Println("SELFTEST OK")
		return
	}

	state = StateTitle
	audio.PlayMusic("title_theme", 0.3)

	a.Loop(
		func(*pgzgo.App) { update() },
		func(*pgzgo.App) { draw() },
	)
}
