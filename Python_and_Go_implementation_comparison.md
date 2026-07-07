# Eggzy — Python vs. Go implementation comparison

This document analyses how the Go port in this folder relates to the original
`eggzy.py`. It covers the structural mapping, the language‑paradigm differences
that shaped the port, the framework substitutions, and a set of subtle
numeric/semantic details that had to be reproduced exactly for the game to
behave the same way.

The goal throughout the port was **behavioural fidelity**: the Go code is a
faithful translation of the game logic, deviating only where a language or
library difference forces a different expression of the same idea, or where a
platform feature (game controllers, a writable home folder) is out of scope for
the port.

The single Python file (`eggzy.py`, ~1,600 lines) is split into ~18 focused Go
files (~2,100 lines total). The extra volume is almost entirely language-level boilerplate that
Python gets for free: explicit struct/interface declarations and XML-binding tags.
The sprite-font/asset/loop plumbing itself now lives in the pgzgo harness, not per game.

---

## 1. High‑level architecture

Both versions share the same conceptual design:

- A **tile‑based platformer**. Levels are Tiled `.tmx` maps; a companion `.tsx`
  tileset declares which tile IDs are solid. The player collects every gem in a
  level to open the exit door, then walks off the right edge to advance — all
  against a countdown timer that gems and stomped enemies top up.
- An **actor hierarchy** — `Actor → CollideActor → GravityActor → Player/Enemy`
  — where each layer adds behaviour (anchored drawing → pixel‑stepped wall
  collision → gravity/jumping).
- A **title → controls → play → game‑over** state machine.
- A **ghost‑replay system**: every play‑through records the player's per‑frame
  position/level/sprite; up to 10 are persisted and replayed as translucent
  "ghost" runners on future games.

The two largest pieces of logic in both — `Player.update`/`(*Player).Update` (the
full movement state machine) and `Game.load_level`/`(*Game).loadLevel` +
`generate_block_rects` — are ported statement‑for‑statement.

### File layout

| Concern | Python | Go |
|---|---|---|
| Constants / tables | top of `eggzy.py` | `constants.go` |
| Anchors, Actor, Collide/Gravity actors | `eggzy.py` | `actor.go` |
| Player | `Player` | `player.go` |
| Enemy | `Enemy` | `enemy.go` |
| Gem / Door / Animation | those classes | `gem.go`, `door.go`, `animation.go` |
| Ghost replay | `GhostPlayer` | `ghost.go` |
| Replay save/load | `save_replays`/`load_replays` | `replay.go` |
| Game (levels, physics glue, UI) | `Game` | `game.go` |
| Tiled TMX/TSX parsing | inline `ET.parse` | `tilemap.go` |
| Input | `Controls`/`KeyboardControls` | `input.go` |
| Assets / sprite font | Pygame Zero `images`/`draw_text` | `assets.go` |
| Audio | Pygame Zero `sounds`/`music` | pgzgo `Audio` |
| State machine / entry point | `update`/`draw`/module code | `main.go` |

---

## 2. Language paradigm: classes/inheritance → interfaces/embedding

This is the biggest structural difference.

### Python: classical inheritance

```python
class CollideActor(Actor): ...
class GravityActor(CollideActor):
    def update(self, detect=True): ...
    def get_collidable_width(self):  # overridden by Player/Enemy
        return getattr(images, self.image).get_width()
class Player(GravityActor):
    def get_collidable_width(self): return PLAYER_WIDTH
class Enemy(GravityActor):
    def get_collidable_width(self):
        return ENEMY_TYPES_WIDTH_OVERRIDES[self.biome][self.type]
```

`CollideActor.get_rect` calls `self.get_collidable_width()` and Python's dynamic
dispatch routes it to the subclass override — the base class never needs to know
which concrete type it is.

### Go: struct embedding + a `self` interface back‑reference

Go has no inheritance, so the port uses **struct embedding** for code reuse and
a small **interface** for the one place the base class must call back into the
concrete type:

```go
type collidable interface {
    collidableWidth(as *Assets) float64
    collidableHeight(as *Assets) float64
}

type CollideActor struct {
    Actor
    self collidable   // back-reference to the concrete Player/Enemy
}

func (c *CollideActor) getRectAt(as *Assets, cx, cy float64) Rect {
    w := c.self.collidableWidth(as)   // dispatches to Player/Enemy
    h := c.self.collidableHeight(as)
    return Rect{cx - math.Floor(w/2), cy - h, w, h}
}
```

Because embedding is not inheritance, the concrete constructor must wire the
back‑reference explicitly — `p.self = p` in `NewPlayer`, `e.self = e` in
`NewEnemy`. This `self`‑field idiom is the same one used in the other ports in
this repo (e.g. the cars in Leading Edge) and recurs anywhere the base type must
invoke an overridden method.

