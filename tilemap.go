package main

import (
	"embed"
	"encoding/xml"
	"strconv"
	"strings"
)

// tilemapsFS embeds the Tiled maps (.tmx), tilesets (.tsx) and their tileset
// PNGs into the binary. All internal references are same-directory filenames,
// so the whole tilemaps/ tree is self-contained.
//
//go:embed tilemaps
var tilemapsFS embed.FS

// --- Tiled TMX (map) structures ---

type tmxMap struct {
	Width       int            `xml:"width,attr"`
	Height      int            `xml:"height,attr"`
	Properties  []tmxProperty  `xml:"properties>property"`
	Tileset     tmxTilesetRef  `xml:"tileset"`
	Layer       tmxLayer       `xml:"layer"`
	ObjectGroup tmxObjectGroup `xml:"objectgroup"`
}

type tmxProperty struct {
	Name  string `xml:"name,attr"`
	Value string `xml:"value,attr"`
}

type tmxTilesetRef struct {
	Source string `xml:"source,attr"`
}

type tmxLayer struct {
	Width  int    `xml:"width,attr"`
	Height int    `xml:"height,attr"`
	Data   string `xml:"data"`
}

type tmxObjectGroup struct {
	Objects []tmxObject `xml:"object"`
}

type tmxObject struct {
	Name       string        `xml:"name,attr"`
	X          float64       `xml:"x,attr"`
	Y          float64       `xml:"y,attr"`
	Properties []tmxProperty `xml:"properties>property"`
}

// --- Tiled TSX (tileset) structures ---

type tsxTileset struct {
	Columns int       `xml:"columns,attr"`
	Image   tsxImage  `xml:"image"`
	Tiles   []tsxTile `xml:"tile"`
}

type tsxImage struct {
	Source string `xml:"source,attr"`
	Width  int    `xml:"width,attr"`
	Height int    `xml:"height,attr"`
}

type tsxTile struct {
	ID int `xml:"id,attr"`
}

func loadTMX(path string) (*tmxMap, error) {
	data, err := tilemapsFS.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var m tmxMap
	if err := xml.Unmarshal(data, &m); err != nil {
		return nil, err
	}
	return &m, nil
}

func loadTSX(path string) (*tsxTileset, error) {
	data, err := tilemapsFS.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var t tsxTileset
	if err := xml.Unmarshal(data, &t); err != nil {
		return nil, err
	}
	return &t, nil
}

// propValue returns the value of a named property, or def if absent.
func propValue(props []tmxProperty, name, def string) string {
	for _, p := range props {
		if p.Name == name {
			return p.Value
		}
	}
	return def
}

// propLookup returns a property's value and whether it existed.
func propLookup(props []tmxProperty, name string) (string, bool) {
	for _, p := range props {
		if p.Name == name {
			return p.Value, true
		}
	}
	return "", false
}

// parseCSV parses a Tiled CSV tile-data blob into a flat list of tile IDs.
func parseCSV(s string) []int {
	var out []int
	for _, part := range strings.Split(s, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		if n, err := strconv.Atoi(part); err == nil {
			out = append(out, n)
		}
	}
	return out
}
