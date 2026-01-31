package save

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"petri/internal/types"
)

func setupTestDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	SetBaseDir(dir)
	t.Cleanup(func() {
		ResetBaseDir()
	})
	return dir
}

func TestCreateWorld(t *testing.T) {
	setupTestDir(t)

	worldID, err := CreateWorld()
	if err != nil {
		t.Fatalf("CreateWorld failed: %v", err)
	}

	if worldID == "" {
		t.Error("Expected non-empty world ID")
	}

	// Verify timestamp-based ID format (world-YYYYMMDD-HHMMSS = 21 chars)
	if len(worldID) != 21 || worldID[:6] != "world-" {
		t.Errorf("Expected timestamp-based ID format (21 chars), got '%s' (%d chars)", worldID, len(worldID))
	}

	// Verify directory was created
	dir, _ := WorldDir(worldID)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		t.Error("World directory was not created")
	}

	// Verify metadata was created
	meta, err := LoadMeta(worldID)
	if err != nil {
		t.Fatalf("Failed to load metadata: %v", err)
	}

	if meta.ID != worldID {
		t.Errorf("Expected meta.ID=%s, got %s", worldID, meta.ID)
	}
	if meta.Name == "" {
		t.Error("Expected non-empty name")
	}
}

func TestCreateWorld_MultipleWorlds(t *testing.T) {
	setupTestDir(t)

	// Create worlds with small delays to ensure unique timestamps
	world1, _ := CreateWorld()
	time.Sleep(1100 * time.Millisecond) // Ensure different second
	world2, _ := CreateWorld()
	time.Sleep(1100 * time.Millisecond)
	world3, _ := CreateWorld()

	if world1 == world2 || world2 == world3 {
		t.Errorf("World IDs should be unique: %s, %s, %s", world1, world2, world3)
	}

	// Verify all have valid metadata with timestamp-derived names
	meta1, _ := LoadMeta(world1)
	meta2, _ := LoadMeta(world2)
	meta3, _ := LoadMeta(world3)

	// All worlds should have non-empty names
	for i, meta := range []*WorldMeta{meta1, meta2, meta3} {
		if meta.Name == "" {
			t.Errorf("World %d: expected non-empty name", i+1)
		}
	}
}

func TestListWorlds_Empty(t *testing.T) {
	setupTestDir(t)

	worlds, err := ListWorlds()
	if err != nil {
		t.Fatalf("ListWorlds failed: %v", err)
	}

	if len(worlds) != 0 {
		t.Errorf("Expected 0 worlds, got %d", len(worlds))
	}
}

func TestListWorlds_WithWorlds(t *testing.T) {
	setupTestDir(t)

	CreateWorld()
	time.Sleep(1100 * time.Millisecond) // Ensure different timestamp
	CreateWorld()

	worlds, err := ListWorlds()
	if err != nil {
		t.Fatalf("ListWorlds failed: %v", err)
	}

	if len(worlds) != 2 {
		t.Errorf("Expected 2 worlds, got %d", len(worlds))
	}
}

func TestSaveAndLoadWorld(t *testing.T) {
	setupTestDir(t)

	worldID, _ := CreateWorld()

	state := &SaveState{
		Version:         1,
		SavedAt:         time.Now(),
		ElapsedGameTime: 123.456,
		MapWidth:        40,
		MapHeight:       30,
		Characters: []CharacterSave{
			{
				ID:       1,
				Name:     "TestChar",
				Position: types.Position{X: 10, Y: 15},
				Health:   100.0,
				Hunger:   80.0,
			},
		},
		Items: []ItemSave{
			{
				ID:       1,
				Position: types.Position{X: 5, Y: 5},
				ItemType: "mushroom",
				Color:    "red",
			},
		},
	}

	if err := SaveWorld(worldID, state); err != nil {
		t.Fatalf("SaveWorld failed: %v", err)
	}

	loaded, err := LoadWorld(worldID)
	if err != nil {
		t.Fatalf("LoadWorld failed: %v", err)
	}

	if loaded.ElapsedGameTime != 123.456 {
		t.Errorf("Expected elapsed time 123.456, got %f", loaded.ElapsedGameTime)
	}
	if loaded.MapWidth != 40 {
		t.Errorf("Expected width 40, got %d", loaded.MapWidth)
	}
	if len(loaded.Characters) != 1 {
		t.Fatalf("Expected 1 character, got %d", len(loaded.Characters))
	}
	if loaded.Characters[0].Name != "TestChar" {
		t.Errorf("Expected name 'TestChar', got '%s'", loaded.Characters[0].Name)
	}
}

