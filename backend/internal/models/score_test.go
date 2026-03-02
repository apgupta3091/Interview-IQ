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
		want             int
	}{
		{"perfect: 1 attempt, no solution", 1, false, 100},
		{"2 attempts, no solution", 2, false, 90},
		{"3 attempts, no solution", 3, false, 80},
		{"1 attempt, looked at solution", 1, true, 75},
		{"3 attempts, looked at solution", 3, true, 55},
		{"5 attempts, no solution (cap at -40)", 5, false, 60},
		{"6 attempts, no solution (still capped)", 6, false, 60},
		{"many attempts + solution (floor at 5)", 10, true, 35},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := ComputeScore(tc.attempts, tc.lookedAtSolution)
			if got != tc.want {
				t.Errorf("ComputeScore(%d, %v) = %d, want %d", tc.attempts, tc.lookedAtSolution, got, tc.want)
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

	// day 3 — still in grace period
	got = ApplyDecay(score, time.Now().Add(-3*24*time.Hour))
	if got != 100.0 {
		t.Errorf("day 3: expected 100.0 (grace period), got %f", got)
	}

	// day 10 — 7 days of decay after grace: 100 - (7 * 2) = 86
	got = ApplyDecay(score, time.Now().Add(-10*24*time.Hour))
	if got < 85.0 || got > 87.0 {
		t.Errorf("day 10: expected ~86, got %f", got)
	}

	// day 21 — 18 days of decay: 100 - (18 * 2) = 64
	got = ApplyDecay(score, time.Now().Add(-21*24*time.Hour))
	if got < 63.0 || got > 65.0 {
		t.Errorf("day 21: expected ~64, got %f", got)
	}

	// day 60 — would be 100 - (57*2) = -14, but floors at 30
	got = ApplyDecay(score, time.Now().Add(-60*24*time.Hour))
	if got != 30.0 {
		t.Errorf("day 60: expected 30.0 (floor), got %f", got)
	}
}
