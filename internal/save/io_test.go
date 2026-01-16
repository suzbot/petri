package save

import (
	"os"
	"path/filepath"
	"testing"
	"time"
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
	if meta.Name != "World 1" {
		t.Errorf("Expected name 'World 1', got '%s'", meta.Name)
	}
}

func TestCreateWorld_MultipleWorlds(t *testing.T) {
	setupTestDir(t)

	world1, _ := CreateWorld()
	world2, _ := CreateWorld()
	world3, _ := CreateWorld()

	if world1 == world2 || world2 == world3 {
		t.Error("World IDs should be unique")
	}

	// Verify names increment
	meta1, _ := LoadMeta(world1)
	meta2, _ := LoadMeta(world2)
	meta3, _ := LoadMeta(world3)

	if meta1.Name != "World 1" {
		t.Errorf("Expected 'World 1', got '%s'", meta1.Name)
	}
	if meta2.Name != "World 2" {
		t.Errorf("Expected 'World 2', got '%s'", meta2.Name)
	}
	if meta3.Name != "World 3" {
		t.Errorf("Expected 'World 3', got '%s'", meta3.Name)
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
				ID:     1,
				Name:   "TestChar",
				X:      10,
				Y:      15,
				Health: 100.0,
				Hunger: 80.0,
			},
		},
		Items: []ItemSave{
			{
				ID:       1,
				X:        5,
				Y:        5,
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

func TestGenerateWorldID_Sequential(t *testing.T) {
	setupTestDir(t)

	id1, _ := GenerateWorldID()
	EnsureWorldDir(id1) // Create the directory

	id2, _ := GenerateWorldID()
	EnsureWorldDir(id2)

	id3, _ := GenerateWorldID()

	if id1 != "world-0001" {
		t.Errorf("Expected 'world-0001', got '%s'", id1)
	}
	if id2 != "world-0002" {
		t.Errorf("Expected 'world-0002', got '%s'", id2)
	}
	if id3 != "world-0003" {
		t.Errorf("Expected 'world-0003', got '%s'", id3)
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
