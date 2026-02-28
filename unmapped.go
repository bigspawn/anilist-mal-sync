package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Sync direction constants (string values for JSON serialization).
const (
	DirectionForwardStr = "forward" // AniList → MAL
	DirectionReverseStr = "reverse" // MAL → AniList
)

// UnmappedEntry represents an entry that could not be mapped during sync.
type UnmappedEntry struct {
	AniListID int    `json:"anilist_id"`
	MALID     int    `json:"mal_id,omitempty"`
	Title     string `json:"title"`
	MediaType string `json:"media_type"`       // "anime" or "manga"
	Direction string `json:"direction"`        // "forward" or "reverse"
	Reason    string `json:"reason,omitempty"` // why it was unmapped
}

// UnmappedState holds the list of unmapped entries from the last sync.
type UnmappedState struct {
	Entries   []UnmappedEntry `json:"entries"`
	UpdatedAt time.Time       `json:"updated_at"`
}

// LoadUnmappedState loads unmapped state from a JSON file.
func LoadUnmappedState(path string) (*UnmappedState, error) {
	if path == "" {
		path = getDefaultUnmappedPath()
	}

	data, err := os.ReadFile(path) // #nosec G304 - Path from user config
	if err != nil {
		if os.IsNotExist(err) {
			return &UnmappedState{}, nil
		}
		return nil, fmt.Errorf("read unmapped state: %w", err)
	}

	var state UnmappedState
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, fmt.Errorf("parse unmapped state: %w", err)
	}

	return &state, nil
}

// Save writes the unmapped state to a JSON file.
func (s *UnmappedState) Save(path string) error {
	if path == "" {
		path = getDefaultUnmappedPath()
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o750); err != nil { // #nosec G301 - State directory
		return fmt.Errorf("create unmapped state directory: %w", err)
	}

	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal unmapped state: %w", err)
	}

	if err := os.WriteFile(path, data, 0o600); err != nil {
		return fmt.Errorf("write unmapped state: %w", err)
	}

	return nil
}

func getDefaultUnmappedPath() string {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return ""
	}
	return filepath.Join(configDir, "anilist-mal-sync", "state", "unmapped.json")
}
