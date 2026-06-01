package main

import (
	"bytes"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io/fs"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/dhowden/tag"
	rl "github.com/gen2brain/raylib-go/raylib"
)

type Track struct {
	Path, Title, Artist, Album string
	Duration                   float32
	Cover                      []byte
	CoverTex                   *rl.Texture2D
}

type App struct {
	tracks       []Track
	filtered     []int
	selected     int
	playing      int
	query        string
	music        rl.Music
	loaded       bool
	paused       bool
	scroll       float32
	volume       float32
	lastClick    time.Time
	lastClickIdx int
	status       string
	tempFiles    []string
}

var exts = map[string]bool{".mp3": true, ".flac": true, ".ogg": true, ".wav": true, ".xm": true, ".mod": true}

func main() {
	rl.SetConfigFlags(rl.FlagWindowResizable | rl.FlagMsaa4xHint)
	rl.InitWindow(1180, 760, "Musa — Raylib Music Player")
	defer rl.CloseWindow()
	rl.InitAudioDevice()
	defer rl.CloseAudioDevice()
	rl.SetTargetFPS(60)

	app := &App{selected: -1, playing: -1, volume: 0.85}
	app.scanMusic()
	app.applyFilter()
	defer app.unload()

	for !rl.WindowShouldClose() {
		app.update()
		app.draw()
	}
}

func (a *App) scanMusic() {
	home, _ := os.UserHomeDir()
	root := filepath.Join(home, "Music")
	filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || !exts[strings.ToLower(filepath.Ext(path))] {
			return nil
		}
		t := Track{Path: path, Title: strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))}
		if f, err := os.Open(path); err == nil {
			if m, err := tag.ReadFrom(f); err == nil {
				if m.Title() != "" {
					t.Title = m.Title()
				}
				t.Artist, t.Album = m.Artist(), m.Album()
				if p := m.Picture(); p != nil {
					t.Cover = append([]byte(nil), p.Data...)
				}
			}
			f.Close()
		}
		a.tracks = append(a.tracks, t)
		return nil
	})
	sort.Slice(a.tracks, func(i, j int) bool { return strings.ToLower(a.tracks[i].Title) < strings.ToLower(a.tracks[j].Title) })
}

func (a *App) unload() {
	if a.loaded {
		rl.UnloadMusicStream(a.music)
	}
	for i := range a.tracks {
		if a.tracks[i].CoverTex != nil {
			rl.UnloadTexture(*a.tracks[i].CoverTex)
		}
	}
	for _, p := range a.tempFiles {
		_ = os.Remove(p)
	}
}

func (a *App) applyFilter() {
	a.filtered = a.filtered[:0]
	q := strings.ToLower(strings.TrimSpace(a.query))
	for i, t := range a.tracks {
		hay := strings.ToLower(t.Title + " " + t.Artist + " " + t.Album + " " + t.Path)
		if q == "" || strings.Contains(hay, q) {
			a.filtered = append(a.filtered, i)
		}
	}
	if len(a.filtered) == 0 {
		a.selected = -1
	} else if a.selected < 0 || a.selected >= len(a.filtered) {
		a.selected = 0
	}
}

