package app

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type AppState struct {
	Track    int     `json:"track"`
	Album    int     `json:"album"`
	Position float32 `json:"position"`
}

func statePath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".musa-state.json")
}

func loadState() AppState {
	s := AppState{Track: -1, Album: 0, Position: 0}
	b, err := os.ReadFile(statePath())
	if err != nil {
		return s
	}
	_ = json.Unmarshal(b, &s)
	return s
}

func saveState(s AppState) {
	b, _ := json.MarshalIndent(s, "", "  ")
	_ = os.WriteFile(statePath(), b, 0644)
}