Method resolution that Python does implicitly (`super().update(...)`) becomes an
**explicit call to the embedded method**: `p.gravUpdate(g, !p.hurt)` in
`(*Player).Update` is the transliteration of `super().update(not self.hurt)`.

### Polymorphic update/draw lists

Python iterates a heterogeneous list polymorphically:

```python
for obj in [self.player] + self.doors + self.animations + self.gems + self.enemies + self.ghost_players:
    if obj: obj.update()
```

Go has no covariant "list of anything with `.Update()`" without an interface, and
here each slice is already a concrete type, so the port simply unrolls the
concatenation into typed loops in the original order:

```go
if g.player != nil { g.player.Update(g) }
for _, d := range g.doors      { d.Update(g) }
for _, a := range g.animations { a.Update(g) }
for _, gm := range g.gems      { gm.Update(g) }
for _, e := range g.enemies    { e.Update(g) }
for _, gp := range g.ghostPlayers { gp.Update(g) }
```

The `Draw` order (`ghosts, doors, animations, player, gems, enemies`) is likewise
unrolled verbatim — the draw order matters (player behind gems/enemies), so it is
preserved exactly.

---

## 3. The `game` global → an explicit `*Game` parameter

Python reaches a module‑level `game` global from inside every actor method
(`game.player`, `game.play_sound(...)`, `game.animations.append(...)`,
`game.position_blocked(...)`). The Go port has no such global; instead **`g *Game`
is threaded as a parameter** through every method that needs it:
`(*Player).Update(g)`, `(*Enemy).stomped(g)`, `(*Gem).Update(g)`, and so on.

This is a purely mechanical but pervasive change — nearly every actor method
signature gains a `g *Game`. It makes the data‑flow explicit and avoids the
initialization‑order hazard the Python code has to guard against (`if game is not
None` inside `Player.reset`, because the player is constructed *during* `Game`
construction before the global is assigned). In Go, `reset` simply receives the
fully‑constructed `g`, so the nil‑check disappears.

---

## 4. Framework: Pygame Zero → pgzgo (on go-sdl3)

pgzgo supplies this machinery; the game adds only the pieces specific to it.

| Pygame Zero feature | pgzgo equivalent (over go-sdl3) |
|---|---|
| `Actor("name", …)` auto-loads a PNG | `Screen.Texture` — pgzgo's lazily-cached texture |
| `screen.blit(name, (x,y))` | `Screen.Blit` / `BlitCentred` |
| tileset sub-blit | `Screen.BlitTile` (pgzgo) |
| Anchor tuples resolved internally | `Anchor` struct + `offset(w,h)` (§6) |
| `keyboard.left`, `keyboard.space` | `app.Keyboard.Held(sc)` snapshot; edge latch in `input.go` |
| `sounds.foo.play()` via `getattr` | `Audio.PlaySound(name, count)` |
| `music.play`/`set_volume` | `Audio.PlayMusic` |
| `ET.parse(...)` (ElementTree) | `encoding/xml` structs in `tilemap.go` (§5) |
| the `update()`/`draw()` loop | `app.Loop(update, draw)` — pgzgo's fixed-step, FPS-capped loop |
| sprite font via `draw_text` | its own `Assets.DrawText`/`charImageAndWidth` over `Screen` (§7) |

### The game loop and frame timing

pgzgo's `app.Loop` runs the fixed-step, FPS-capped loop, calling `update` then `draw` each tick.
Both games are frame‑count driven (`game.timer` / `g.timer` increments once per
update and animation frames are `timer // interval`), so the timing model matches
as long as the loop runs at 60 Hz. The Python `DEBUG_SLOWMO` frame‑skip is not
ported (it was a debug aid).

---

## 5. Tiled map parsing: ElementTree → encoding/xml

Python walks the XML DOM imperatively with XPath‑ish `find` calls:

```python
map_root = ET.parse(os.path.join(path, filename)).getroot()
self.background_image = properties_node.find("./property[@name='Background']").attrib["value"]
map_data = layer_node.find("data").text.split(",")
```

Go binds the same documents into typed structs with `xml` tags and unmarshals
once (`tilemap.go`):

```go
type tmxMap struct {
    Width       int            `xml:"width,attr"`
    Properties  []tmxProperty  `xml:"properties>property"`
    Tileset     tmxTilesetRef  `xml:"tileset"`
    Layer       tmxLayer       `xml:"layer"`
    ObjectGroup tmxObjectGroup `xml:"objectgroup"`
}
```

Because Go can't express XPath attribute predicates (`[@name='Background']`), the
named‑property lookup becomes a helper: `propValue(props, "Background", def)` and
`propLookup(props, name) (string, bool)` — the latter mirroring Python's
"is the node present?" check for optional properties like `TutorialText` and
`Background Offset Y`.

