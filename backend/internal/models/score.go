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
	bruteForcePenalty = 15 // deducted when only a brute-force (non-optimal) solution was reached;
	// less severe than peeking but still acknowledges the gap to optimal
	minScore = 5 // floor — logging any problem is worth something

	decayGraceDays  = 7    // no decay within the first 7 days
	decayPerDay     = 1.0  // points lost per day after grace period
	decayFloorRatio = 0.40 // score never drops below 40% of original
)

// ComputeScore calculates the raw score for a problem attempt.
// It does not factor in time decay — that is applied separately at read time.
//
// solutionType must be one of "none", "brute_force", or "optimal".
// "brute_force" applies a penalty because reaching an optimal solution is the goal;
// "optimal" and "none" carry no additional penalty from this dimension.
func ComputeScore(attempts int, lookedAtSolution bool, solutionType string) int {
	score := baseScore

	// penalise extra attempts beyond the first, capped
	extraAttempts := max(attempts-1, 0)
	penalty := min(extraAttempts*attemptPenalty, maxAttemptPenalty)
	score -= penalty

	if lookedAtSolution {
		score -= solutionPenalty
	}

	// Brute-force solution reached but not optimal — partial credit reduction.
	// Skip the penalty if the user also peeked at the solution; the solution
	// penalty is already the more significant one.
	if solutionType == "brute_force" && !lookedAtSolution {
		score -= bruteForcePenalty
	}

	score = max(score, minScore)
	return score
}

// ApplyDecay reduces a score based on how many days have passed since it was solved.
//
// Model:
//   - Days 0–7:  no decay (grace period, pattern is still fresh)
//   - Days 8+:   linear decay of 1 point per day
//   - Floor:     score never drops below 40% of the original (you don't fully forget patterns)
//
// Example for a score of 100:
//
//	Day 7 → 100 | Day 14 → 93 | Day 30 → 77 | Day 60 → 47 (floors at 40)
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
