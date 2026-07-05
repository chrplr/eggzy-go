# Eggzy — Go port

**▶ Play it in your browser: <https://chrplr.github.io/eggzy-go/>**

A Go re-implementation of the Pygame Zero game **Eggzy** from *Code the Classics
Volume 2* (Raspberry Pi Press), built on
[go-sdl3](https://github.com/Zyko0/go-sdl3) and the
[pgzgo](https://github.com/chrplr/pgzgo) harness.

All images, sounds and music are embedded, so `go build` produces a single
self-contained binary that needs no asset files at run time. Keyboard and gamepad
are both supported.

## Download

Prebuilt, self-contained binaries — no install, no dependencies, assets embedded.
Grab the latest for your platform:

- **Linux** (amd64) — [eggzy-linux-amd64.tar.gz](https://github.com/chrplr/eggzy-go/releases/latest/download/eggzy-linux-amd64.tar.gz)
- **macOS** (Apple Silicon) — [eggzy-macos-arm64.tar.gz](https://github.com/chrplr/eggzy-go/releases/latest/download/eggzy-macos-arm64.tar.gz)
- **Windows** (amd64) — [eggzy-windows-amd64.zip](https://github.com/chrplr/eggzy-go/releases/latest/download/eggzy-windows-amd64.zip)

All versions are on the [releases page](https://github.com/chrplr/eggzy-go/releases).

## Run

```sh
go run .
```

go-sdl3 bundles the SDL3, SDL3_image and SDL3_mixer libraries and extracts them at
startup, so no system SDL install is needed.

## Headless self-test

```sh
go run . -selftest   # steps the game logic without a window, then exits
```

## Provenance & license

Ported to Go from the Python original in *Code the Classics Volume 2*. The game
design and original assets are © their respective authors / Raspberry Pi Press —
add the appropriate license before redistributing.

The original Python code and assets are in Raspberry Pi Press's [Code the Classics — Volume 2](https://github.com/raspberrypipress/Code-the-Classics-Vol2) repository.

The Go source code of this port is distributed under the MIT License.

See `Python_and_Go_implementation_comparison.md` for a walkthrough of how the port
maps onto the original.