The rest of `loadLevel` is a faithful port, including the fiddly details:

- **Tile‑ID offset**: every CSV tile ID is decremented by 1 so empty tiles
  become `-1` (`tiles[idx] - 1`), exactly as Python does.
- **Object name parsing**: enemy objects are named like `EnemyR21` where the
  last three characters encode facing (`R`/`L`), type (`0–3`) and the level cycle
  on which the enemy first appears. Python slices with negative indices
  (`object_name[-1]`, `[-2]`, `[-3]`); Go indexes from `len(name)`
  (`name[n-1]`, `name[n-2]`, `name[n-3]`) and converts ASCII digits with
  `int(name[n-1] - '0')`.
- **Position truncation**: Tiled writes floats; both versions truncate toward
  zero — `int(float(...))` in Python, `float64(int(obj.X))` in Go.

### Block‑rect generation

`generate_block_rects` is the trickiest algorithm in the file, and it is ported
line‑for‑line: (1) merge horizontal runs of solid tiles into rectangles;
(2) repeatedly merge any rectangle with an equal‑width rectangle directly below
it; (3) extend any rectangle touching the top of the level upward to
`LEVEL_Y_BOUNDARY` so the player can't stand on scenery poking off the top of the
screen. The only representational change: Python mutates `Rect` objects in a list
and uses `list.remove`; Go operates on a `[]Rect` by index and deletes with the
`append(s[:i], s[i+1:]...)` slice trick. The nested‑loop "restart after every
merge" structure (Python's `while any_found` + `break`) is preserved so the
output is identical.

---

## 6. The anchor system

Pygame Zero anchors are heterogeneous tuples: `("center", "center")`,
`("center", "bottom")`, `("center", 60)`, `(0, 0)`. Go has no such union type, so
`actor.go` models it with a small struct that tags each axis as *centre*,
*bottom*, or an *absolute pixel offset*:

```go
type Anchor struct {
    xKind, yKind int      // akCenter | akBottom | akAbs
    xVal, yVal   float64
}
func anchorAbsY(v float64) Anchor { return Anchor{akCenter, akAbs, 0, v} }
```

So the Python enemy anchor table

```python
ENEMY_TYPES_ANCHOR_POINTS = {Biome.CASTLE: [("center", 40),("center", 40),("center", 95),("center", "bottom")], ...}
```

becomes

```go
BiomeCastle: {anchorAbsY(40), anchorAbsY(40), anchorAbsY(95), AnchorCentreBottom},
```

`Anchor.offset(w, h)` computes the pixel offset from `(X, Y)` to the sprite's
top‑left, reproducing Pygame Zero's resolution rules. This lets the actor drawing
code and the collision code agree on where the sprite sits — important because
the player's *feet* are at anchor `("center", 60)`, not at the sprite bottom.

---

## 7. Sprite font

Both games render text from per‑glyph PNGs (`font065.png` for `'A'`, etc.).
Python:

```python
def get_char_image_and_width(char, font):
    if char == " ": return None, 22
    if char in SPECIAL_FONT_SYMBOLS_INVERSE:
        image = getattr(images, SPECIAL_FONT_SYMBOLS_INVERSE[char])
    else:
        image = getattr(images, f"{font}{ord(char):03d}")
    return image, image.get_width()
```

Go's `Assets.charImageAndWidth` reproduces this: space → advance 22 with no
glyph; the two controller‑button placeholder symbols `'%'`/`'#'` map to the
`xb_a`/`xb_b` sprites; everything else formats as `fmt.Sprintf("%s%03d", font,
int(char))`. `DrawText` applies the same centre/right alignment arithmetic,
including Python's integer‑division centring (`text_width // 2`), which the Go
port matches with `int(a.TextWidth(...)) / 2` rather than a float divide, to keep
pixel positions identical.

---

## 8. Audio and the "sound with count" idiom

Python's `play_sound(name, count)` uses `getattr` to pick a random numbered
variant:

```python
sound = getattr(sounds, name + str(randint(0, count - 1)))
sound.play()
```

