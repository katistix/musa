package app

import (
	"strings"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type Controller struct {
	Connected bool
	Index     int32
	Name      string
	DualShock bool
}

func DetectController() Controller {
	for i := int32(0); i < 4; i++ {
		if rl.IsGamepadAvailable(i) {
			name := rl.GetGamepadName(i)
			low := strings.ToLower(name)
			return Controller{Connected: true, Index: i, Name: name, DualShock: strings.Contains(low, "dualshock") || strings.Contains(low, "wireless controller") || strings.Contains(low, "playstation") || strings.Contains(low, "ps4")}
		}
	}
	return Controller{}
}

func padPressed(button int32) bool {
	return rl.IsGamepadAvailable(0) && rl.IsGamepadButtonPressed(0, button)
}

func padAxis(axis int32) float32 {
	if !rl.IsGamepadAvailable(0) {
		return 0
	}
	v := rl.GetGamepadAxisMovement(0, axis)
	if v > -.35 && v < .35 {
		return 0
	}
	return v
}
