package models

import (
	"testing"
	"time"
)

func TestComputeScore(t *testing.T) {
	cases := []struct {
		name             string
		attempts         int
		lookedAtSolution bool
		solutionType     string
		want             int
	}{
		// optimal / none — no extra penalty from solution type
		{"perfect: 1 attempt, optimal", 1, false, "optimal", 100},
		{"perfect: 1 attempt, none", 1, false, "none", 100},
		{"2 attempts, optimal", 2, false, "optimal", 90},
		{"3 attempts, optimal", 3, false, "optimal", 80},
		{"1 attempt, looked at solution", 1, true, "none", 75},
		{"3 attempts, looked at solution", 3, true, "none", 55},
		{"5 attempts, optimal (cap at -40)", 5, false, "optimal", 60},
		{"6 attempts, optimal (still capped)", 6, false, "optimal", 60},
		{"many attempts + solution (floor at 5)", 10, true, "none", 35},

		// brute_force — -15 penalty when not peeking
		{"1 attempt, brute force", 1, false, "brute_force", 85},
		{"3 attempts, brute force", 3, false, "brute_force", 65},
		{"5 attempts, brute force (attempt cap already hit)", 5, false, "brute_force", 45},

		// brute_force + peeked — solution penalty dominates; brute_force penalty skipped
		{"1 attempt, brute force + peeked", 1, true, "brute_force", 75},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := ComputeScore(tc.attempts, tc.lookedAtSolution, tc.solutionType)
			if got != tc.want {
				t.Errorf("ComputeScore(%d, %v, %q) = %d, want %d",
					tc.attempts, tc.lookedAtSolution, tc.solutionType, got, tc.want)
			}
		})
	}
}

func TestApplyDecay(t *testing.T) {
	score := 100

	// day 0 — grace period, no decay
	got := ApplyDecay(score, time.Now())
	if got != 100.0 {
		t.Errorf("day 0: expected 100.0, got %f", got)
	}

	// day 7 — still in grace period
	got = ApplyDecay(score, time.Now().Add(-7*24*time.Hour))
	if got != 100.0 {
		t.Errorf("day 7: expected 100.0 (grace period), got %f", got)
	}

	// day 14 — 7 days of decay after grace: 100 - (7 * 1) = 93
	got = ApplyDecay(score, time.Now().Add(-14*24*time.Hour))
	if got < 92.0 || got > 94.0 {
		t.Errorf("day 14: expected ~93, got %f", got)
	}

	// day 30 — 23 days of decay: 100 - (23 * 1) = 77
	got = ApplyDecay(score, time.Now().Add(-30*24*time.Hour))
	if got < 76.0 || got > 78.0 {
		t.Errorf("day 30: expected ~77, got %f", got)
	}

	// day 80 — would be 100 - (73*1) = 27, but floors at 40
	got = ApplyDecay(score, time.Now().Add(-80*24*time.Hour))
	if got != 40.0 {
		t.Errorf("day 80: expected 40.0 (floor), got %f", got)
	}
}
