package system

import (
	"fmt"

	"petri/internal/config"
	"petri/internal/entity"
)

// UpdateSurvival updates hunger, thirst, energy, poison, and health for a character
func UpdateSurvival(char *entity.Character, deltaTime float64, log *ActionLog) {
	if char.IsDead {
		return
	}

	// Handle frustration timer
	if char.IsFrustrated {
		char.FrustrationTimer -= deltaTime
		if char.FrustrationTimer <= 0 {
			char.IsFrustrated = false
			char.FrustrationTimer = 0
			char.CurrentActivity = "Idle"
			if log != nil {
				log.Add(char.ID, char.Name, "activity", "Calmed down")
			}
		}
	}

	// Decrement satisfaction cooldowns
	if char.HungerCooldown > 0 {
		char.HungerCooldown -= deltaTime
		if char.HungerCooldown < 0 {
			char.HungerCooldown = 0
		}
	}
	if char.ThirstCooldown > 0 {
		char.ThirstCooldown -= deltaTime
		if char.ThirstCooldown < 0 {
			char.ThirstCooldown = 0
		}
	}
	if char.EnergyCooldown > 0 {
		char.EnergyCooldown -= deltaTime
		if char.EnergyCooldown < 0 {
			char.EnergyCooldown = 0
		}
	}

	prevHunger := char.Hunger
	prevThirst := char.Thirst
	prevEnergy := char.Energy

	// Increase hunger over time (always, even while sleeping) - unless on cooldown
	if char.HungerCooldown <= 0 {
		char.Hunger += config.HungerIncreaseRate * deltaTime
		if char.Hunger > 100 {
			char.Hunger = 100
		}
	}

	// Increase thirst over time (always, even while sleeping) - unless on cooldown
	if char.ThirstCooldown <= 0 {
		char.Thirst += config.ThirstIncreaseRate * deltaTime
		if char.Thirst > 100 {
			char.Thirst = 100
		}
	}

	// Handle energy: restore while sleeping, decrease while awake
	if char.IsSleeping {
		restoreRate := config.GroundEnergyRestoreRate
		if char.AtBed {
			restoreRate = config.BedEnergyRestoreRate
		}
		char.Energy += restoreRate * deltaTime
		if char.Energy > 100 {
			char.Energy = 100
		}

		// Check if another stat is urgent enough to wake (must be Moderate+ AND worse raw urgency)
		hungerTier := char.HungerTier()
		thirstTier := char.ThirstTier()
		energyUrgency := char.EnergyUrgency()

		hungerWakes := hungerTier >= entity.TierModerate && char.HungerUrgency() > energyUrgency
		thirstWakes := thirstTier >= entity.TierModerate && char.ThirstUrgency() > energyUrgency

		if hungerWakes || thirstWakes {
			char.IsSleeping = false
			char.AtBed = false
			char.CurrentActivity = "Waking up"
			if log != nil {
				if hungerWakes && (!thirstWakes || char.HungerUrgency() > char.ThirstUrgency()) {
					log.Add(char.ID, char.Name, "sleep", "Woke up due to hunger")
				} else {
					log.Add(char.ID, char.Name, "sleep", "Woke up due to thirst")
				}
			}
		}

		// Wake thresholds: 100 in bed, 75 on ground (VI-B-8)
		wakeThreshold := 75.0
		if char.AtBed {
			wakeThreshold = 100.0
		}
		if char.IsSleeping && char.Energy >= wakeThreshold {
			char.IsSleeping = false
			char.AtBed = false
			char.CurrentActivity = "Waking up"
			// Set energy cooldown and boost mood when waking fully rested
			if wakeThreshold >= 100 {
				char.EnergyCooldown = config.SatisfactionCooldown
				prevMoodTier := char.MoodTier()
				char.Mood += config.MoodBoostOnConsumption
				if char.Mood > 100 {
					char.Mood = 100
				}
				// Log mood tier transition from boost
				if log != nil && char.MoodTier() != prevMoodTier {
					log.Add(char.ID, char.Name, "mood", fmt.Sprintf("Feeling %s", char.MoodLevel()))
				}
			}
			if log != nil {
				if wakeThreshold >= 100 {
					log.Add(char.ID, char.Name, "sleep", "Woke up fully rested")
				} else {
					log.Add(char.ID, char.Name, "sleep", "Woke up partially rested")
				}
			}
		}
	} else {
		// Decrease energy over time (base rate) when awake - unless on cooldown
		if char.EnergyCooldown <= 0 {
			char.Energy -= config.EnergyDecreaseRate * deltaTime
			if char.Energy < 0 {
				char.Energy = 0
			}
		}
	}

	// Log hunger milestones
	if log != nil {
		if prevHunger < 50 && char.Hunger >= 50 {
			log.Add(char.ID, char.Name, "hunger", "Getting hungry")
		} else if prevHunger < 75 && char.Hunger >= 75 {
			log.Add(char.ID, char.Name, "hunger", "Very hungry!")
		} else if prevHunger < 90 && char.Hunger >= 90 {
			log.Add(char.ID, char.Name, "hunger", "Ravenous!")
		} else if prevHunger < 100 && char.Hunger >= 100 {
			log.Add(char.ID, char.Name, "hunger", "Starving!")
		}
	}

	// Log thirst milestones
	if log != nil {
		if prevThirst < 50 && char.Thirst >= 50 {
			log.Add(char.ID, char.Name, "thirst", "Getting thirsty")
		} else if prevThirst < 75 && char.Thirst >= 75 {
			log.Add(char.ID, char.Name, "thirst", "Very thirsty!")
		} else if prevThirst < 90 && char.Thirst >= 90 {
			log.Add(char.ID, char.Name, "thirst", "Parched!")
		} else if prevThirst < 100 && char.Thirst >= 100 {
			log.Add(char.ID, char.Name, "thirst", "Dehydrated!")
		}
	}

	// Log energy milestones (only when awake)
	// Use independent if statements so multiple threshold crossings in one tick all get logged
	if !char.IsSleeping && log != nil {
		if prevEnergy > 50 && char.Energy <= 50 {
			log.Add(char.ID, char.Name, "energy", "Getting tired")
		}
		if prevEnergy > 25 && char.Energy <= 25 {
			log.Add(char.ID, char.Name, "energy", "Very tired!")
		}
		if prevEnergy > 10 && char.Energy <= 10 {
			log.Add(char.ID, char.Name, "energy", "Exhausted!")
		}
		if prevEnergy > 0 && char.Energy <= 0 {
			log.Add(char.ID, char.Name, "energy", "Collapsed from exhaustion!")
		}
	}

	// Handle poison (always, even while sleeping)
	if char.Poisoned {
		char.PoisonTimer -= deltaTime

		// Apply poison damage
		char.Health -= config.PoisonDamageRate * deltaTime

		// Check if poison has worn off
		if char.PoisonTimer <= 0 {
			char.Poisoned = false
			char.PoisonTimer = 0

			if log != nil {
				log.Add(char.ID, char.Name, "poison", "Poison wore off")
			}
		}
	}

	// Handle starvation damage (always, even while sleeping)
	if char.Hunger >= 100 {
		oldHealth := char.Health
		char.Health -= config.StarvationDamageRate * deltaTime

		// Log starvation damage periodically
		if log != nil && int(oldHealth/5) > int(char.Health/5) {
			log.Add(char.ID, char.Name, "health",
				fmt.Sprintf("Starving! Health: %d/100", int(char.Health)))
		}
	}

	// Handle dehydration damage (always, even while sleeping)
	if char.Thirst >= 100 {
		oldHealth := char.Health
		char.Health -= config.DehydrationDamageRate * deltaTime

		// Log dehydration damage periodically
		if log != nil && int(oldHealth/5) > int(char.Health/5) {
			log.Add(char.ID, char.Name, "health",
				fmt.Sprintf("Dehydrated! Health: %d/100", int(char.Health)))
		}
	}

	// Check for death
	if char.Health <= 0 {
		char.Health = 0
		char.IsDead = true
		char.IsSleeping = false // Can't be sleeping if dead
		char.CurrentActivity = "Dead"

		if log != nil {
			log.Add(char.ID, char.Name, "death", "Died")
		}
	}

	// Update mood based on need states
	UpdateMood(char, deltaTime, log)
}