Go's `Audio.PlaySound(name, count)` does the same explicitly — pick
`randIntn(count)`, build `name + index`, load/cache and play. Both wrap the call
so a missing file or absent sound hardware is non‑fatal (Python's broad
`except`; Go's nil‑texture/err‑ignoring loads). `Game.play_sound` also keeps the
"don't play sounds when there is no player (menu / self‑test)" guard —
`if g.player == nil { return }`.

---

## 9. Numeric and semantic details reproduced exactly

Platformer feel is extremely sensitive to integer/float behaviour, so several
details were transliterated with care:

- **Gravity cadence**: `if game.timer % (3 if lower_gravity_timer > 0 else 2) == 0`
  is preserved — velocity only increments on every 2nd (or 3rd) frame, giving the
  characteristic floaty arc. Go: `mod := 2; if lowerGravityTimer > 0 { mod = 3 }`.
- **Dash vector**: Python builds `pygame.math.Vector2(dx, dy).normalize() *
  DASH_SPEED` then **truncates each component to int**
  (`self.vel_x = int(v.x)`). Go computes `math.Hypot(dx, dy)` and applies the same
  truncation: `float64(int(dx / length * DashSpeed))`. The `int()` truncation is
  load‑bearing — it makes diagonal dashes slightly slower than `DASH_SPEED`, and
  the port keeps it.
- **Stomp threshold**: the "top 20% (or 50% when moving down) of the enemy rect"
  hit test is ported with the same 0.2/0.5 factor and the `stomped_last_frame`
  carry‑over that prevents an immediate re‑hit after bouncing off a head.
- **Jump‑cut rule**: the exact compound condition that bleeds off upward velocity
  when the jump button is released — `not landed and vel_y < 0 and not
  button_down(0) and dash_timer < -10 and enemy_stomped_timer <= 0` — is copied
  verbatim (note the literal `-10`, distinct from the `DashTimerTrailCutoff`
  constant that also happens to be `-10`).
- **`get_rect` centring**: `centre_x - (w // 2)` uses Python floor division; Go
  uses `math.Floor(w/2)` to match for the (even) collision widths in play.
- **`move` is pixel‑stepped**: both step one pixel at a time up to `speed`,
  returning `true` on the first blocked step, so an actor never tunnels into a
  wall. `int(speed)` truncation of the step count is identical in both.

---

## 10. Faithfully preserved Python quirks

Two genuine oddities in the original are reproduced deliberately, with comments,
rather than "fixed" — changing them would change behaviour:

1. **`enemy_stomped_timer` is never decremented.** Python sets it to 3 on a stomp
   and reads it in the jump‑cut condition, but nothing ever counts it back down;
   the Go port likewise sets `p.enemyStompedTimer = 3` and never decrements it.
2. **The `self.flame_image` typo.** In the `hurt` branch of `determine_sprite`,
   Python writes `self.flame_image = "blank"` — assigning a *new* attribute
   rather than `self.flame.image`. The intended flame sprite is therefore never
   cleared through that path (it stays whatever `self.flame.image = "blank"` at
   the top of the method already set it to). The Go port mirrors the effect: in
   the `hurt` case it sets only `p.Image` and leaves the already‑blanked
   `p.flame.Image` alone.

---

## 11. Intentional differences (out of scope for the port)

A few Python features are deliberately not carried over, because they concern the
host platform rather than the game:

- **Game controllers.** Python has `JoystickControls` and hot‑plug detection; the
  Go port ships only `Controls` (keyboard: arrows, `SPACE` = jump, `Z` = dash).
  `buttonName` still returns the keyboard labels used to fill in `{JUMP}`/`{DASH}`
  in tutorial text.
- **Save‑folder discovery.** Python's `get_save_folder` writes to `~/.code-the-
  classics-vol-2` when run from the Raspberry Pi pre‑installed home directory.
  The Go port writes the replay file to a path given by a `-replays` flag
  (default `eggzy-replays.txt`), avoiding a pre‑existing `eggzy-replays/`
  directory next to the assets. The on‑disk **format is identical** (one replay
  per line; frames `;`‑separated; `x,y,level,sprite` fields `,`‑separated), so the
  files are interchangeable.
- **Debug switches** (`DEBUG_SHOW_*`, `DEBUG_MOVEMENT`, `DEBUG_SLOWMO`) are
  dropped.
- **Version checks** for Python/Pygame‑Zero at startup have no Go analogue.
- **A `-selftest` flag** is *added* in Go: it loads all 21 levels headlessly,
  steps the physics, and prints per‑level grid/rect/object counts. This exists
  only to verify the port without a display and has no Python counterpart.

---

## 12. Summary

The port is a close, behaviour‑preserving translation. The substantive rewrites
are all forced by the language or framework:

- inheritance → **embedding + a `self` interface** for the one virtual call;
- the `game` global → an **explicit `*Game` parameter** everywhere;
- Pygame Zero's implicit asset/sound/font/loop machinery → the **pgzgo harness**
  (a thin `assets.go` adds only tileset loading + the sprite font);
- ElementTree XPath → **`encoding/xml` structs + property‑lookup helpers**;
- anchor tuples → a small **`Anchor` struct**.

Everything that affects how the game *plays* — the movement state machine, the
gravity cadence, the dash math, the stomp thresholds, the block‑rect
consolidation, the replay format, and even two original bugs — is reproduced as‑is.
The verification path is `go build` + `-selftest` (all 21 levels load and step
without panics); on‑screen visuals and audio require a real display to confirm.
