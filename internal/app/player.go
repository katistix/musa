package app

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type Player struct {
	music     rl.Music
	loaded    bool
	Paused    bool
	Volume    float32
	Status    string
	tempFiles []string
}

func NewPlayer() *Player { return &Player{Volume: .85} }

func (p *Player) Update() {
	if p.loaded && validMusic(p.music) && !p.Paused {
		rl.UpdateMusicStream(p.music)
	}
}
func (p *Player) Loaded() bool { return p.loaded && validMusic(p.music) }
func (p *Player) Pos() float32 {
	if !p.Loaded() {
		return 0
	}
	return rl.GetMusicTimePlayed(p.music)
}
func (p *Player) Len() float32 {
	if !p.Loaded() {
		return 1
	}
	l := rl.GetMusicTimeLength(p.music)
	if l <= 0 {
		return 1
	}
	return l
}
func (p *Player) Seek(ratio float32) {
	if p.Loaded() {
		rl.SeekMusicStream(p.music, clamp(ratio, 0, 1)*p.Len())
	}
}

func (p *Player) TogglePause() {
	if !p.Loaded() {
		return
	}
	p.Paused = !p.Paused
	if p.Paused {
		rl.PauseMusicStream(p.music)
	} else {
		rl.ResumeMusicStream(p.music)
	}
}

func (p *Player) Play(path string) bool {
	p.Stop()
	playPath, cleanup, err := playablePath(path)
	if err != nil {
		p.Status = err.Error()
		return false
	}
	if cleanup != "" {
		p.tempFiles = append(p.tempFiles, cleanup)
	}
	m := rl.LoadMusicStream(playPath)
	if !validMusic(m) {
		p.Status = "Unsupported or unreadable: " + filepath.Base(path)
		return false
	}
	p.music = m
	rl.SetMusicVolume(p.music, p.Volume)
	rl.PlayMusicStream(p.music)
	p.loaded, p.Paused, p.Status = true, false, ""
	return true
}

func (p *Player) Stop() {
	if p.loaded {
		rl.StopMusicStream(p.music)
		rl.UnloadMusicStream(p.music)
		p.loaded = false
	}
}
func (p *Player) Close() {
	p.Stop()
	for _, f := range p.tempFiles {
		_ = os.Remove(f)
	}
}

func playablePath(path string) (string, string, error) {
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
	_ = out.Close()
	cmd := exec.Command("ffmpeg", "-y", "-v", "error", "-i", path, "-f", "wav", "-acodec", "pcm_s16le", outPath)
	if err := cmd.Run(); err != nil {
		_ = os.Remove(outPath)
		return "", "", fmt.Errorf("Could not decode FLAC: %s", filepath.Base(path))
	}
	return outPath, outPath, nil
}

func validMusic(m rl.Music) bool { return m.CtxData != nil && m.Stream.Buffer != nil }
