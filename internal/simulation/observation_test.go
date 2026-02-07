package simulation

import (
	"fmt"
	"testing"

	"petri/internal/config"
	"petri/internal/entity"
	"petri/internal/system"
)

// TestObserveBalanceMetrics runs simulations and collects balance metrics
func TestObserveBalanceMetrics(t *testing.T) {
	const (
		numRuns       = 5
		ticksPerRun   = 16000 // ~2400 game-seconds (~20 world days) at 0.15s/tick
		sampleEvery   = 1000  // sample metrics every N ticks
		delta         = 0.15  // tick duration
	)

	fmt.Println("\n=== BALANCE OBSERVATION REPORT ===")
	fmt.Printf("Running %d simulations, %d ticks each (%.0f game-seconds)\n\n",
		numRuns, ticksPerRun, float64(ticksPerRun)*delta)

	var totalDeaths, totalSurvivors int
	var totalEdibleEnd, totalFlowerEnd int
	var totalPreferences int
	deathCauses := make(map[string]int)
	moodAtEnd := make(map[string]int)

	for run := 0; run < numRuns; run++ {
		world := CreateTestWorld(WorldOptions{})
		chars := world.GameMap.Characters()
		startChars := len(chars)

		fmt.Printf("--- Run %d ---\n", run+1)
		fmt.Printf("Start: %d characters, %d items\n", startChars, len(world.GameMap.Items()))

		// Track item counts over time
		var edibleCounts, flowerCounts []int

		for tick := 0; tick < ticksPerRun; tick++ {
			RunTick(world, delta)

			// Sample periodically
			if tick%sampleEvery == 0 {
				edible, flowers := countItems(world)
				edibleCounts = append(edibleCounts, edible)
				flowerCounts = append(flowerCounts, flowers)
			}
		}

		// End-of-run stats
		chars = world.GameMap.Characters()
		alive := 0
		for _, c := range chars {
			if !c.IsDead {
				alive++
				moodAtEnd[moodTier(c.Mood)]++
				totalPreferences += len(c.Preferences)
			} else {
				deathCauses[inferDeathCause(c)]++
			}
		}
		deaths := startChars - alive

		edibleEnd, flowerEnd := countItems(world)

		fmt.Printf("End: %d alive, %d dead\n", alive, deaths)
		fmt.Printf("Items: %d edible, %d flowers\n", edibleEnd, flowerEnd)
		fmt.Printf("Item trend - Edible: %d→%d, Flowers: %d→%d\n",
			edibleCounts[0], edibleEnd, flowerCounts[0], flowerEnd)

		totalDeaths += deaths
		totalSurvivors += alive
		totalEdibleEnd += edibleEnd
		totalFlowerEnd += flowerEnd
	}

	fmt.Println("\n=== AGGREGATE RESULTS ===")
	fmt.Printf("Total characters: %d (%d survived, %d died)\n",
		totalSurvivors+totalDeaths, totalSurvivors, totalDeaths)
	fmt.Printf("Survival rate: %.1f%%\n", float64(totalSurvivors)/float64(totalSurvivors+totalDeaths)*100)
	fmt.Printf("Avg end items: %.1f edible, %.1f flowers\n",
		float64(totalEdibleEnd)/float64(numRuns), float64(totalFlowerEnd)/float64(numRuns))
	fmt.Printf("Avg preferences formed per survivor: %.1f\n",
		float64(totalPreferences)/float64(totalSurvivors))

	fmt.Println("\nDeath causes:")
	for cause, count := range deathCauses {
		fmt.Printf("  %s: %d\n", cause, count)
	}

	fmt.Println("\nSurvivor mood distribution:")
	for mood, count := range moodAtEnd {
		fmt.Printf("  %s: %d\n", mood, count)
	}
}

