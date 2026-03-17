package models

import "time"

type Problem struct {
	ID, UserID, Score, OriginalScore, Attempts, TimeTakenMins int
	Name, Difficulty, SolutionType                            string
	Notes                                                     string
	Categories                                                []string
	LookedAtSolution                                          bool
	SolvedAt, CreatedAt                                       time.Time
}

// LeetCodeProblem is a row from the leetcode_problems catalog table.
type LeetCodeProblem struct {
	ID, LcID int
	Title    string
	Slug     string
	// Difficulty is one of "easy", "medium", "hard" (lowercased at insert time).
	Difficulty string
	// Tags holds our mapped category slugs (e.g. "array", "hash-map").
	Tags     []string
	PaidOnly bool
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
	// ScoreReady is true when the category has at least 3 submitted problems,
	// which is the minimum needed for a meaningful strength score.
	ScoreReady bool
}

type WeakestResult struct {
	Category        string
	Strength        float64
	Recommendations []string
}

// Note is a standalone user note attached to a problem (keyed by problem name,
// not by a specific attempt).
type Note struct {
	ID          int
	UserID      int
	ProblemName string
	Content     string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
