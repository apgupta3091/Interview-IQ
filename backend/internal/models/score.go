package models

import (
	"math"
	"time"
)

const (
	baseScore         = 100
	attemptPenalty    = 10 // deducted per extra attempt beyond the first
	maxAttemptPenalty = 40 // cap so score can't go below 60 from attempts alone
	solutionPenalty   = 25 // deducted if the user looked at the solution
	minScore          = 5  // floor — logging any problem is worth something

	decayGraceDays  = 3    // no decay within the first 3 days
	decayPerDay     = 2.0  // points lost per day after grace period
	decayFloorRatio = 0.30 // score never drops below 30% of original
)

// ComputeScore calculates the raw score for a problem attempt.
// It does not factor in time decay — that is applied separately at read time.
func ComputeScore(attempts int, lookedAtSolution bool) int {
	score := baseScore

	// penalise extra attempts beyond the first, capped
	extraAttempts := max(attempts-1, 0)
	penalty := min(extraAttempts*attemptPenalty, maxAttemptPenalty)
	score -= penalty

	if lookedAtSolution {
		score -= solutionPenalty
	}

	score = max(score, minScore)
	return score
}

// ApplyDecay reduces a score based on how many days have passed since it was solved.
//
// Model:
//   - Days 0–3:  no decay (grace period, pattern is still fresh)
//   - Days 4+:   linear decay of 2 points per day
//   - Floor:     score never drops below 30% of the original (you don't fully forget patterns)
//
// Example for a score of 100:
//
//	Day 3 → 100 | Day 10 → 86 | Day 21 → 64 | Day 30 → 46 (floors at 30)
func ApplyDecay(score int, solvedAt time.Time) float64 {
	daysSince := time.Since(solvedAt).Hours() / 24
	if daysSince <= decayGraceDays {
		return float64(score)
	}

	floor := math.Ceil(float64(score) * decayFloorRatio)
	decayDays := daysSince - decayGraceDays
	decayed := float64(score) - (decayDays * decayPerDay)

	decayed = max(decayed, floor)
	return math.Round(decayed*10) / 10
}
