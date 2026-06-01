package app

import (
	"fmt"
	"math"
	"strings"

	"musa/internal/music"
	"musa/internal/ui"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type Mode int

const (
	AlbumMode Mode = iota
	TrackMode
)

type App struct {
	lib          music.Library
	player       *Player
	mode         Mode
	album        int
	track        int
	playingTrack int
	carouselX    float32
	query        string
}

func Run() {
	rl.SetConfigFlags(rl.FlagWindowResizable | rl.FlagMsaa4xHint)
	rl.InitWindow(1220, 760, "Musa — your music shelf")
	defer rl.CloseWindow()
	rl.InitAudioDevice()
	defer rl.CloseAudioDevice()
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
	if rl.IsKeyPressed(rl.KeyTab) {
		if a.mode == AlbumMode {
			a.mode = TrackMode
		} else {
			a.mode = AlbumMode
		}
	}
	if rl.IsKeyPressed(rl.KeySpace) {
		a.player.TogglePause()
	}
	mw := rl.GetMouseWheelMove()
	if ctrl() && mw != 0 {
		a.player.Volume = clamp(a.player.Volume+mw*.05, 0, 1)
		return
	}
	if a.mode == AlbumMode {
		a.updateAlbums(mw)
	} else {
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
		a.mode = TrackMode
		a.track = 0
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
	if !rl.IsGamepadAvailable(0) {
		return
	}
	if rl.IsGamepadButtonPressed(0, rl.GamepadButtonRightFaceDown) {
		if a.mode == AlbumMode {
			a.mode = TrackMode
			a.track = 0
		} else {
			a.playSelected()
		}
	}
	if rl.IsGamepadButtonPressed(0, rl.GamepadButtonRightFaceRight) {
		a.mode = AlbumMode
	}
	if rl.IsGamepadButtonPressed(0, rl.GamepadButtonMiddleRight) {
		a.player.TogglePause()
	}
	if a.mode == AlbumMode {
		if rl.IsGamepadButtonPressed(0, rl.GamepadButtonLeftFaceRight) {
			a.album = minInt(a.album+1, len(a.lib.Albums)-1)
		}
		if rl.IsGamepadButtonPressed(0, rl.GamepadButtonLeftFaceLeft) {
			a.album = maxInt(a.album-1, 0)
		}
	} else {
		tracks := a.albumTracks()
		if rl.IsGamepadButtonPressed(0, rl.GamepadButtonLeftFaceDown) {
			a.track = minInt(a.track+1, len(tracks)-1)
		}
		if rl.IsGamepadButtonPressed(0, rl.GamepadButtonLeftFaceUp) {
			a.track = maxInt(a.track-1, 0)
		}
	}
}

func (a *App) handleClick() {
	m := rl.GetMousePosition()
	w, h := float32(rl.GetScreenWidth()), float32(rl.GetScreenHeight())
	bar := rl.Rectangle{X: 28, Y: h - 58, Width: w - 56, Height: 9}
	if rl.CheckCollisionPointRec(m, bar) {
		a.player.Seek((m.X - bar.X) / bar.Width)
		return
	}
	if a.mode == AlbumMode && len(a.lib.Albums) > 0 {
		center := w / 2
		spacing := float32(210)
		for i := range a.lib.Albums {
			x := center + (float32(i)-a.carouselX)*spacing
			s := float32(172) * (1 - min(abs(float32(i)-a.carouselX)*.12, .42))
			if m.X >= x-s/2 && m.X <= x+s/2 && m.Y >= 185 && m.Y <= 185+s {
				a.album = i
				if abs(float32(i)-a.carouselX) < .15 {
					a.mode = TrackMode
				}
				return
			}
		}
	}
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

func (a *App) Draw() {
	w, h := float32(rl.GetScreenWidth()), float32(rl.GetScreenHeight())
	rl.BeginDrawing()
	defer rl.EndDrawing()
	rl.ClearBackground(rl.Black)
	ui.Gradient(w, h)
	rl.DrawText("Musa", 28, 22, 34, rl.RayWhite)
	rl.DrawText("←/→ browse · Enter/✕ open-play · ○ back · Options/Space pause", 126, 36, 15, ui.Fade(rl.LightGray, .78))
	if a.mode == AlbumMode {
		a.drawShelf(w, h)
	} else {
		a.drawAlbumTracks(w, h)
	}
	a.drawPlayer(w, h)
}

func (a *App) drawShelf(w, h float32) {
	if len(a.lib.Albums) == 0 {
		rl.DrawText("No music found in ~/Music", 40, 120, 24, rl.RayWhite)
		return
	}
	center := w / 2
	spacing := float32(210)
	baseY := float32(185)
	for i := range a.lib.Albums {
		d := float32(i) - a.carouselX
		if abs(d) > 3.5 {
			continue
		}
		scale := 1 - min(abs(d)*.12, .42)
		s := 190 * scale
		x := center + d*spacing - s/2
		y := baseY + abs(d)*24
		alpha := uint8(255 * scale)
		tint := rl.Color{R: 255, G: 255, B: 255, A: alpha}
		rl.DrawRectangleRounded(rl.Rectangle{X: x + 10, Y: y + 14, Width: s, Height: s}, .06, 8, rl.Color{R: 0, G: 0, B: 0, A: uint8(80 * scale)})
		ui.CoverOrDisc(a.lib.Cover(i), x, y, s, tint)
		if i == a.album {
			rl.DrawRectangleRoundedLines(rl.Rectangle{X: x - 6, Y: y - 6, Width: s + 12, Height: s + 12}, .06, 8, rl.Color{R: 122, G: 220, B: 190, A: 255})
		}
	}
	a.drawAlbumInfo(w, 430)
}

func (a *App) drawAlbumInfo(w, y float32) {
	al := a.lib.Albums[a.album]
	ui.TextFit(al.Title, 80, y, w-160, 32, rl.RayWhite)
	ui.TextFit(fmt.Sprintf("%s · %d tracks", al.Artist, len(al.Tracks)), 82, y+42, w-164, 18, ui.Fade(rl.LightGray, .78))
}

func (a *App) drawAlbumTracks(w, h float32) {
	if len(a.lib.Albums) == 0 {
		return
	}
	al := a.lib.Albums[a.album]
	ui.CoverOrDisc(a.lib.Cover(a.album), 34, 92, 210, rl.White)
	ui.TextFit(al.Title, 270, 96, w-310, 30, rl.RayWhite)
	ui.TextFit(al.Artist, 272, 134, w-312, 18, ui.Fade(rl.LightGray, .78))
	x, y, row := float32(270), float32(184), float32(38)
	for i, ti := range al.Tracks {
		if y+float32(i)*row > h-90 {
			break
		}
		t := a.lib.Tracks[ti]
		yy := y + float32(i)*row
		if i == a.track {
			rl.DrawRectangleRounded(rl.Rectangle{X: x - 12, Y: yy - 6, Width: w - x - 38, Height: 32}, .25, 8, rl.Color{R: 60, G: 72, B: 105, A: 210})
		}
		col := rl.RayWhite
		if ti == a.playingTrack {
			col = rl.Color{R: 122, G: 220, B: 190, A: 255}
		}
		ui.TextFit(fmt.Sprintf("%02d  %s", i+1, t.Title), x, yy, w-x-54, 17, col)
	}
}

func (a *App) drawPlayer(w, h float32) {
	bar := rl.Rectangle{X: 28, Y: h - 58, Width: w - 56, Height: 9}
	r := clamp(a.player.Pos()/a.player.Len(), 0, 1)
	rl.DrawRectangleRounded(bar, .5, 8, rl.Color{R: 45, G: 49, B: 67, A: 255})
	rl.DrawRectangleRounded(rl.Rectangle{X: bar.X, Y: bar.Y, Width: bar.Width * r, Height: bar.Height}, .5, 8, rl.Color{R: 122, G: 220, B: 190, A: 255})
	line := fmt.Sprintf("%s / %s", ui.Dur(a.player.Pos()), ui.Dur(a.player.Len()))
	if a.playingTrack >= 0 {
		t := a.lib.Tracks[a.playingTrack]
		line += "  ·  " + strings.TrimSpace(t.Artist+" — "+t.Title)
	}
	if a.player.Status != "" {
		line = a.player.Status
	}
	ui.TextFit(line, 28, h-36, w-56, 15, ui.Fade(rl.RayWhite, .82))
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
