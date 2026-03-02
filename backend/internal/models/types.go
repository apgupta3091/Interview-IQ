package models

import "time"

type Problem struct {
	ID, UserID, Score, Attempts, TimeTakenMins int
	Name, Category, Difficulty                 string
	LookedAtSolution                           bool
	DecayedScore                               float64
	SolvedAt, CreatedAt                        time.Time
}

type CategoryRawScore struct {
	Category string
	Score    int
	SolvedAt time.Time
}

type CategoryStats struct {
	Category     string
	Strength     float64
	ProblemCount int
}

type WeakestResult struct {
	Category        string
	Strength        float64
	Recommendations []string
}
