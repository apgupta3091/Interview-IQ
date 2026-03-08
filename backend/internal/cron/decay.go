package cron

import (
	"context"
	"log/slog"
	"time"

	"github.com/apgupta3091/interview-iq/internal/repository"
)

// RunDecayCron starts a goroutine that runs DecayAllProblems once per day at 10pm EST.
// The goroutine exits when ctx is cancelled (e.g. on server shutdown).
func RunDecayCron(ctx context.Context, repo repository.ProblemRepository) {
	go func() {
		for {
			next := nextRunTime()
			slog.Info("decay cron scheduled", "next_run", next)
			select {
			case <-ctx.Done():
				return
			case <-time.After(time.Until(next)):
			}
			n, err := repo.DecayAllProblems(ctx, time.Now().UTC())
			if err != nil {
				slog.ErrorContext(ctx, "decay cron failed", "error", err)
			} else {
				slog.InfoContext(ctx, "decay cron completed", "rows_updated", n)
			}
		}
	}()
}

// nextRunTime returns the next 10pm EST wall-clock time after now.
func nextRunTime() time.Time {
	loc, err := time.LoadLocation("America/New_York")
	if err != nil {
		// Fallback if tzdata is unavailable (shouldn't happen with _ "time/tzdata" in main).
		loc = time.FixedZone("EST", -5*60*60)
	}
	now := time.Now().In(loc)
	t := time.Date(now.Year(), now.Month(), now.Day(), 22, 0, 0, 0, loc)
	if !t.After(now) {
		t = t.Add(24 * time.Hour)
	}
	return t
}
