package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// ReplayFrame is one recorded frame of the player: position, level and sprite.
type ReplayFrame struct {
	X, Y   float64
	Level  int
	Sprite string
}

// saveReplays writes replays to a text file: one replay per line, frames
// separated by ';', fields by ','.
func saveReplays(path string, replays [][]ReplayFrame) {
	f, err := os.Create(path)
	if err != nil {
		fmt.Printf("Error while saving replays: %v\n", err)
		return
	}
	defer f.Close()
	w := bufio.NewWriter(f)
	for _, replay := range replays {
		parts := make([]string, len(replay))
		for i, e := range replay {
			parts[i] = fmt.Sprintf("%d,%d,%d,%s", int(e.X), int(e.Y), e.Level, e.Sprite)
		}
		w.WriteString(strings.Join(parts, ";"))
		w.WriteString("\n")
	}
	w.Flush()
}

// loadReplays reads replays and returns them plus the high score (the length of
// the longest replay, i.e. most frames survived).
func loadReplays(path string) ([][]ReplayFrame, int) {
	file, err := os.Open(path)
	if err != nil {
		return nil, 0
	}
	defer file.Close()

	var replays [][]ReplayFrame
	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 0, 64*1024), 16*1024*1024)
	for scanner.Scan() {
		line := strings.TrimRight(scanner.Text(), "\r\n")
		if line == "" {
			continue
		}
		var replay []ReplayFrame
		for _, entry := range strings.Split(line, ";") {
			fields := strings.Split(entry, ",")
			if len(fields) < 4 {
				continue
			}
			x, _ := strconv.ParseFloat(fields[0], 64)
			y, _ := strconv.ParseFloat(fields[1], 64)
			level, _ := strconv.Atoi(fields[2])
			replay = append(replay, ReplayFrame{X: x, Y: y, Level: level, Sprite: fields[3]})
		}
		replays = append(replays, replay)
	}

	highScore := 0
	for _, r := range replays {
		if len(r) > highScore {
			highScore = len(r)
		}
	}
	return replays, highScore
}
