package game

import (
	"testing"

	"petri/internal/config"
	"petri/internal/types"
)

// =============================================================================
// Pond Generation Tests
// =============================================================================

func TestSpawnPonds_WaterTileCountInRange(t *testing.T) {
	t.Parallel()

	// Run multiple times to exercise randomness
	for i := 0; i < 20; i++ {
		m := NewMap(config.MapWidth, config.MapHeight)
		SpawnPonds(m)

		// Count pond tiles only (exclude springs)
		pondCount := 0
		for _, pos := range m.WaterPositions() {
			if m.WaterAt(pos) == WaterPond {
				pondCount++
			}
		}

		minTiles := config.PondMinCount * config.PondMinSize // 1*4 = 4
		maxTiles := config.PondMaxCount * config.PondMaxSize // 5*16 = 80
		if pondCount < minTiles || pondCount > maxTiles {
			t.Errorf("iteration %d: pond tile count %d outside valid range [%d, %d]",
				i, pondCount, minTiles, maxTiles)
		}
	}
}

func TestSpawnPonds_NoUndersizedBlobs(t *testing.T) {
	t.Parallel()

	for i := 0; i < 20; i++ {
		m := NewMap(config.MapWidth, config.MapHeight)
		SpawnPonds(m)

		// Flood-fill all pond tiles into connected components
		components := findPondComponents(m)
		for j, comp := range components {
			if len(comp) < config.PondMinSize {
				t.Errorf("iteration %d: pond component %d has %d tiles, minimum is %d",
					i, j, len(comp), config.PondMinSize)
			}
		}
	}
}

// =============================================================================
// Map Connectivity Tests
// =============================================================================

func TestIsMapConnected_EmptyMap(t *testing.T) {
	t.Parallel()

	m := NewMap(20, 20)
	if !isMapConnected(m) {
		t.Error("isMapConnected() should return true for empty map")
	}
}

func TestIsMapConnected_WaterDoesNotPartition(t *testing.T) {
	t.Parallel()

	m := NewMap(20, 20)
	// Small pond in the middle — doesn't partition
	m.AddWater(types.Position{X: 10, Y: 10}, WaterPond)
	m.AddWater(types.Position{X: 11, Y: 10}, WaterPond)
	m.AddWater(types.Position{X: 10, Y: 11}, WaterPond)
	m.AddWater(types.Position{X: 11, Y: 11}, WaterPond)

	if !isMapConnected(m) {
		t.Error("isMapConnected() should return true when water doesn't partition map")
	}
}

func TestIsMapConnected_WaterWallPartitions(t *testing.T) {
	t.Parallel()

	m := NewMap(20, 20)
	// Water wall across full width at y=10, splitting map in two
	for x := 0; x < 20; x++ {
		m.AddWater(types.Position{X: x, Y: 10}, WaterPond)
	}

	if isMapConnected(m) {
		t.Error("isMapConnected() should return false when water wall partitions map")
	}
}

// =============================================================================
// Test Helpers
// =============================================================================

// findPondComponents flood-fills all pond tiles and returns connected components
func findPondComponents(m *Map) [][]types.Position {
	visited := make(map[types.Position]bool)
	var components [][]types.Position

	for _, pos := range m.WaterPositions() {
		if m.WaterAt(pos) != WaterPond {
			continue
		}
		if visited[pos] {
			continue
		}

		// BFS from this pond tile
		var component []types.Position
		queue := []types.Position{pos}
		visited[pos] = true

		for len(queue) > 0 {
			cur := queue[0]
			queue = queue[1:]
			component = append(component, cur)

			// Check cardinal neighbors
			for _, dir := range [][2]int{{0, -1}, {1, 0}, {0, 1}, {-1, 0}} {
				neighbor := types.Position{X: cur.X + dir[0], Y: cur.Y + dir[1]}
				if !visited[neighbor] && m.WaterAt(neighbor) == WaterPond {
					visited[neighbor] = true
					queue = append(queue, neighbor)
				}
			}
		}

		components = append(components, component)
	}

	return components
}
