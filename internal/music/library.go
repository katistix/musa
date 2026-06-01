package music

import (
	"bytes"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/dhowden/tag"
	rl "github.com/gen2brain/raylib-go/raylib"
)

type Track struct {
	Path, Title, Artist, Album string
	Cover                      []byte
}

type Album struct {
	Title, Artist string
	Tracks        []int
	Cover         []byte
	CoverTex      *rl.Texture2D
}

type Library struct {
	Tracks []Track
	Albums []Album
}

var exts = map[string]bool{".mp3": true, ".flac": true, ".ogg": true, ".wav": true, ".xm": true, ".mod": true}

func Scan() Library {
	home, _ := os.UserHomeDir()
	root := filepath.Join(home, "Music")
	var lib Library
	filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || !exts[strings.ToLower(filepath.Ext(path))] {
			return nil
		}
		t := Track{Path: path, Title: strings.TrimSuffix(filepath.Base(path), filepath.Ext(path)), Album: filepath.Base(filepath.Dir(path))}
		if f, err := os.Open(path); err == nil {
			if m, err := tag.ReadFrom(f); err == nil {
				if m.Title() != "" {
					t.Title = m.Title()
				}
				t.Artist = m.Artist()
				if m.Album() != "" {
					t.Album = m.Album()
				}
				if p := m.Picture(); p != nil {
					t.Cover = append([]byte(nil), p.Data...)
				}
			}
			_ = f.Close()
		}
		lib.Tracks = append(lib.Tracks, t)
		return nil
	})
	lib.buildAlbums()
	return lib
}

func (l *Library) buildAlbums() {
	byKey := map[string]int{}
	for i, t := range l.Tracks {
		album, artist := safe(t.Album, filepath.Base(filepath.Dir(t.Path))), safe(t.Artist, "Unknown Artist")
		key := strings.ToLower(artist + "\x00" + album)
		ai, ok := byKey[key]
		if !ok {
			ai = len(l.Albums)
			byKey[key] = ai
			l.Albums = append(l.Albums, Album{Title: album, Artist: artist})
		}
		l.Albums[ai].Tracks = append(l.Albums[ai].Tracks, i)
		if len(l.Albums[ai].Cover) == 0 && len(t.Cover) > 0 {
			l.Albums[ai].Cover = append([]byte(nil), t.Cover...)
		}
	}
	sort.Slice(l.Albums, func(i, j int) bool {
		return strings.ToLower(l.Albums[i].Artist+l.Albums[i].Title) < strings.ToLower(l.Albums[j].Artist+l.Albums[j].Title)
	})
}

func (l *Library) Unload() {
	for i := range l.Albums {
		if l.Albums[i].CoverTex != nil {
			rl.UnloadTexture(*l.Albums[i].CoverTex)
		}
	}
}

func (l *Library) Cover(album int) *rl.Texture2D {
	if album < 0 || album >= len(l.Albums) {
		return nil
	}
	a := &l.Albums[album]
	if a.CoverTex != nil {
		return a.CoverTex
	}
	if len(a.Cover) == 0 {
		return nil
	}
	img, _, err := image.Decode(bytes.NewReader(a.Cover))
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
	a.CoverTex = &tex
	return a.CoverTex
}

func safe(s, fallback string) string {
	if strings.TrimSpace(s) == "" {
		return fallback
	}
	return strings.TrimSpace(s)
}