func (a *App) update() {
	if a.loaded && validMusic(a.music) && !a.paused {
		rl.UpdateMusicStream(a.music)
	}
	for {
		ch := rl.GetCharPressed()
		if ch == 0 {
			break
		}
		if ch >= 32 && ch < 127 {
			a.query += string(rune(ch))
			a.applyFilter()
		}
	}
	if rl.IsKeyPressed(rl.KeyBackspace) && len(a.query) > 0 {
		a.query = a.query[:len(a.query)-1]
		a.applyFilter()
	}
	if rl.IsKeyPressed(rl.KeyEscape) {
		a.query = ""
		a.applyFilter()
	}
	if rl.IsKeyPressed(rl.KeyDown) && a.selected < len(a.filtered)-1 {
		a.selected++
	}
	if rl.IsKeyPressed(rl.KeyUp) && a.selected > 0 {
		a.selected--
	}
	if rl.IsKeyPressed(rl.KeyEnter) && a.selected >= 0 {
		a.play(a.filtered[a.selected])
	}
	if rl.IsKeyPressed(rl.KeySpace) && a.loaded {
		a.paused = !a.paused
		if a.paused {
			rl.PauseMusicStream(a.music)
		} else {
			rl.ResumeMusicStream(a.music)
		}
	}
	mw := rl.GetMouseWheelMove()
	if mw != 0 {
		a.scroll -= mw * 32
		if a.scroll < 0 {
			a.scroll = 0
		}
	}
	if rl.IsKeyDown(rl.KeyLeftControl) || rl.IsKeyDown(rl.KeyRightControl) {
		a.volume = clamp(a.volume+mw*0.05, 0, 1)
		if a.loaded {
			rl.SetMusicVolume(a.music, a.volume)
		}
	}
	a.handleMouse()
}

func (a *App) handleMouse() {
	m := rl.GetMousePosition()
	w, h := float32(rl.GetScreenWidth()), float32(rl.GetScreenHeight())
	listX, listY, listW, rowH := float32(24), float32(104), w*0.58, float32(50)
	if rl.IsMouseButtonPressed(rl.MouseLeftButton) {
		if m.X >= listX && m.X <= listX+listW && m.Y >= listY && m.Y <= h-92 {
			idx := int((m.Y - listY + a.scroll) / rowH)
			if idx >= 0 && idx < len(a.filtered) {
				a.selected = idx
				if idx == a.lastClickIdx && time.Since(a.lastClick) < 350*time.Millisecond {
					a.play(a.filtered[idx])
				}
				a.lastClickIdx, a.lastClick = idx, time.Now()
			}
		}
		bar := rl.Rectangle{X: 24, Y: h - 64, Width: w - 48, Height: 10}
		if a.loaded && validMusic(a.music) && rl.CheckCollisionPointRec(m, bar) {
			length := rl.GetMusicTimeLength(a.music)
			if length > 0 {
				rl.SeekMusicStream(a.music, clamp((m.X-bar.X)/bar.Width, 0, 1)*length)
			}
		}
	}
}

func (a *App) play(i int) {
	if i < 0 || i >= len(a.tracks) {
		return
	}
	if a.loaded {
		rl.StopMusicStream(a.music)
		rl.UnloadMusicStream(a.music)
		a.loaded = false
	}
	playPath, cleanup, err := a.playablePath(a.tracks[i].Path)
	if err != nil {
		a.playing, a.paused = -1, false
		a.status = err.Error()
		return
	}
	if cleanup != "" {
		a.tempFiles = append(a.tempFiles, cleanup)
	}
	m := rl.LoadMusicStream(playPath)
	if !validMusic(m) {
		a.playing, a.paused = -1, false
		a.status = "Unsupported or unreadable: " + filepath.Base(a.tracks[i].Path)
		return
	}
	a.music = m
	rl.SetMusicVolume(a.music, a.volume)
	rl.PlayMusicStream(a.music)
	a.loaded, a.paused, a.playing = true, false, i
	a.status = ""
}

func (a *App) playablePath(path string) (string, string, error) {
	// Raylib's macOS audio backend often cannot stream FLAC directly. For FLAC
	// we transparently decode to a temporary WAV and still play it through raylib.
	if strings.ToLower(filepath.Ext(path)) != ".flac" {
		return path, "", nil
	}
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		return "", "", fmt.Errorf("FLAC requires ffmpeg: brew install ffmpeg")
	}
	out, err := os.CreateTemp("", "musa-*.wav")
	if err != nil {
		return "", "", err
	}
	outPath := out.Name()
	out.Close()
	cmd := exec.Command("ffmpeg", "-y", "-v", "error", "-i", path, "-f", "wav", "-acodec", "pcm_s16le", outPath)
	if err := cmd.Run(); err != nil {
		_ = os.Remove(outPath)
		return "", "", fmt.Errorf("Could not decode FLAC: %s", filepath.Base(path))
	}
	return outPath, outPath, nil
}

