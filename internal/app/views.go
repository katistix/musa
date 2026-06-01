package app

import (
	"fmt"
	"strings"

	"musa/internal/ui"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func (a *App) Draw() {
	w, h := float32(rl.GetScreenWidth()), float32(rl.GetScreenHeight())
	if a.nowAnim > .01 {
		a.ensureSceneTarget(int32(w), int32(h))
		rl.BeginTextureMode(a.scene)
		a.drawBase(w, h, true)
		rl.EndTextureMode()
		rl.BeginDrawing()
		defer rl.EndDrawing()
		a.drawBlurredScene(w, h, a.nowAnim)
		a.drawNowPlaying(w, h)
		return
	}
	rl.BeginDrawing()
	defer rl.EndDrawing()
	a.drawBase(w, h, false)
}

func (a *App) drawBase(w, h float32, underOverlay bool) {
	rl.ClearBackground(rl.Black)
	ui.Gradient(w, h)
	baseMode := a.mode
	if a.mode == NowPlayingMode {
		baseMode = a.prevMode
	}
	a.drawHeader(w)
	switch baseMode {
	case AlbumMode:
		a.drawShelf(w, h)
	case TrackMode:
		a.drawAlbumTracks(w, h)
	}
	if !underOverlay {
		a.drawPlayer(w, h)
		a.drawHints(w, h, baseMode)
	}
}

func (a *App) ensureSceneTarget(w, h int32) {
	if a.scene.ID != 0 && a.sceneW == w && a.sceneH == h {
		return
	}
	if a.scene.ID != 0 {
		rl.UnloadRenderTexture(a.scene)
	}
	if a.blurA.ID != 0 {
		rl.UnloadRenderTexture(a.blurA)
	}
	if a.blurB.ID != 0 {
		rl.UnloadRenderTexture(a.blurB)
	}
	a.scene = rl.LoadRenderTexture(w, h)
	a.blurA = rl.LoadRenderTexture(w, h)
	a.blurB = rl.LoadRenderTexture(w, h)
	a.sceneW, a.sceneH = w, h
}

func (a *App) drawBlurredScene(w, h, p float32) {
	if !ui.BlurReady {
		rl.DrawRectangle(0, 0, int32(w), int32(h), rl.Color{R: 4, G: 6, B: 12, A: uint8(230 * p)})
		return
	}
	src := rl.Rectangle{X: 0, Y: 0, Width: w, Height: -h}
	dst := rl.Rectangle{X: 0, Y: 0, Width: w, Height: h}
	ui.SetBlurResolution(w, h)

	rl.BeginTextureMode(a.blurA)
	rl.ClearBackground(rl.Blank)
	rl.BeginShaderMode(ui.BlurShader)
	ui.SetBlurDirection(6*p, 0)
	rl.DrawTexturePro(a.scene.Texture, src, dst, rl.Vector2{}, 0, rl.White)
	rl.EndShaderMode()
	rl.EndTextureMode()

	rl.BeginTextureMode(a.blurB)
	rl.ClearBackground(rl.Blank)
	rl.BeginShaderMode(ui.BlurShader)
	ui.SetBlurDirection(0, 6*p)
	rl.DrawTexturePro(a.blurA.Texture, src, dst, rl.Vector2{}, 0, rl.White)
	rl.EndShaderMode()
	rl.EndTextureMode()

	rl.DrawTexturePro(a.blurB.Texture, src, dst, rl.Vector2{}, 0, rl.White)
	rl.DrawRectangle(0, 0, int32(w), int32(h), rl.Color{R: 4, G: 6, B: 12, A: uint8(220 * p)})
}

func (a *App) drawHeader(w float32) {
	ui.Text("Musa", 42, 30, 42, rl.RayWhite)
	status := "Keyboard available"
	if a.controller.Connected {
		status = "DualShock connected"
	}
	ui.TextFit(status, 172, 48, w-210, 20, ui.Fade(rl.LightGray, .72))
}

func (a *App) drawHints(w, h float32, mode Mode) {
	if a.controller.Connected {
		hints := []ui.Hint{{Button: "Dpad", Label: "Browse"}, {Button: "Cross", Label: "Open"}, {Button: "Triangle", Label: "Now Playing"}, {Button: "Options", Label: "Pause"}}
		if mode == TrackMode {
			hints = []ui.Hint{{Button: "Dpad", Label: "Select"}, {Button: "Cross", Label: "Play"}, {Button: "Circle", Label: "Back"}, {Button: "Triangle", Label: "Now Playing"}, {Button: "Options", Label: "Pause"}}
		}
		ui.DrawHints(hints, w, h)
		return
	}
	ui.DrawHints([]ui.Hint{{Button: "Arrows", Label: "Browse"}, {Button: "Enter", Label: "Open/Play"}, {Button: "N", Label: "Now Playing"}, {Button: "Space", Label: "Pause"}}, w, h)
}

func (a *App) drawShelf(w, h float32) {
	if len(a.lib.Albums) == 0 {
		ui.Text("No music found in ~/Music", 56, 150, 32, rl.RayWhite)
		return
	}
	center := w / 2
	focusSize := min(h*.33, w*.22)
	spacing := focusSize * 1.55
	baseY := h * .23
	for i := range a.lib.Albums {
		d := float32(i) - a.carouselX
		if abs(d) > 2.0 {
			continue
		}
		scale := 1 - min(abs(d)*.35, .52)
		s := focusSize * scale
		x := center + d*spacing - s/2
		y := baseY + (focusSize-s)/2
		alpha := uint8(255 * scale)
		tint := rl.Color{R: 255, G: 255, B: 255, A: alpha}
		ui.CoverOrDisc(a.lib.Cover(i), x, y, s, tint)
		if i == a.album {
			rl.DrawRectangleRoundedLines(rl.Rectangle{X: x - 10, Y: y - 10, Width: s + 20, Height: s + 20}, .05, 12, rl.Color{R: 122, G: 220, B: 190, A: 255})
		}
	}
	a.drawAlbumInfo(w, baseY+focusSize+56)
}

func (a *App) drawAlbumInfo(w, y float32) {
	al := a.lib.Albums[a.album]
	ui.TextFit(al.Title, 120, y, w-240, 40, rl.RayWhite)
	ui.TextFit(fmt.Sprintf("%s   %d tracks", al.Artist, len(al.Tracks)), 124, y+54, w-248, 23, ui.Fade(rl.LightGray, .82))
}

func (a *App) drawAlbumTracks(w, h float32) {
	if len(a.lib.Albums) == 0 {
		return
	}
	al := a.lib.Albums[a.album]
	ui.CoverOrDisc(a.lib.Cover(a.album), 56, 126, 260, rl.White)
	ui.TextFit(al.Title, 356, 132, w-410, 40, rl.RayWhite)
	ui.TextFit(al.Artist, 360, 184, w-414, 24, ui.Fade(rl.LightGray, .8))
	x, y, row := float32(360), float32(248), float32(48)
	for i, ti := range al.Tracks {
		if y+float32(i)*row > h-110 {
			break
		}
		a.drawTrackRow(i, ti, x, y+float32(i)*row, row, w)
	}
}

func (a *App) drawTrackRow(i, ti int, x, y, row, w float32) {
	t := a.lib.Tracks[ti]
	if i == a.track {
		rl.DrawRectangleRounded(rl.Rectangle{X: x - 18, Y: y - 8, Width: w - x - 56, Height: row - 8}, .32, 10, rl.Color{R: 60, G: 72, B: 105, A: 225})
	}
	col := rl.RayWhite
	if ti == a.playingTrack {
		col = rl.Color{R: 122, G: 220, B: 190, A: 255}
	}
	ui.TextFit(fmt.Sprintf("%02d  %s", i+1, t.Title), x, y, w-x-72, 24, col)
}

func (a *App) drawPlayer(w, h float32) {
	bar := rl.Rectangle{X: 64, Y: h - 118, Width: w - 128, Height: 10}
	r := clamp(a.player.Pos()/a.player.Len(), 0, 1)
	rl.DrawRectangleRounded(bar, .5, 8, rl.Color{R: 45, G: 49, B: 67, A: 255})
	rl.DrawRectangleRounded(rl.Rectangle{X: bar.X, Y: bar.Y, Width: bar.Width * r, Height: bar.Height}, .5, 8, rl.Color{R: 122, G: 220, B: 190, A: 255})
	line := fmt.Sprintf("%s / %s", ui.Dur(a.player.Pos()), ui.Dur(a.player.Len()))
	if a.playingTrack >= 0 {
		t := a.lib.Tracks[a.playingTrack]
		line += "   " + strings.TrimSpace(t.Artist+" - "+t.Title)
	}
	if a.player.Status != "" {
		line = a.player.Status
	}
	ui.TextFit(line, 64, h-94, w-128, 19, ui.Fade(rl.RayWhite, .86))
}

func (a *App) drawNowPlaying(w, h float32) {
	p := easeOutBack(clamp(a.nowAnim, 0, 1))
	offset := (1 - p) * 54
	alpha := uint8(255 * clamp(p, 0, 1))
	rl.DrawRectangle(0, 0, int32(w), int32(h), rl.Color{R: 6, G: 8, B: 14, A: uint8(42 * p)})
	i := a.playingTrack
	if i < 0 {
		ui.TextFit("Nothing playing", 72, h*.36+offset, w-144, 58, rl.Color{R: 255, G: 255, B: 255, A: alpha})
		ui.TextFit("Pick a track from one of your records.", 76, h*.36+72+offset, w-152, 28, ui.Fade(rl.LightGray, .82*p))
		a.drawNowPlayingWaveform(w, h)
		a.drawNowPlayingHints(w, h)
		return
	}
	cover := min(h*.50, w*.34) * (.92 + .08*p)
	x := 74 + offset*.35
	y := h*.18 + offset*.25
	albumIdx := a.albumForTrack(i)
	rl.DrawRectangleRounded(rl.Rectangle{X: x + 18, Y: y + 22, Width: cover, Height: cover}, .08, 12, rl.Color{R: 0, G: 0, B: 0, A: uint8(95 * p)})
	ui.CoverOrDisc(a.lib.Cover(albumIdx), x, y, cover, rl.Color{R: 255, G: 255, B: 255, A: alpha})
	t := a.lib.Tracks[i]
	tx := x + cover + 58
	ui.TextFit("NOW PLAYING", tx, y+18, w-tx-72, 20, rl.Color{R: 122, G: 220, B: 190, A: alpha})
	ui.TextFit(t.Title, tx, y+62, w-tx-72, 54, rl.Color{R: 255, G: 255, B: 255, A: alpha})
	ui.TextFit(ui.Meta(t.Artist, t.Album), tx+2, y+132, w-tx-74, 30, ui.Fade(rl.LightGray, .88*p))
	a.drawNowPlayingWaveform(w, h)
	a.drawNowPlayingHints(w, h)
}

func (a *App) drawNowPlayingHints(w, h float32) {
	if a.controller.Connected {
		ui.DrawHints([]ui.Hint{{Button: "RStick", Label: "Scrub"}, {Button: "Dpad", Label: "Skip 10s"}, {Button: "Cross", Label: "Play/Pause"}, {Button: "Circle", Label: "Back"}, {Button: "Triangle", Label: "Close"}}, w, h)
		return
	}
	ui.DrawHints([]ui.Hint{{Button: "Left/Right", Label: "Seek"}, {Button: "Space", Label: "Pause"}, {Button: "N", Label: "Close"}}, w, h)
}

func (a *App) drawNowPlayingWaveform(w, h float32) {
	wave := a.player.Waveform
	p := clamp(a.nowAnim, 0, 1)
	left, right := float32(72), w-72
	top, height := h-230+(1-p)*28, float32(96)
	mid := top + height/2
	if len(wave) == 0 {
		rl.DrawRectangleRounded(rl.Rectangle{X: left, Y: top, Width: right - left, Height: height}, .08, 10, rl.Color{R: 30, G: 36, B: 52, A: uint8(190 * p)})
		ui.TextFit("No waveform yet", left+24, mid-13, right-left-48, 24, ui.Fade(rl.RayWhite, .72*p))
		return
	}
	gap := float32(2)
	bw := (right-left)/float32(len(wave)) - gap
	if bw < 2 {
		bw = 2
	}
	played := clamp(a.player.Pos()/a.player.Len(), 0, 1)
	for i, v := range wave {
		x := left + float32(i)*(bw+gap)
		amp := 10 + v*height*.92
		progress := float32(i) / float32(len(wave)-1)
		col := rl.Color{R: 56, G: 70, B: 86, A: uint8(185 * p)}
		if progress <= played {
			col = rl.Color{R: 122, G: 220, B: 190, A: uint8(235 * p)}
		}
		rl.DrawRectangleRounded(rl.Rectangle{X: x, Y: mid - amp/2, Width: bw, Height: amp}, .7, 4, col)
	}
	ui.TextFit(ui.Dur(a.player.Pos()), left, top+height+12, 120, 20, ui.Fade(rl.RayWhite, .86*p))
	ui.TextFit(ui.Dur(a.player.Len()), right-90, top+height+12, 90, 20, ui.Fade(rl.RayWhite, .86*p))
}

func (a *App) albumForTrack(track int) int {
	for ai, al := range a.lib.Albums {
		for _, ti := range al.Tracks {
			if ti == track {
				return ai
			}
		}
	}
	return a.album
}
