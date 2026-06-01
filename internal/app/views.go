package app

import (
	"fmt"
	"strings"

	"musa/internal/ui"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func (a *App) Draw() {
	w, h := float32(rl.GetScreenWidth()), float32(rl.GetScreenHeight())
	rl.BeginDrawing()
	defer rl.EndDrawing()
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
	if a.nowAnim > .01 {
		a.drawNowPlaying(w, h)
	} else {
		a.drawPlayer(w, h)
	}
}

func (a *App) drawHeader(w float32) {
	ui.Text("Musa", 42, 30, 42, rl.RayWhite)
	hint := "Arrows browse   Enter play/open   N now playing   Space pause"
	if a.controller.Connected {
		hint = "DualShock: D-pad or left stick browse   Cross play/open   Triangle now playing   Circle back   Options pause"
	}
	ui.TextFit(hint, 172, 48, w-210, 20, ui.Fade(rl.LightGray, .82))
}

func (a *App) drawShelf(w, h float32) {
	if len(a.lib.Albums) == 0 {
		ui.Text("No music found in ~/Music", 56, 150, 32, rl.RayWhite)
		return
	}
	center, spacing, baseY := w/2, float32(235), float32(175)
	for i := range a.lib.Albums {
		d := float32(i) - a.carouselX
		if abs(d) > 3.3 {
			continue
		}
		scale := 1 - min(abs(d)*.12, .42)
		s := 220 * scale
		x := center + d*spacing - s/2
		y := baseY + abs(d)*26
		tint := rl.Color{R: 255, G: 255, B: 255, A: uint8(255 * scale)}
		rl.DrawRectangleRounded(rl.Rectangle{X: x + 16, Y: y + 20, Width: s, Height: s}, .07, 10, rl.Color{R: 0, G: 0, B: 0, A: uint8(95 * scale)})
		ui.CoverOrDisc(a.lib.Cover(i), x, y, s, tint)
		if i == a.album {
			rl.DrawRectangleRoundedLines(rl.Rectangle{X: x - 8, Y: y - 8, Width: s + 16, Height: s + 16}, .07, 10, rl.Color{R: 122, G: 220, B: 190, A: 255})
		}
	}
	a.drawAlbumInfo(w, h-210)
}

func (a *App) drawAlbumInfo(w, y float32) {
	al := a.lib.Albums[a.album]
	ui.TextFit(al.Title, 80, y, w-160, 42, rl.RayWhite)
	ui.TextFit(fmt.Sprintf("%s   %d tracks", al.Artist, len(al.Tracks)), 84, y+58, w-168, 24, ui.Fade(rl.LightGray, .82))
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
	bar := rl.Rectangle{X: 46, Y: h - 70, Width: w - 92, Height: 12}
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
	ui.TextFit(line, 46, h-42, w-92, 20, ui.Fade(rl.RayWhite, .86))
}

func (a *App) drawNowPlaying(w, h float32) {
	p := easeOutBack(clamp(a.nowAnim, 0, 1))
	offset := (1 - p) * 54
	alpha := uint8(255 * clamp(p, 0, 1))
	rl.DrawRectangle(0, 0, int32(w), int32(h), rl.Color{R: 6, G: 8, B: 14, A: uint8(120 * p)})
	i := a.playingTrack
	if i < 0 {
		ui.TextFit("Nothing playing", 72, h*.36+offset, w-144, 58, rl.Color{R: 255, G: 255, B: 255, A: alpha})
		ui.TextFit("Pick a track from one of your records.", 76, h*.36+72+offset, w-152, 28, ui.Fade(rl.LightGray, .82*p))
		a.drawNowPlayingWaveform(w, h)
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
	ui.TextFit("Circle or Triangle to return", tx+2, y+cover-34, w-tx-74, 22, ui.Fade(rl.LightGray, .68*p))
	a.drawNowPlayingWaveform(w, h)
}

func (a *App) drawNowPlayingWaveform(w, h float32) {
	wave := a.player.Waveform
	p := clamp(a.nowAnim, 0, 1)
	left, right := float32(72), w-72
	top, height := h-178+(1-p)*28, float32(112)
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
	ui.TextFit(ui.Dur(a.player.Pos()), left, top+height+16, 120, 20, ui.Fade(rl.RayWhite, .86*p))
	ui.TextFit(ui.Dur(a.player.Len()), right-90, top+height+16, 90, 20, ui.Fade(rl.RayWhite, .86*p))
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