func (a *App) cover(i int) *rl.Texture2D {
	if i < 0 || i >= len(a.tracks) {
		return nil
	}
	t := &a.tracks[i]
	if t.CoverTex != nil {
		return t.CoverTex
	}
	if len(t.Cover) == 0 {
		return nil
	}
	img, _, err := image.Decode(bytes.NewReader(t.Cover))
	if err != nil {
		return nil
	}
	b := img.Bounds()
	rgba := image.NewRGBA(b)
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			rgba.Set(x, y, img.At(x, y))
		}
	}
	ri := rl.NewImage(rgba.Pix, int32(b.Dx()), int32(b.Dy()), 1, rl.UncompressedR8g8b8a8)
	tex := rl.LoadTextureFromImage(ri)
	t.CoverTex = &tex
	return t.CoverTex
}

func (a *App) draw() {
	w, h := float32(rl.GetScreenWidth()), float32(rl.GetScreenHeight())
	rl.BeginDrawing()
	defer rl.EndDrawing()
	rl.ClearBackground(rl.Color{R: 14, G: 16, B: 24, A: 255})
	drawGradient(w, h)
	rl.DrawText("Musa", 24, 18, 32, rl.RayWhite)
	rl.DrawText(fmt.Sprintf("%d songs in ~/Music", len(a.tracks)), 118, 31, 15, fade(rl.LightGray, .75))
	search := rl.Rectangle{X: 24, Y: 60, Width: w * 0.58, Height: 32}
	rl.DrawRectangleRounded(search, .25, 12, rl.Color{R: 32, G: 36, B: 52, A: 255})
	q := a.query
	if q == "" {
		q = "Search title, artist, album..."
	}
	drawTextFit(q, search.X+14, search.Y+9, search.Width-28, 16, fade(rl.RayWhite, ternary(a.query == "", .42, .95)))
	a.drawList(w, h)
	a.drawNowPlaying(w, h)
}

func (a *App) drawList(w, h float32) {
	listX, listY, listW, rowH := float32(24), float32(104), w*0.58, float32(50)
	bottom := h - 92
	rl.BeginScissorMode(int32(listX), int32(listY), int32(listW), int32(bottom-listY))
	defer rl.EndScissorMode()
	visible := int((bottom-listY)/rowH) + 3
	start := int(a.scroll / rowH)
	if start < 0 {
		start = 0
	}
	for n := 0; n < visible && start+n < len(a.filtered); n++ {
		fi := start + n
		ti := a.filtered[fi]
		t := a.tracks[ti]
		y := listY + float32(fi)*rowH - a.scroll
		bg := rl.Color{R: 25, G: 28, B: 40, A: 190}
		if fi == a.selected {
			bg = rl.Color{R: 72, G: 82, B: 125, A: 230}
		}
		if ti == a.playing {
			bg = rl.Color{R: 56, G: 92, B: 88, A: 230}
		}
		rl.DrawRectangleRounded(rl.Rectangle{X: listX, Y: y, Width: listW, Height: rowH - 6}, .14, 8, bg)
		drawCoverOrDisc(a.cover(ti), listX+7, y+7, 36)
		drawTextFit(t.Title, listX+52, y+7, listW-62, 17, rl.RayWhite)
		drawTextFit(compactMeta(t), listX+52, y+29, listW-62, 13, fade(rl.LightGray, .75))
	}
}

