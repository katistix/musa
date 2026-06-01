package ui

import (
	"os"

	rl "github.com/gen2brain/raylib-go/raylib"
)

var font rl.Font
var customFont bool

var fontCandidates = []string{
	"assets/fonts/SpaceGrotesk-Bold.ttf",
	"assets/fonts/SpaceGrotesk-SemiBold.ttf",
	"assets/fonts/SpaceGrotesk-Medium.ttf",
	"assets/fonts/SpaceGrotesk-Regular.ttf",
	"assets/fonts/SF-Mono-Semibold.otf",
	"assets/fonts/SF-Mono-Bold.otf",
	"assets/fonts/SF-Mono-Regular.otf",
	"/System/Library/Fonts/SFNSMono.ttf",
	"/System/Library/Fonts/Supplemental/PTMono.ttc",
	"/System/Library/Fonts/Supplemental/Andale Mono.ttf",
}

func LoadFont() {
	for _, path := range fontCandidates {
		if _, err := os.Stat(path); err == nil {
			font = rl.LoadFontEx(path, 48, nil, 0)
			if font.Texture.ID != 0 {
				rl.SetTextureFilter(font.Texture, rl.FilterBilinear)
				customFont = true
				return
			}
		}
	}
}

func UnloadFont() {
	if customFont {
		rl.UnloadFont(font)
	}
}

func Text(s string, x, y float32, size int32, c rl.Color) {
	if customFont {
		pos := rl.Vector2{X: x, Y: y}
		rl.DrawTextEx(font, s, pos, float32(size), 0.35, c)
		// Slight overdraw gives the UI a heavier, more legible weight even when
		// the available system font is only regular weight.
		rl.DrawTextEx(font, s, rl.Vector2{X: x + 0.55, Y: y}, float32(size), 0.35, c)
		return
	}
	rl.DrawText(s, int32(x), int32(y), size, c)
	rl.DrawText(s, int32(x+1), int32(y), size, c)
}

func Measure(s string, size int32) int32 {
	if customFont {
		return int32(rl.MeasureTextEx(font, s, float32(size), 0.35).X + 1)
	}
	return rl.MeasureText(s, size)
}
