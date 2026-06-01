package app

import (
	"math"

	"musa/internal/music"
	"musa/internal/ui"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type Mode int

const (
	AlbumMode Mode = iota
	TrackMode
	NowPlayingMode
)

type App struct {
	lib          music.Library
	player       *Player
	mode         Mode
	prevMode     Mode
	album        int
	track        int
	playingTrack int
	carouselX    float32
	padCooldown  float32
	controller   Controller
	nowAnim      float32
}

func Run() {
	rl.SetConfigFlags(rl.FlagWindowResizable | rl.FlagMsaa4xHint)
	rl.InitWindow(1280, 720, "Musa - your music shelf")
	rl.SetExitKey(0)
	defer rl.CloseWindow()
	rl.InitAudioDevice()
	defer rl.CloseAudioDevice()
	ui.LoadFont()
	defer ui.UnloadFont()
	rl.SetTargetFPS(60)
	a := &App{lib: music.Scan(), player: NewPlayer(), playingTrack: -1}
	defer a.Close()
	for !rl.WindowShouldClose() {
		a.Update()
		a.Draw()
	}
}

func (a *App) Close() { a.player.Close(); a.lib.Unload() }

func (a *App) Update() {
	a.player.Update()
	a.controller = DetectController()
	if a.padCooldown > 0 {
		a.padCooldown -= rl.GetFrameTime()
	}
	if rl.IsKeyPressed(rl.KeyN) {
		a.toggleNowPlaying()
	}
	if rl.IsKeyPressed(rl.KeyTab) || rl.IsKeyPressed(rl.KeyEscape) {
		a.back()
	}
	if rl.IsKeyPressed(rl.KeySpace) {
		a.player.TogglePause()
	}
	target := float32(0)
	if a.mode == NowPlayingMode {
		target = 1
	}
	a.nowAnim += (target - a.nowAnim) * .18
	mw := rl.GetMouseWheelMove()
	if ctrl() && mw != 0 {
		a.player.Volume = clamp(a.player.Volume+mw*.05, 0, 1)
		return
	}
	switch a.mode {
	case AlbumMode:
		a.updateAlbums(mw)
	case TrackMode:
		a.updateTracks(mw)
	}
	a.updateGamepad()
	if rl.IsMouseButtonPressed(rl.MouseLeftButton) {
		a.handleClick()
	}
}

func (a *App) updateAlbums(wheel float32) {
	if len(a.lib.Albums) == 0 {
		return
	}
	if rl.IsKeyPressed(rl.KeyRight) || rl.IsKeyPressed(rl.KeyD) || wheel < 0 {
		a.album = minInt(a.album+1, len(a.lib.Albums)-1)
	}
	if rl.IsKeyPressed(rl.KeyLeft) || rl.IsKeyPressed(rl.KeyA) || wheel > 0 {
		a.album = maxInt(a.album-1, 0)
	}
	if rl.IsKeyPressed(rl.KeyEnter) {
		a.openAlbum()
	}
	a.carouselX += (float32(a.album) - a.carouselX) * .16
}

func (a *App) updateTracks(wheel float32) {
	tracks := a.albumTracks()
	if len(tracks) == 0 {
		return
	}
	if rl.IsKeyPressed(rl.KeyDown) || rl.IsKeyPressed(rl.KeyS) || wheel < 0 {
		a.track = minInt(a.track+1, len(tracks)-1)
	}
	if rl.IsKeyPressed(rl.KeyUp) || rl.IsKeyPressed(rl.KeyW) || wheel > 0 {
		a.track = maxInt(a.track-1, 0)
	}
	if rl.IsKeyPressed(rl.KeyLeft) {
		a.mode = AlbumMode
	}
	if rl.IsKeyPressed(rl.KeyEnter) {
		a.playSelected()
	}
}

func (a *App) updateGamepad() {
	if !a.controller.Connected {
		return
	}
	if padPressed(rl.GamepadButtonRightFaceUp) {
		a.toggleNowPlaying()
	} // Triangle
	if padPressed(rl.GamepadButtonRightFaceDown) {
		if a.mode == AlbumMode {
			a.openAlbum()
		} else if a.mode == TrackMode {
			a.playSelected()
		} else {
			a.player.TogglePause()
		}
	}
	if padPressed(rl.GamepadButtonRightFaceRight) || padPressed(rl.GamepadButtonMiddleLeft) {
		if a.mode == NowPlayingMode {
			a.toggleNowPlaying()
		} else {
			a.back()
		}
	} // Circle / Share
	if padPressed(rl.GamepadButtonMiddleRight) {
		a.player.TogglePause()
	} // Options
	if a.padCooldown > 0 || a.mode == NowPlayingMode {
		return
	}
	dx, dy, moved := padAxis(rl.GamepadAxisLeftX), padAxis(rl.GamepadAxisLeftY), false
	if a.mode == AlbumMode && len(a.lib.Albums) > 0 {
		if padPressed(rl.GamepadButtonLeftFaceRight) || dx > 0 {
			a.album = minInt(a.album+1, len(a.lib.Albums)-1)
			moved = true
		}
		if padPressed(rl.GamepadButtonLeftFaceLeft) || dx < 0 {
			a.album = maxInt(a.album-1, 0)
			moved = true
		}
	} else if tracks := a.albumTracks(); len(tracks) > 0 {
		if padPressed(rl.GamepadButtonLeftFaceDown) || dy > 0 {
			a.track = minInt(a.track+1, len(tracks)-1)
			moved = true
		}
		if padPressed(rl.GamepadButtonLeftFaceUp) || dy < 0 {
			a.track = maxInt(a.track-1, 0)
			moved = true
		}
	}
	if moved {
		a.padCooldown = .16
	}
}

func (a *App) handleClick() {
	m := rl.GetMousePosition()
	w, h := float32(rl.GetScreenWidth()), float32(rl.GetScreenHeight())
	bar := rl.Rectangle{X: 46, Y: h - 70, Width: w - 92, Height: 12}
	if rl.CheckCollisionPointRec(m, bar) {
		a.player.Seek((m.X - bar.X) / bar.Width)
		return
	}
	if a.mode != AlbumMode || len(a.lib.Albums) == 0 {
		return
	}
	center, spacing := w/2, float32(235)
	for i := range a.lib.Albums {
		d := float32(i) - a.carouselX
		s := float32(210) * (1 - min(abs(d)*.12, .42))
		x := center + d*spacing - s/2
		if m.X >= x && m.X <= x+s && m.Y >= 175 && m.Y <= 175+s {
			a.album = i
			if abs(d) < .15 {
				a.openAlbum()
			}
			return
		}
	}
}

func (a *App) openAlbum() { a.mode = TrackMode; a.track = 0 }
func (a *App) back() {
	if a.mode == TrackMode {
		a.mode = AlbumMode
	}
}

func (a *App) toggleNowPlaying() {
	if a.mode == NowPlayingMode {
		a.mode = a.prevMode
		return
	}
	a.prevMode = a.mode
	a.mode = NowPlayingMode
}
func (a *App) playSelected() {
	tracks := a.albumTracks()
	if a.track < 0 || a.track >= len(tracks) {
		return
	}
	ti := tracks[a.track]
	if a.player.Play(a.lib.Tracks[ti].Path) {
		a.playingTrack = ti
	}
}
func (a *App) albumTracks() []int {
	if a.album < 0 || a.album >= len(a.lib.Albums) {
		return nil
	}
	return a.lib.Albums[a.album].Tracks
}

func ctrl() bool { return rl.IsKeyDown(rl.KeyLeftControl) || rl.IsKeyDown(rl.KeyRightControl) }
func clamp(v, lo, hi float32) float32 {
	return float32(math.Max(float64(lo), math.Min(float64(hi), float64(v))))
}
func min(a, b float32) float32 {
	if a < b {
		return a
	}
	return b
}
func abs(v float32) float32 {
	if v < 0 {
		return -v
	}
	return v
}
func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func easeOutBack(x float32) float32 {
	c1 := float32(1.70158)
	c3 := c1 + 1
	return 1 + c3*float32(math.Pow(float64(x-1), 3)) + c1*float32(math.Pow(float64(x-1), 2))
}
