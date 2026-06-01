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
	MenuMode
)

type App struct {
	lib          music.Library
	player       *Player
	mode         Mode
	prevMode     Mode
	album        int
	track        int
	playingTrack int
	menuSel      int
	state        AppState
	carouselX    float32
	padCooldown  float32
	controller   Controller
	nowAnim      float32
	trackAnim    float32
	trackDir     float32
	pendingDir   float32
	lastTrack    int
	scrubAccum   float32
	scene        rl.RenderTexture2D
	blurA        rl.RenderTexture2D
	blurB        rl.RenderTexture2D
	sceneW       int32
	sceneH       int32
}

func Run() {
	rl.SetConfigFlags(rl.FlagWindowResizable | rl.FlagMsaa4xHint)
	rl.InitWindow(1280, 720, "Musa - your music shelf")
	setDockIcon("assets/icon.png")
	if icon := rl.LoadImage("assets/icon.png"); icon.Data != nil {
		rl.SetWindowIcon(*icon)
		rl.UnloadImage(icon)
	}
	rl.SetExitKey(0)
	defer rl.CloseWindow()
	rl.InitAudioDevice()
	defer rl.CloseAudioDevice()
	ui.LoadFont()
	defer ui.UnloadFont()
	ui.LoadShaders()
	defer ui.UnloadShaders()
	rl.SetTargetFPS(60)

	a := &App{lib: music.Scan(), player: NewPlayer(), playingTrack: -1, state: loadState(), lastTrack: -1}
	a.restoreState()
	defer a.Close()
	for !rl.WindowShouldClose() {
		a.Update()
		a.Draw()
	}
}

func (a *App) Close() {
	a.persistState()
	a.player.Close()
	a.lib.Unload()
	if a.scene.ID != 0 {
		rl.UnloadRenderTexture(a.scene)
	}
	if a.blurA.ID != 0 {
		rl.UnloadRenderTexture(a.blurA)
	}
	if a.blurB.ID != 0 {
		rl.UnloadRenderTexture(a.blurB)
	}
}