func (a *App) drawNowPlaying(w, h float32) {
	panel := rl.Rectangle{X: w * 0.64, Y: 60, Width: w*0.34 - 24, Height: h - 152}
	rl.DrawRectangleRounded(panel, .04, 18, rl.Color{R: 22, G: 25, B: 38, A: 230})
	i := a.playing
	if i < 0 && a.selected >= 0 {
		i = a.filtered[a.selected]
	}
	pad := float32(22)
	if i >= 0 {
		cover := min(panel.Width-pad*2, panel.Height-142)
		drawCoverOrDisc(a.cover(i), panel.X+pad, panel.Y+pad, cover)
		t := a.tracks[i]
		drawTextFit(t.Title, panel.X+pad, panel.Y+pad+cover+18, panel.Width-pad*2, 22, rl.RayWhite)
		drawTextFit(compactMeta(t), panel.X+pad, panel.Y+pad+cover+48, panel.Width-pad*2, 15, fade(rl.LightGray, .75))
	} else {
		rl.DrawText("No music selected", int32(panel.X+pad), int32(panel.Y+pad), 22, rl.RayWhite)
	}
	bar := rl.Rectangle{X: 24, Y: h - 64, Width: w - 48, Height: 10}
	pos, length := float32(0), float32(1)
	if a.loaded && validMusic(a.music) {
		pos, length = rl.GetMusicTimePlayed(a.music), rl.GetMusicTimeLength(a.music)
		if length <= 0 {
			length = 1
		}
	}
	rl.DrawRectangleRounded(bar, .5, 8, rl.Color{R: 45, G: 49, B: 67, A: 255})
	rl.DrawRectangleRounded(rl.Rectangle{X: bar.X, Y: bar.Y, Width: bar.Width * clamp(pos/length, 0, 1), Height: bar.Height}, .5, 8, rl.Color{R: 122, G: 220, B: 190, A: 255})
	line := fmt.Sprintf("%s / %s   Space play/pause · Enter/double-click play · Ctrl+wheel volume %.0f%%", dur(pos), dur(length), a.volume*100)
	if a.status != "" {
		line = a.status
	}
	drawTextFit(line, 24, h-42, w-48, 15, fade(rl.RayWhite, .8))
}

func drawCoverOrDisc(tex *rl.Texture2D, x, y, s float32) {
	if tex != nil {
		rl.DrawTexturePro(*tex, rl.Rectangle{Width: float32(tex.Width), Height: float32(tex.Height)}, rl.Rectangle{X: x, Y: y, Width: s, Height: s}, rl.Vector2{}, 0, rl.White)
		return
	}
	rl.DrawCircle(int32(x+s/2), int32(y+s/2), s/2, rl.Color{R: 70, G: 76, B: 105, A: 255})
	rl.DrawCircle(int32(x+s/2), int32(y+s/2), s/6, rl.Color{R: 14, G: 16, B: 24, A: 255})
}
func drawGradient(w, h float32) {
	for i := int32(0); i < int32(h); i++ {
		c := uint8(24 + 18*float32(i)/h)
		rl.DrawLine(0, i, int32(w), i, rl.Color{R: 18, G: c, B: 38, A: 255})
	}
}
func validMusic(m rl.Music) bool { return m.CtxData != nil && m.Stream.Buffer != nil }
func compactMeta(t Track) string {
	s := strings.Trim(strings.TrimSpace(t.Artist)+"  •  "+strings.TrimSpace(t.Album), " •")
	if s == "" {
		return filepath.Base(filepath.Dir(t.Path))
	}
	return s
}
func drawTextFit(s string, x, y, maxW float32, size int32, c rl.Color) {
	if s == "" {
		return
	}
	r := []rune(s)
	for len(r) > 1 && rl.MeasureText(string(r), size) > int32(maxW) {
		r = r[:len(r)-1]
	}
	if len(r) < len([]rune(s)) && len(r) > 1 {
		r[len(r)-1] = '…'
	}
	rl.DrawText(string(r), int32(x), int32(y), size, c)
}
func fade(c rl.Color, f float32) rl.Color { c.A = uint8(float32(c.A) * f); return c }
func clamp(v, lo, hi float32) float32 {
	return float32(math.Max(float64(lo), math.Min(float64(hi), float64(v))))
}
func min(a, b float32) float32 {
	if a < b {
		return a
	}
	return b
}
func ternary(b bool, x, y float32) float32 {
	if b {
		return x
	}
	return y
}
func dur(s float32) string {
	if s < 0 || math.IsNaN(float64(s)) {
		s = 0
	}
	return fmt.Sprintf("%d:%02d", int(s)/60, int(s)%60)
}