// TestObserveFoodScarcity specifically watches food availability
func TestObserveFoodScarcity(t *testing.T) {
	const (
		ticks = 16000 // ~2400 game-seconds (~20 world days)
		delta = 0.15
	)

	fmt.Println("\n=== FOOD SCARCITY OBSERVATION ===")

	world := CreateTestWorld(WorldOptions{})

	var eatEvents, spawnEvents int
	initialEdible, _ := countItems(world)

	fmt.Printf("Initial edible items: %d\n", initialEdible)
	fmt.Printf("Characters: %d\n", len(world.GameMap.Characters()))
	fmt.Printf("Spawn chance: %.0f%%, Interval base: %.1fs\n",
		config.ItemSpawnChance*100, config.ItemLifecycle["berry"].SpawnInterval)

	// Run simulation tracking consumption - checkpoints every 5 world days
	checkpoints := []int{4000, 8000, 12000, 16000} // 5, 10, 15, 20 world days
	checkIdx := 0

	for tick := 0; tick < ticks; tick++ {
		preItems := len(world.GameMap.Items())
		RunTick(world, delta)
		postItems := len(world.GameMap.Items())

		// Approximate: if items decreased, something was eaten
		if postItems < preItems {
			eatEvents++
		} else if postItems > preItems {
			spawnEvents += (postItems - preItems)
		}

		if checkIdx < len(checkpoints) && tick == checkpoints[checkIdx] {
			edible, flowers := countItems(world)
			aliveCount := countAliveChars(world)
			worldDays := float64(tick) * delta / 120
			fmt.Printf("Day %.0f (tick %d): %d edible, %d flowers, %d alive\n",
				worldDays, tick, edible, flowers, aliveCount)
			checkIdx++
		}
	}

	edibleEnd, flowersEnd := countItems(world)
	fmt.Printf("\nFinal (day 20): %d edible, %d flowers\n", edibleEnd, flowersEnd)
	fmt.Printf("Approximate eat events: %d, spawn events: %d\n", eatEvents, spawnEvents)
	fmt.Printf("Net item change: %+d\n", (edibleEnd+flowersEnd)-initialEdible)
}

// TestObserveFlowerGrowth tracks flower population specifically
func TestObserveFlowerGrowth(t *testing.T) {
	const (
		ticks = 16000 // ~2400 game-seconds (~20 world days)
		delta = 0.15
	)

	fmt.Println("\n=== FLOWER GROWTH OBSERVATION ===")
	fmt.Printf("Flower spawn interval: %.0fs, death interval: %.0fs\n",
		config.ItemLifecycle["flower"].SpawnInterval,
		config.ItemLifecycle["flower"].DeathInterval)
	fmt.Printf("(Note: intervals multiplied by item count ~20 in lifecycle.go)\n")

	world := CreateTestWorld(WorldOptions{})

	_, initialFlowers := countItems(world)
	initialEdible, _ := countItems(world)
	prevFlowers := initialFlowers

	var totalSpawns, totalDeaths int

	fmt.Printf("Initial: %d edible, %d flowers\n\n", initialEdible, initialFlowers)

	// Checkpoints every 5 world days
	checkpoints := []int{4000, 8000, 12000, 16000}
	checkIdx := 0

	for tick := 0; tick < ticks; tick++ {
		RunTick(world, delta)

		// Track flower changes
		_, flowers := countItems(world)
		if flowers > prevFlowers {
			totalSpawns += flowers - prevFlowers
		} else if flowers < prevFlowers {
			totalDeaths += prevFlowers - flowers
		}
		prevFlowers = flowers

		if checkIdx < len(checkpoints) && tick == checkpoints[checkIdx] {
			edible, flowers := countItems(world)
			worldDays := float64(tick) * delta / 120
			ratio := float64(flowers) / float64(edible+flowers) * 100
			fmt.Printf("Day %.0f: %d edible, %d flowers (%.1f%% flowers)\n",
				worldDays, edible, flowers, ratio)
			checkIdx++
		}
	}

	edibleEnd, flowersEnd := countItems(world)
	ratio := float64(flowersEnd) / float64(edibleEnd+flowersEnd) * 100
	fmt.Printf("\nFinal: %d edible, %d flowers (%.1f%% flowers)\n",
		edibleEnd, flowersEnd, ratio)
	fmt.Printf("Flower lifecycle: %d spawns, %d deaths (net: %+d)\n",
		totalSpawns, totalDeaths, flowersEnd-initialFlowers)

	if flowersEnd > initialFlowers*3 {
		fmt.Println("⚠️  WARNING: Flower overpopulation detected!")
	}
}

func countItems(world *TestWorld) (edible, flowers int) {
	for _, item := range world.GameMap.Items() {
		if item.ItemType == "flower" {
			flowers++
		} else {
			edible++
		}
	}
	return
}