func TestBackupRotation(t *testing.T) {
	setupTestDir(t)

	worldID, _ := CreateWorld()

	// First save
	state1 := &SaveState{
		Version:         1,
		ElapsedGameTime: 100.0,
	}
	SaveWorld(worldID, state1)

	// Verify no backup yet (first save)
	if BackupExists(worldID) {
		t.Error("Backup should not exist after first save")
	}

	// Second save should create backup
	state2 := &SaveState{
		Version:         1,
		ElapsedGameTime: 200.0,
	}
	SaveWorld(worldID, state2)

	if !BackupExists(worldID) {
		t.Error("Backup should exist after second save")
	}

	// Main save should have new data
	loaded, _ := LoadWorld(worldID)
	if loaded.ElapsedGameTime != 200.0 {
		t.Errorf("Expected 200.0, got %f", loaded.ElapsedGameTime)
	}

	// Backup should have old data
	backup, _ := LoadWorldFromBackup(worldID)
	if backup.ElapsedGameTime != 100.0 {
		t.Errorf("Expected backup to have 100.0, got %f", backup.ElapsedGameTime)
	}
}

func TestWorldExists(t *testing.T) {
	setupTestDir(t)

	if WorldExists("nonexistent") {
		t.Error("WorldExists should return false for nonexistent world")
	}

	worldID, _ := CreateWorld()

	// World exists after creation but has no state yet
	if WorldExists(worldID) {
		t.Error("WorldExists should return false before state is saved")
	}

	// Save state
	SaveWorld(worldID, &SaveState{Version: 1})

	if !WorldExists(worldID) {
		t.Error("WorldExists should return true after state is saved")
	}
}

func TestSaveAndLoadMeta(t *testing.T) {
	setupTestDir(t)

	worldID, _ := CreateWorld()

	now := time.Now()
	meta := &WorldMeta{
		ID:             worldID,
		Name:           "Custom Name",
		CreatedAt:      now,
		LastPlayedAt:   now,
		CharacterCount: 5,
		AliveCount:     3,
	}

	if err := SaveMeta(worldID, meta); err != nil {
		t.Fatalf("SaveMeta failed: %v", err)
	}

	loaded, err := LoadMeta(worldID)
	if err != nil {
		t.Fatalf("LoadMeta failed: %v", err)
	}

	if loaded.Name != "Custom Name" {
		t.Errorf("Expected name 'Custom Name', got '%s'", loaded.Name)
	}
	if loaded.CharacterCount != 5 {
		t.Errorf("Expected 5 characters, got %d", loaded.CharacterCount)
	}
	if loaded.AliveCount != 3 {
		t.Errorf("Expected 3 alive, got %d", loaded.AliveCount)
	}
}

func TestGenerateWorldID_TimestampFormat(t *testing.T) {
	setupTestDir(t)

	id, err := GenerateWorldID()
	if err != nil {
		t.Fatalf("GenerateWorldID failed: %v", err)
	}

	// Should be in format "world-YYYYMMDD-HHMMSS" (21 chars)
	if len(id) != 21 {
		t.Errorf("Expected ID length 21, got %d: '%s'", len(id), id)
	}

	if id[:6] != "world-" {
		t.Errorf("Expected ID to start with 'world-', got '%s'", id)
	}

	// Parse the timestamp part to verify format
	timestampPart := id[6:] // Skip "world-"
	_, err = time.Parse("20060102-150405", timestampPart)
	if err != nil {
		t.Errorf("Expected valid timestamp format, got '%s': %v", timestampPart, err)
	}
}

