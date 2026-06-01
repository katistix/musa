package app

import (
	"image"
	"image/draw"
	"image/png"
	"math"
	"os"
	"path/filepath"
)

func prepareIcon(path string) string {
	f, err := os.Open(path)
	if err != nil {
		return path
	}
	defer f.Close()
	img, err := png.Decode(f)
	if err != nil {
		return path
	}
	b := img.Bounds()
	s := b.Dx()
	if b.Dy() < s {
		s = b.Dy()
	}
	out := image.NewNRGBA(image.Rect(0, 0, s, s))
	draw.Draw(out, out.Bounds(), img, b.Min, draw.Src)
	applyMacMask(out)
	tmp := filepath.Join(os.TempDir(), "musa-rounded-icon.png")
	wf, err := os.Create(tmp)
	if err != nil {
		return path
	}
	defer wf.Close()
	if err := png.Encode(wf, out); err != nil {
		return path
	}
	return tmp
}

func applyMacMask(img *image.NRGBA) {
	w, h := float64(img.Bounds().Dx()), float64(img.Bounds().Dy())
	cx, cy := w/2, h/2
	// Superellipse mask approximates the macOS app icon squircle.
	a, b := w*.47, h*.47
	n := 5.0
	for y := 0; y < int(h); y++ {
		for x := 0; x < int(w); x++ {
			dx := math.Abs((float64(x)+0.5-cx)/a)
			dy := math.Abs((float64(y)+0.5-cy)/b)
			v := math.Pow(dx, n) + math.Pow(dy, n)
			if v > 1 {
				i := img.PixOffset(x, y)
				img.Pix[i+3] = 0
			}
		}
	}
}