func countAliveChars(world *TestWorld) int {
	count := 0
	for _, c := range world.GameMap.Characters() {
		if !c.IsDead {
			count++
		}
	}
	return count
}

func inferDeathCause(c *entity.Character) string {
	if c.Health > 0 {
		return "unknown"
	}
	if c.Hunger >= 100 {
		return "starvation"
	}
	if c.Thirst >= 100 {
		return "dehydration"
	}
	if c.Poisoned {
		return "poison"
	}
	return "unknown"
}

func moodTier(mood float64) string {
	switch {
	case mood <= 20:
		return "Miserable"
	case mood <= 40:
		return "Unhappy"
	case mood <= 60:
		return "Neutral"
	case mood <= 80:
		return "Happy"
	default:
		return "Joyful"
	}
}


func actionName(a entity.ActionType) string {
	switch a {
	case entity.ActionNone:
		return "None"
	case entity.ActionMove:
		return "Move"
	case entity.ActionConsume:
		return "Consume"
	case entity.ActionDrink:
		return "Drink"
	case entity.ActionSleep:
		return "Sleep"
	case entity.ActionLook:
		return "Look"
	case entity.ActionTalk:
		return "Talk"
	case entity.ActionPickup:
		return "Pickup"
	default:
		return "Unknown"
	}
}

// TestObserveTimeToFirstDeath measures how long until characters start dying
func TestObserveTimeToFirstDeath(t *testing.T) {
	const (
		numRuns  = 10
		maxTicks = 20000 // ~3000 game-seconds (~25 world days)
		delta    = 0.15
	)

	fmt.Println("\n=== TIME TO DEATH OBSERVATION ===")
	fmt.Printf("Running %d simulations, max %d ticks (%.0f game-seconds, ~%.1f world days)\n\n",
		numRuns, maxTicks, float64(maxTicks)*delta, float64(maxTicks)*delta/120)

	var deathTicks []int
	var deathCauses []string
	var noDeathRuns int

	for run := 0; run < numRuns; run++ {
		world := CreateTestWorld(WorldOptions{})
		initialAlive := countAliveChars(world)
		deathTick := -1
		cause := ""

		for tick := 0; tick < maxTicks; tick++ {
			RunTick(world, delta)

			currentAlive := countAliveChars(world)
			if currentAlive < initialAlive {
				deathTick = tick
				// Find the dead character
				for _, c := range world.GameMap.Characters() {
					if c.IsDead {
						cause = inferDeathCause(c)
						break
					}
				}
				break
			}
		}

		if deathTick >= 0 {
			gameSeconds := float64(deathTick) * delta
			// Find the dead character and print detailed diagnostics
			for _, c := range world.GameMap.Characters() {
				if c.IsDead {
					edible, _ := countItems(world)
					foodResult := system.FindFoodTarget(c, world.GameMap.Items())
					foodInfo := "none found"
					if foodResult.Item != nil {
						ipos := foodResult.Item.Pos()
						cpos := c.Pos()
						dist := cpos.DistanceTo(ipos)
						foodInfo = fmt.Sprintf("dist:%d pref:%d", dist, foodResult.NetPreference)
					}
					fmt.Printf("Run %d: Death at tick %d (%.0fs) - %s | Speed:%d Poisoned:%v Food:%s Edible:%d\n",
						run+1, deathTick, gameSeconds, cause, c.EffectiveSpeed(), c.Poisoned, foodInfo, edible)
					break
				}
			}
			deathTicks = append(deathTicks, deathTick)
			deathCauses = append(deathCauses, cause)
		} else {
			fmt.Printf("Run %d: No deaths in %d ticks\n", run+1, maxTicks)
			noDeathRuns++
		}
	}

	fmt.Println("\n=== SUMMARY ===")
	if len(deathTicks) > 0 {
		minTick, maxTick, sum := deathTicks[0], deathTicks[0], 0
		for _, t := range deathTicks {
			sum += t
			if t < minTick {
				minTick = t
			}
			if t > maxTick {
				maxTick = t
			}
		}
		avgTick := float64(sum) / float64(len(deathTicks))

		fmt.Printf("Deaths occurred: %d/%d runs\n", len(deathTicks), numRuns)
		fmt.Printf("Time to first death:\n")
		fmt.Printf("  Min: tick %d (%.0fs)\n", minTick, float64(minTick)*delta)
		fmt.Printf("  Max: tick %d (%.0fs)\n", maxTick, float64(maxTick)*delta)
		fmt.Printf("  Avg: tick %.0f (%.0fs)\n", avgTick, avgTick*delta)

		// Count causes
		causeCounts := make(map[string]int)
		for _, c := range deathCauses {
			causeCounts[c]++
		}
		fmt.Println("Death causes:")
		for cause, count := range causeCounts {
			fmt.Printf("  %s: %d\n", cause, count)
		}
	} else {
		fmt.Println("No deaths occurred in any run!")
	}
	fmt.Printf("Runs with no deaths: %d/%d\n", noDeathRuns, numRuns)
}