func TestLoadWorld_NotFound(t *testing.T) {
	setupTestDir(t)

	_, err := LoadWorld("nonexistent")
	if err == nil {
		t.Error("Expected error for nonexistent world")
	}
}

func TestLoadMeta_NotFound(t *testing.T) {
	setupTestDir(t)

	_, err := LoadMeta("nonexistent")
	if err == nil {
		t.Error("Expected error for nonexistent metadata")
	}
}

func TestBaseDir_Override(t *testing.T) {
	customDir := "/tmp/petri-test-custom"
	SetBaseDir(customDir)
	defer ResetBaseDir()

	dir, err := BaseDir()
	if err != nil {
		t.Fatalf("BaseDir failed: %v", err)
	}

	if dir != customDir {
		t.Errorf("Expected '%s', got '%s'", customDir, dir)
	}
}

func TestWorldDir(t *testing.T) {
	dir := setupTestDir(t)

	worldDir, err := WorldDir("test-world")
	if err != nil {
		t.Fatalf("WorldDir failed: %v", err)
	}

	expected := filepath.Join(dir, "worlds", "test-world")
	if worldDir != expected {
		t.Errorf("Expected '%s', got '%s'", expected, worldDir)
	}
}

func TestDeleteWorld(t *testing.T) {
	setupTestDir(t)

	// Create a world with state and meta
	worldID, err := CreateWorld()
	if err != nil {
		t.Fatalf("CreateWorld failed: %v", err)
	}

	// Save some state
	state := &SaveState{Version: 1, ElapsedGameTime: 100.0}
	if err := SaveWorld(worldID, state); err != nil {
		t.Fatalf("SaveWorld failed: %v", err)
	}

	// Verify world exists
	if !WorldExists(worldID) {
		t.Fatal("World should exist before delete")
	}

	// Delete the world
	if err := DeleteWorld(worldID); err != nil {
		t.Fatalf("DeleteWorld failed: %v", err)
	}

	// Verify world no longer exists
	if WorldExists(worldID) {
		t.Error("World should not exist after delete")
	}

	// Verify directory is gone
	dir, _ := WorldDir(worldID)
	if _, err := os.Stat(dir); !os.IsNotExist(err) {
		t.Error("World directory should be removed")
	}

	// Verify world is not in list
	worlds, _ := ListWorlds()
	for _, w := range worlds {
		if w.ID == worldID {
			t.Error("Deleted world should not appear in list")
		}
	}
}

func TestListWorlds_CleansUpGhostDirectories(t *testing.T) {
	baseDir := setupTestDir(t)

	// Create a valid world
	validWorldID, _ := CreateWorld()
	SaveWorld(validWorldID, &SaveState{Version: 1})

	// Manually create a ghost directory (has state.json but no meta.json)
	ghostDir := filepath.Join(baseDir, "worlds", "ghost-world")
	if err := os.MkdirAll(ghostDir, 0755); err != nil {
		t.Fatalf("Failed to create ghost dir: %v", err)
	}
	ghostState := filepath.Join(ghostDir, "state.json")
	if err := os.WriteFile(ghostState, []byte(`{"version":1}`), 0644); err != nil {
		t.Fatalf("Failed to create ghost state: %v", err)
	}

	// Verify ghost directory exists
	if _, err := os.Stat(ghostDir); os.IsNotExist(err) {
		t.Fatal("Ghost directory should exist before ListWorlds")
	}

	// List worlds - should clean up ghost
	worlds, err := ListWorlds()
	if err != nil {
		t.Fatalf("ListWorlds failed: %v", err)
	}

	// Should only have the valid world
	if len(worlds) != 1 {
		t.Errorf("Expected 1 world, got %d", len(worlds))
	}
	if len(worlds) > 0 && worlds[0].ID != validWorldID {
		t.Errorf("Expected valid world ID %s, got %s", validWorldID, worlds[0].ID)
	}

	// Ghost directory should be cleaned up
	if _, err := os.Stat(ghostDir); !os.IsNotExist(err) {
		t.Error("Ghost directory should be removed by ListWorlds")
	}
}