// UpdateMood adjusts mood based on the highest need tier
func UpdateMood(char *entity.Character, deltaTime float64, log *ActionLog) {
	if char.IsDead {
		return
	}

	prevTier := char.MoodTier()

	// Find highest need tier (hunger, thirst, energy, health - excluding mood itself)
	highestTier := char.HungerTier()
	if char.ThirstTier() > highestTier {
		highestTier = char.ThirstTier()
	}
	if char.EnergyTier() > highestTier {
		highestTier = char.EnergyTier()
	}
	if char.HealthTier() > highestTier {
		highestTier = char.HealthTier()
	}

	// Adjust mood based on highest need tier
	switch highestTier {
	case entity.TierNone:
		// All needs met - mood increases slowly
		char.Mood += config.MoodIncreaseRate * deltaTime
	case entity.TierMild:
		// Mild need - no mood change (neutral)
	case entity.TierModerate:
		// Moderate need - mood decreases slowly
		char.Mood -= config.MoodDecreaseRateSlow * deltaTime
	case entity.TierSevere:
		// Severe need - mood decreases at medium rate
		char.Mood -= config.MoodDecreaseRateMedium * deltaTime
	case entity.TierCrisis:
		// Crisis need - mood decreases quickly
		char.Mood -= config.MoodDecreaseRateFast * deltaTime
	}

	// Additional penalties for status effects (additive with need-based decay)
	if char.Poisoned {
		char.Mood -= config.MoodPenaltyPoisoned * deltaTime
	}
	if char.IsFrustrated {
		char.Mood -= config.MoodPenaltyFrustrated * deltaTime
	}

	// Clamp mood to 0-100
	if char.Mood > 100 {
		char.Mood = 100
	}
	if char.Mood < 0 {
		char.Mood = 0
	}

	// Log tier transitions
	newTier := char.MoodTier()
	if newTier != prevTier && log != nil {
		log.Add(char.ID, char.Name, "mood", fmt.Sprintf("Feeling %s", char.MoodLevel()))
	}
}