// TestObserveGourdReproduction verifies gourds are spawning and reproducing
func TestObserveGourdReproduction(t *testing.T) {
	const (
		ticks = 10000 // ~1500 game-seconds - enough for multiple spawn cycles
		delta = 0.15
	)

	fmt.Println("\n=== GOURD REPRODUCTION OBSERVATION ===")
	fmt.Printf("Running simulation for %d ticks (%.0f game-seconds) with NO characters\n",
		ticks, float64(ticks)*delta)
	fmt.Printf("Spawn interval: %.0fs (~%d ticks)\n\n",
		config.ItemLifecycle["gourd"].SpawnInterval,
		int(config.ItemLifecycle["gourd"].SpawnInterval/delta))

	// Create world without characters so items don't get eaten
	world := CreateTestWorld(WorldOptions{NoCharacters: true})

	countByType := func() map[string]int {
		counts := make(map[string]int)
		for _, item := range world.GameMap.Items() {
			counts[item.ItemType]++
		}
		return counts
	}

	initial := countByType()
	fmt.Println("Initial item counts:")
	for itemType, count := range initial {
		fmt.Printf("  %s: %d\n", itemType, count)
	}

	// Run simulation
	samplePoints := []int{2500, 5000, 7500, 10000}
	sampleIdx := 0

	for tick := 0; tick < ticks; tick++ {
		RunTick(world, delta)

		if sampleIdx < len(samplePoints) && tick == samplePoints[sampleIdx] {
			counts := countByType()
			fmt.Printf("\nTick %d (%.0fs):\n", tick, float64(tick)*delta)
			for itemType, count := range counts {
				change := count - initial[itemType]
				sign := ""
				if change > 0 {
					sign = "+"
				}
				fmt.Printf("  %s: %d (%s%d)\n", itemType, count, sign, change)
			}
			sampleIdx++
		}
	}

	final := countByType()
	fmt.Println("\n=== REPRODUCTION SUMMARY ===")

	gourdInitial := initial["gourd"]
	gourdFinal := final["gourd"]
	fmt.Printf("Gourds: %d → %d (change: %+d)\n", gourdInitial, gourdFinal, gourdFinal-gourdInitial)

	if gourdFinal <= gourdInitial {
		fmt.Println("⚠️  WARNING: Gourds did not reproduce!")
	} else {
		fmt.Println("✓ Gourds are reproducing")
	}
}

