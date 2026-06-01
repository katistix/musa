package ui

import (
	"fmt"
	"math"
	"strings"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func TextFit(s string, x, y, maxW float32, size int32, c rl.Color) {
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

func CoverOrDisc(tex *rl.Texture2D, x, y, s float32, tint rl.Color) {
	if tex != nil {
		rl.DrawTexturePro(*tex, rl.Rectangle{Width: float32(tex.Width), Height: float32(tex.Height)}, rl.Rectangle{X: x, Y: y, Width: s, Height: s}, rl.Vector2{}, 0, tint)
		return
	}
	rl.DrawCircle(int32(x+s/2), int32(y+s/2), s/2, rl.Color{R: 70, G: 76, B: 105, A: tint.A})
	rl.DrawCircle(int32(x+s/2), int32(y+s/2), s/6, rl.Color{R: 14, G: 16, B: 24, A: tint.A})
}

func Gradient(w, h float32) {
	for i := int32(0); i < int32(h); i++ {
		c := uint8(24 + 18*float32(i)/h)
		rl.DrawLine(0, i, int32(w), i, rl.Color{R: 18, G: c, B: 38, A: 255})
	}
}
func Fade(c rl.Color, f float32) rl.Color { c.A = uint8(float32(c.A) * f); return c }
func Dur(s float32) string {
	if s < 0 || math.IsNaN(float64(s)) {
		s = 0
	}
	return fmt.Sprintf("%d:%02d", int(s)/60, int(s)%60)
}
func Meta(artist, album string) string {
	return strings.Trim(strings.TrimSpace(artist)+"  •  "+strings.TrimSpace(album), " •")
}