func (a *App) Update() {
	a.player.Update()
	a.updatePlayback()
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
	if rl.IsKeyPressed(rl.KeyR) && ctrl() {
		a.rescanLibrary()
	}
	if rl.IsKeyPressed(rl.KeySpace) {
		a.player.TogglePause()
	}

	target := float32(0)
	if a.mode == NowPlayingMode {
		target = 1
	}
	a.nowAnim += (target - a.nowAnim) * .34
	a.trackAnim *= .80
	if a.trackAnim < .01 {
		a.trackAnim = 0
		a.lastTrack = -1
	}

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
	case MenuMode:
		a.updateMenu(mw)
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

func (a *App) updateMenu(wheel float32) {
	if rl.IsKeyPressed(rl.KeyDown) || wheel < 0 {
		a.menuSel = minInt(a.menuSel+1, 0)
	}
	if rl.IsKeyPressed(rl.KeyUp) || wheel > 0 {
		a.menuSel = maxInt(a.menuSel-1, 0)
	}
	if rl.IsKeyPressed(rl.KeyEnter) {
		a.activateMenu()
	}
}

func (a *App) activateMenu() {
	switch a.menuSel {
	case 0:
		a.rescanLibrary()
		a.mode = a.prevMode
	}
}

func (a *App) rescanLibrary() {
	currentPath := ""
	if a.playingTrack >= 0 && a.playingTrack < len(a.lib.Tracks) {
		currentPath = a.lib.Tracks[a.playingTrack].Path
	}
	a.lib.Unload()
	a.lib = music.Scan()
	if a.album >= len(a.lib.Albums) {
		a.album = maxInt(len(a.lib.Albums)-1, 0)
	}
	if currentPath != "" {
		for i, t := range a.lib.Tracks {
			if t.Path == currentPath {
				a.playingTrack = i
				break
			}
		}
	}
}

func (a *App) updatePlayback() {
	if a.player.Finished() {
		a.Next()
	}
}

func (a *App) restoreState() {
	if a.state.Album >= 0 && a.state.Album < len(a.lib.Albums) {
		a.album = a.state.Album
	}
	if a.state.Track >= 0 && a.state.Track < len(a.lib.Tracks) {
		a.playingTrack = a.state.Track
		_ = a.player.Play(a.lib.Tracks[a.state.Track].Path)
		a.player.TogglePause()
		if a.state.Position > 0 {
			a.player.Seek(a.state.Position / a.player.Len())
		}
	}
}

func (a *App) persistState() {
	a.state.Album = a.album
	a.state.Track = a.playingTrack
	a.state.Position = 0
	if a.player.Loaded() {
		a.state.Position = a.player.Pos()
	}
	saveState(a.state)
}

func (a *App) PlayNow(ti int) {
	if ti < 0 || ti >= len(a.lib.Tracks) {
		return
	}
	if a.playingTrack != -1 && a.playingTrack != ti {
		a.lastTrack = a.playingTrack
		a.trackAnim = 1
		if a.pendingDir != 0 {
			a.trackDir = a.pendingDir
			a.pendingDir = 0
		} else if nextDir(a, ti) >= 0 {
			a.trackDir = 1
		} else {
			a.trackDir = -1
		}
	}
	if a.player.Play(a.lib.Tracks[ti].Path) {
		a.playingTrack = ti
	}
}

func (a *App) Next() {
	if a.playingTrack < 0 {
		return
	}
	alIdx := a.albumForTrack(a.playingTrack)
	al := a.lib.Albums[alIdx]
	curr := 0
	for i, ti := range al.Tracks {
		if ti == a.playingTrack {
			curr = i
			break
		}
	}
	a.pendingDir = -1
	a.PlayNow(al.Tracks[(curr+1)%len(al.Tracks)])
}

func (a *App) Prev() {
	if a.playingTrack < 0 {
		return
	}
	if a.player.Pos() > 3 {
		a.player.Seek(0)
		return
	}
	alIdx := a.albumForTrack(a.playingTrack)
	al := a.lib.Albums[alIdx]
	curr := 0
	for i, ti := range al.Tracks {
		if ti == a.playingTrack {
			curr = i
			break
		}
	}
	a.pendingDir = 1
	a.PlayNow(al.Tracks[(curr-1+len(al.Tracks))%len(al.Tracks)])
}

func (a *App) updateGamepad() {
	if !rl.IsWindowFocused() || !a.controller.Connected {
		return
	}
	if padPressed(rl.GamepadButtonRightFaceUp) {
		a.toggleNowPlaying()
	}
	if padPressed(rl.GamepadButtonRightFaceDown) {
		if a.mode == AlbumMode {
			a.openAlbum()
		} else if a.mode == TrackMode {
			a.playSelected()
		} else if a.mode == MenuMode {
			a.activateMenu()
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
	}
	if padPressed(rl.GamepadButtonMiddleRight) {
		a.toggleMenu()
	}
	if padPressed(rl.GamepadButtonLeftTrigger1) || padPressed(rl.GamepadButtonLeftTrigger2) {
		a.Prev()
	}
	if padPressed(rl.GamepadButtonRightTrigger1) || padPressed(rl.GamepadButtonRightTrigger2) {
		a.Next()
	}

	if a.mode == NowPlayingMode {
		a.updateNowPlayingPad()
		return
	}
	if a.padCooldown > 0 {
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

func (a *App) updateNowPlayingPad() {
	if padPressed(rl.GamepadButtonLeftFaceRight) {
		a.player.SeekSeconds(10)
	}
	if padPressed(rl.GamepadButtonLeftFaceLeft) {
		a.player.SeekSeconds(-10)
	}
	turn := padAxis(rl.GamepadAxisRightX)
	if turn == 0 {
		turn = -padAxis(rl.GamepadAxisRightY)
	}
	if turn == 0 {
		a.scrubAccum = 0
		return
	}
	a.scrubAccum += rl.GetFrameTime()
	if a.scrubAccum >= .075 {
		a.scrubAccum = 0
		a.player.SeekSeconds(turn * abs(turn) * 1.15)
	}
}

func (a *App) handleClick() {
	m := rl.GetMousePosition()
	w, h := float32(rl.GetScreenWidth()), float32(rl.GetScreenHeight())
	bar := rl.Rectangle{X: 64, Y: h - 118, Width: w - 128, Height: 10}
	if rl.CheckCollisionPointRec(m, bar) {
		a.player.Seek((m.X - bar.X) / bar.Width)
		return
	}
	if a.mode != AlbumMode || len(a.lib.Albums) == 0 {
		return
	}
	center := w / 2
	spacing := min(h*.33, w*.22) * 1.55
	focusSize := min(h*.33, w*.22)
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
		if m.X >= x && m.X <= x+s && m.Y >= y && m.Y <= y+s {
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
		return
	}
	if a.mode == MenuMode {
		a.mode = a.prevMode
	}
}
func (a *App) toggleMenu() {
	if a.mode == MenuMode {
		a.mode = a.prevMode
		return
	}
	a.prevMode = a.mode
	a.menuSel = 0
	a.mode = MenuMode
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
	a.PlayNow(tracks[a.track])
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
func nextDir(a *App, ti int) int {
	if a.playingTrack < 0 {
		return 1
	}
	al := a.lib.Albums[a.albumForTrack(ti)]
	currIdx, nextIdx := 0, 0
	for i, t := range al.Tracks {
		if t == a.playingTrack {
			currIdx = i
		}
		if t == ti {
			nextIdx = i
		}
	}
	if nextIdx >= currIdx {
		return 1
	}
	return -1
}
func easeOutBack(x float32) float32 {
	c1 := float32(1.70158)
	c3 := c1 + 1
	return 1 + c3*float32(math.Pow(float64(x-1), 3)) + c1*float32(math.Pow(float64(x-1), 2))
}