// TestObserveDeathProgression tracks all deaths over extended simulation
func TestObserveDeathProgression(t *testing.T) {
	const (
		ticks = 20000 // ~3000 game-seconds (~25 world days)
		delta = 0.15
	)

	fmt.Println("\n=== DEATH PROGRESSION OBSERVATION ===")
	fmt.Printf("Running single simulation for %d ticks (%.0f game-seconds, ~%.1f world days)\n\n",
		ticks, float64(ticks)*delta, float64(ticks)*delta/120)

	world := CreateTestWorld(WorldOptions{})
	initialChars := len(world.GameMap.Characters())

	type deathRecord struct {
		tick  int
		name  string
		cause string
	}
	var deaths []deathRecord
	deadSet := make(map[int]bool)

	// Track first character's state for diagnosis
	chars := world.GameMap.Characters()
	trackChar := chars[0]
	fmt.Printf("Tracking %s - initial Hunger: %.0f, Prefs: %d\n\n",
		trackChar.Name, trackChar.Hunger, len(trackChar.Preferences))

	checkpoints := []int{2000, 4000, 6000, 8000, 10000, 12000, 15000, 20000}
	checkIdx := 0
	diagCheckpoints := []int{500, 1000, 1500, 2000, 2500, 3000, 3500, 4000}

	diagIdx := 0
	for tick := 0; tick < ticks; tick++ {
		RunTick(world, delta)

		// Diagnostic: track one character's state
		if diagIdx < len(diagCheckpoints) && tick == diagCheckpoints[diagIdx] && !trackChar.IsDead {
			activity := trackChar.CurrentActivity
			intent := "none"
			if trackChar.Intent != nil {
				intent = actionName(trackChar.Intent.Action)
				if trackChar.Intent.TargetItem != nil {
					intent += " -> " + trackChar.Intent.TargetItem.Description()
				} else if trackChar.Intent.TargetWaterPos != nil {
					intent += " -> water"
				} else if trackChar.Intent.TargetFeature != nil {
					intent += " -> feature"
				}
			}
			// Check what food would be found
			foodResult := system.FindFoodTarget(trackChar, world.GameMap.Items())
			foodInfo := "none"
			if foodResult.Item != nil {
				foodInfo = fmt.Sprintf("%s (pref:%d)", foodResult.Item.Description(), foodResult.NetPreference)
			}
			fmt.Printf("  [Diag %s tick %d] Hunger:%.0f Health:%.0f Poisoned:%v Activity:%s Intent:%s BestFood:%s\n",
				trackChar.Name, tick, trackChar.Hunger, trackChar.Health, trackChar.Poisoned, activity, intent, foodInfo)
			diagIdx++
		}

		// Check for new deaths
		for _, c := range world.GameMap.Characters() {
			if c.IsDead && !deadSet[c.ID] {
				deadSet[c.ID] = true
				deaths = append(deaths, deathRecord{
					tick:  tick,
					name:  c.Name,
					cause: inferDeathCause(c),
				})
				// Detailed death diagnostics
				edible, flowers := countItems(world)
				foodResult := system.FindFoodTarget(c, world.GameMap.Items())
				foodInfo := "none found"
				if foodResult.Item != nil {
					ipos := foodResult.Item.Pos()
					cpos := c.Pos()
					dist := cpos.DistanceTo(ipos)
					foodInfo = fmt.Sprintf("%s (pref:%d dist:%d)", foodResult.Item.Description(), foodResult.NetPreference, dist)
				}
				fmt.Printf("Tick %d (%.0fs): %s died - %s\n", tick, float64(tick)*delta, c.Name, inferDeathCause(c))
				fmt.Printf("  Stats: Hunger:%.0f Health:%.0f Speed:%d Poisoned:%v Frustrated:%v\n",
					c.Hunger, c.Health, c.EffectiveSpeed(), c.Poisoned, c.IsFrustrated)
				fmt.Printf("  World: %d edible items, %d flowers\n", edible, flowers)
				fmt.Printf("  BestFood: %s\n", foodInfo)
				fmt.Printf("  Activity: %s, Prefs: %d\n", c.CurrentActivity, len(c.Preferences))
			}
		}

		// Checkpoint status
		if checkIdx < len(checkpoints) && tick == checkpoints[checkIdx] {
			alive := countAliveChars(world)
			edible, flowers := countItems(world)
			fmt.Printf("  [Checkpoint tick %d: %d/%d alive, %d edible, %d flowers]\n",
				tick, alive, initialChars, edible, flowers)
			checkIdx++
		}

		// Stop if everyone is dead
		if countAliveChars(world) == 0 {
			fmt.Printf("\nAll characters dead at tick %d (%.0fs)\n", tick, float64(tick)*delta)
			break
		}
	}

	fmt.Println("\n=== FINAL SUMMARY ===")
	alive := countAliveChars(world)
	fmt.Printf("Survivors: %d/%d\n", alive, initialChars)
	fmt.Printf("Total deaths: %d\n", len(deaths))

	if len(deaths) > 0 {
		causeCounts := make(map[string]int)
		for _, d := range deaths {
			causeCounts[d.cause]++
		}
		fmt.Println("Death causes:")
		for cause, count := range causeCounts {
			fmt.Printf("  %s: %d\n", cause, count)
		}
	}
}
