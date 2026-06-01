package ui

import rl "github.com/gen2brain/raylib-go/raylib"

type Hint struct{ Button, Label string }

type hintItem struct {
	hint Hint
	w    float32
}

const (
	hintIcon    = float32(30)
	hintGap     = float32(10)
	hintItemGap = float32(28)
	hintText    = int32(18)
)

func DrawHints(hints []Hint, w, h float32) {
	if len(hints) == 0 {
		return
	}
	items, total := measureHints(hints)
	maxW := w - 72
	if total > maxW {
		items, total = compactHints(hints, maxW)
	}
	x := (w - total) / 2
	y := h - 42
	// PlayStation-style footer: hints sit directly on the scene with subtle
	// separation, not inside a heavy floating island.
	rl.DrawRectangleGradientV(0, int32(h-86), int32(w), 86, rl.Color{R: 4, G: 6, B: 12, A: 0}, rl.Color{R: 4, G: 6, B: 12, A: 150})
	for _, item := range items {
		drawButton(item.hint.Button, x, y)
		Text(item.hint.Label, x+hintIcon+hintGap, y+4, hintText, Fade(rl.RayWhite, .84))
		x += item.w + hintItemGap
	}
}

func measureHints(hints []Hint) ([]hintItem, float32) {
	items := make([]hintItem, len(hints))
	total := float32(0)
	for i, h := range hints {
		items[i] = hintItem{hint: h, w: hintIcon + hintGap + float32(Measure(h.Label, hintText))}
		total += items[i].w
	}
	if len(items) > 1 {
		total += hintItemGap * float32(len(items)-1)
	}
	return items, total
}

func compactHints(hints []Hint, maxW float32) ([]hintItem, float32) {
	compact := make([]Hint, len(hints))
	for i, h := range hints {
		compact[i] = Hint{Button: h.Button, Label: shortLabel(h.Label)}
	}
	items, total := measureHints(compact)
	for len(items) > 1 && total > maxW {
		items = items[:len(items)-1]
		total = 0
		for _, it := range items {
			total += it.w
		}
		total += hintItemGap * float32(len(items)-1)
	}
	return items, total
}

func shortLabel(s string) string {
	switch s {
	case "Now Playing":
		return "Now"
	case "Play/Pause":
		return "Pause"
	case "Skip 10s":
		return "Skip"
	case "Open/Play":
		return "Play"
	}
	return s
}

func drawButton(label string, x, y float32) {
	c := rl.Color{R: 242, G: 246, B: 255, A: 242}
	muted := rl.Color{R: 132, G: 148, B: 174, A: 220}
	bg := rl.Color{R: 20, G: 24, B: 34, A: 232}
	accent := rl.Color{R: 122, G: 220, B: 190, A: 230}
	s := hintIcon
	rl.DrawCircle(int32(x+s/2), int32(y+s/2+1), s/2, rl.Color{R: 0, G: 0, B: 0, A: 70})
	switch label {
	case "Cross":
		drawFaceBase(x, y, bg, accent)
		strokeLine(x+10, y+10, x+20, y+20, c, 2.7)
		strokeLine(x+20, y+10, x+10, y+20, c, 2.7)
	case "Circle":
		drawFaceBase(x, y, bg, accent)
		rl.DrawRing(rl.Vector2{X: x + s/2, Y: y + s/2}, s*.23, s*.30, 0, 360, 48, c)
	case "Triangle":
		drawFaceBase(x, y, bg, accent)
		strokeLine(x+15, y+8, x+22, y+21, c, 2.4)
		strokeLine(x+22, y+21, x+8, y+21, c, 2.4)
		strokeLine(x+8, y+21, x+15, y+8, c, 2.4)
	case "Options":
		rl.DrawRectangleRounded(rl.Rectangle{X: x, Y: y + 4, Width: 38, Height: 22}, .5, 14, bg)
		rl.DrawRectangleRoundedLines(rl.Rectangle{X: x, Y: y + 4, Width: 38, Height: 22}, .5, 14, muted)
		strokeLine(x+11, y+12, x+27, y+12, c, 2)
		strokeLine(x+11, y+18, x+27, y+18, c, 2)
	case "Dpad":
		drawDpad(x, y, bg, c)
	case "RStick", "LStick":
		drawStick(x, y, label[:1], bg, c, accent)
	default:
		drawKey(label, x, y, bg, c, muted)
	}
}

func drawFaceBase(x, y float32, bg, accent rl.Color) {
	s := hintIcon
	rl.DrawCircle(int32(x+s/2), int32(y+s/2), s/2, bg)
	rl.DrawRing(rl.Vector2{X: x + s/2, Y: y + s/2}, s/2-1.5, s/2, 0, 360, 64, accent)
}
func strokeLine(x1, y1, x2, y2 float32, c rl.Color, t float32) {
	rl.DrawLineEx(rl.Vector2{X: x1, Y: y1}, rl.Vector2{X: x2, Y: y2}, t, c)
}
func drawDpad(x, y float32, bg, c rl.Color) {
	rl.DrawRectangleRounded(rl.Rectangle{X: x + 11, Y: y + 3, Width: 8, Height: 24}, .35, 8, bg)
	rl.DrawRectangleRounded(rl.Rectangle{X: x + 3, Y: y + 11, Width: 24, Height: 8}, .35, 8, bg)
	rl.DrawRectangleRoundedLines(rl.Rectangle{X: x + 11, Y: y + 3, Width: 8, Height: 24}, .35, 8, c)
	rl.DrawRectangleRoundedLines(rl.Rectangle{X: x + 3, Y: y + 11, Width: 24, Height: 8}, .35, 8, c)
}
func drawStick(x, y float32, l string, bg, c, accent rl.Color) {
	s := hintIcon
	rl.DrawCircle(int32(x+s/2), int32(y+s/2), s/2, bg)
	rl.DrawRing(rl.Vector2{X: x + s/2, Y: y + s/2}, s*.21, s*.31, 0, 360, 48, c)
	rl.DrawCircle(int32(x+s/2), int32(y+s/2), 3, accent)
	Text(l, x+10, y+6, 13, c)
}
func drawKey(label string, x, y float32, bg, c, muted rl.Color) {
	w := float32(Measure(label, 13)) + 16
	if w < 34 {
		w = 34
	}
	rl.DrawRectangleRounded(rl.Rectangle{X: x, Y: y + 3, Width: w, Height: 24}, .35, 10, bg)
	rl.DrawRectangleRoundedLines(rl.Rectangle{X: x, Y: y + 3, Width: w, Height: 24}, .35, 10, muted)
	Text(label, x+8, y+7, 13, c)
}
