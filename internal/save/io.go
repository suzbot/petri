package save

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// baseDirOverride allows tests to use a temporary directory
var baseDirOverride string

// SetBaseDir sets a custom base directory (for testing)
func SetBaseDir(dir string) {
	baseDirOverride = dir
}

// ResetBaseDir clears the base directory override
func ResetBaseDir() {
	baseDirOverride = ""
}

// WorldDir returns the directory for a specific world
func WorldDir(worldID string) (string, error) {
	baseDir, err := BaseDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(baseDir, "worlds", worldID), nil
}

// BaseDir returns the base petri directory (~/.petri)
func BaseDir() (string, error) {
	if baseDirOverride != "" {
		return baseDirOverride, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("could not determine home directory: %w", err)
	}
	return filepath.Join(home, ".petri"), nil
}

// EnsureWorldDir creates the world directory if it doesn't exist
func EnsureWorldDir(worldID string) (string, error) {
	dir, err := WorldDir(worldID)
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("could not create world directory: %w", err)
	}
	return dir, nil
}

// SaveWorld saves a world state to disk with backup rotation
func SaveWorld(worldID string, state *SaveState) error {
	dir, err := EnsureWorldDir(worldID)
	if err != nil {
		return err
	}

	statePath := filepath.Join(dir, "state.json")
	backupPath := filepath.Join(dir, "state.backup")

	// Rotate existing save to backup
	if _, err := os.Stat(statePath); err == nil {
		// Remove old backup if exists
		os.Remove(backupPath)
		// Move current state to backup
		if err := os.Rename(statePath, backupPath); err != nil {
			return fmt.Errorf("could not create backup: %w", err)
		}
	}

	// Marshal state to JSON
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("could not marshal state: %w", err)
	}

	// Write new state
	if err := os.WriteFile(statePath, data, 0644); err != nil {
		return fmt.Errorf("could not write state: %w", err)
	}

	return nil
}

// LoadWorld loads a world state from disk
func LoadWorld(worldID string) (*SaveState, error) {
	dir, err := WorldDir(worldID)
	if err != nil {
		return nil, err
	}

	statePath := filepath.Join(dir, "state.json")

	data, err := os.ReadFile(statePath)
	if err != nil {
		return nil, fmt.Errorf("could not read state: %w", err)
	}

	var state SaveState
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, fmt.Errorf("could not parse state: %w", err)
	}

	return &state, nil
}

// LoadWorldFromBackup loads a world state from the backup file
func LoadWorldFromBackup(worldID string) (*SaveState, error) {
	dir, err := WorldDir(worldID)
	if err != nil {
		return nil, err
	}

	backupPath := filepath.Join(dir, "state.backup")

	data, err := os.ReadFile(backupPath)
	if err != nil {
		return nil, fmt.Errorf("could not read backup: %w", err)
	}

	var state SaveState
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, fmt.Errorf("could not parse backup: %w", err)
	}

	return &state, nil
}

// WorldExists checks if a world with the given ID exists
func WorldExists(worldID string) bool {
	dir, err := WorldDir(worldID)
	if err != nil {
		return false
	}
	statePath := filepath.Join(dir, "state.json")
	_, err = os.Stat(statePath)
	return err == nil
}

// BackupExists checks if a backup exists for the given world
func BackupExists(worldID string) bool {
	dir, err := WorldDir(worldID)
	if err != nil {
		return false
	}
	backupPath := filepath.Join(dir, "state.backup")
	_, err = os.Stat(backupPath)
	return err == nil
}

// SaveMeta saves world metadata
func SaveMeta(worldID string, meta *WorldMeta) error {
	dir, err := EnsureWorldDir(worldID)
	if err != nil {
		return err
	}

	metaPath := filepath.Join(dir, "meta.json")

	data, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		return fmt.Errorf("could not marshal meta: %w", err)
	}

	if err := os.WriteFile(metaPath, data, 0644); err != nil {
		return fmt.Errorf("could not write meta: %w", err)
	}

	return nil
}

// LoadMeta loads world metadata
func LoadMeta(worldID string) (*WorldMeta, error) {
	dir, err := WorldDir(worldID)
	if err != nil {
		return nil, err
	}

	metaPath := filepath.Join(dir, "meta.json")

	data, err := os.ReadFile(metaPath)
	if err != nil {
		return nil, fmt.Errorf("could not read meta: %w", err)
	}

	var meta WorldMeta
	if err := json.Unmarshal(data, &meta); err != nil {
		return nil, fmt.Errorf("could not parse meta: %w", err)
	}

	return &meta, nil
}

// ListWorlds returns metadata for all saved worlds
func ListWorlds() ([]WorldMeta, error) {
	baseDir, err := BaseDir()
	if err != nil {
		return nil, err
	}

	worldsDir := filepath.Join(baseDir, "worlds")

	// Check if worlds directory exists
	if _, err := os.Stat(worldsDir); os.IsNotExist(err) {
		return nil, nil // No worlds yet
	}

	entries, err := os.ReadDir(worldsDir)
	if err != nil {
		return nil, fmt.Errorf("could not read worlds directory: %w", err)
	}

	var worlds []WorldMeta
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		meta, err := LoadMeta(entry.Name())
		if err != nil {
			// Skip worlds with missing/corrupt metadata
			continue
		}

		worlds = append(worlds, *meta)
	}

	return worlds, nil
}

// GenerateWorldID creates a new unique world ID
func GenerateWorldID() (string, error) {
	baseDir, err := BaseDir()
	if err != nil {
		return "", err
	}

	worldsDir := filepath.Join(baseDir, "worlds")

	// Find next available ID
	for i := 1; i <= 9999; i++ {
		id := fmt.Sprintf("world-%04d", i)
		dir := filepath.Join(worldsDir, id)
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			return id, nil
		}
	}

	return "", fmt.Errorf("too many worlds")
}

// GenerateWorldName creates a display name for a new world
func GenerateWorldName() (string, error) {
	worlds, err := ListWorlds()
	if err != nil {
		return "World 1", nil
	}
	return fmt.Sprintf("World %d", len(worlds)+1), nil
}

// CreateWorld creates a new world with initial metadata and returns its ID
func CreateWorld() (string, error) {
	worldID, err := GenerateWorldID()
	if err != nil {
		return "", fmt.Errorf("could not generate world ID: %w", err)
	}

	worldName, err := GenerateWorldName()
	if err != nil {
		return "", fmt.Errorf("could not generate world name: %w", err)
	}

	_, err = EnsureWorldDir(worldID)
	if err != nil {
		return "", fmt.Errorf("could not create world directory: %w", err)
	}

	now := time.Now()
	meta := &WorldMeta{
		ID:             worldID,
		Name:           worldName,
		CreatedAt:      now,
		LastPlayedAt:   now,
		CharacterCount: 0,
		AliveCount:     0,
	}

	if err := SaveMeta(worldID, meta); err != nil {
		return "", fmt.Errorf("could not save world metadata: %w", err)
	}

	return worldID, nil
}
