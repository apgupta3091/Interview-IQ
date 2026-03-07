package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/apgupta3091/interview-iq/internal/models"
)

type mockCategoryRepo struct {
	getRawScoresFn func(ctx context.Context, userID int) ([]models.CategoryRawScore, error)
}

func (m *mockCategoryRepo) GetRawScoresByUser(ctx context.Context, userID int) ([]models.CategoryRawScore, error) {
	return m.getRawScoresFn(ctx, userID)
}

func TestGetStats_Aggregates(t *testing.T) {
	now := time.Now()
	svc := NewCategoryService(&mockCategoryRepo{
		getRawScoresFn: func(_ context.Context, _ int) ([]models.CategoryRawScore, error) {
			return []models.CategoryRawScore{
				{Category: "array", Score: 100, SolvedAt: now},
				{Category: "array", Score: 80, SolvedAt: now},
				{Category: "dp", Score: 60, SolvedAt: now},
			}, nil
		},
	})

	stats, err := svc.GetStats(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(stats) != 2 {
		t.Fatalf("expected 2 categories, got %d", len(stats))
	}

	statsMap := map[string]models.CategoryStats{}
	for _, s := range stats {
		statsMap[s.Category] = s
	}

	if statsMap["array"].ProblemCount != 2 {
		t.Errorf("expected array problem_count=2, got %d", statsMap["array"].ProblemCount)
	}
	if statsMap["array"].Strength != 90.0 {
		t.Errorf("expected array strength=90.0, got %f", statsMap["array"].Strength)
	}
	if statsMap["dp"].ProblemCount != 1 {
		t.Errorf("expected dp problem_count=1, got %d", statsMap["dp"].ProblemCount)
	}
}

// TestGetStats_LatestEntryWins verifies that when the repo already returns only
// the latest log per problem name (via DISTINCT ON), GetStats reflects that
// latest score rather than an average of all historical attempts.
// The mock simulates what the repo would return after the DISTINCT ON filter:
// one entry for "Two Sum" at score 100 (the improvement, not the earlier 65).
func TestGetStats_LatestEntryWins(t *testing.T) {
	now := time.Now()
	svc := NewCategoryService(&mockCategoryRepo{
		getRawScoresFn: func(_ context.Context, _ int) ([]models.CategoryRawScore, error) {
			// Repo returns only the latest entry per name — score 100 for "Two Sum".
			return []models.CategoryRawScore{
				{Category: "array", Score: 100, SolvedAt: now},
			}, nil
		},
	})

	stats, err := svc.GetStats(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(stats) != 1 {
		t.Fatalf("expected 1 category, got %d", len(stats))
	}
	if stats[0].Category != "array" {
		t.Errorf("expected category=array, got %s", stats[0].Category)
	}
	// Strength should reflect the latest score (100), not an average with any
	// earlier attempt. ApplyDecay on a just-solved problem leaves score at 100.
	if stats[0].Strength != 100.0 {
		t.Errorf("expected strength=100.0 (latest attempt wins), got %f", stats[0].Strength)
	}
	if stats[0].ProblemCount != 1 {
		t.Errorf("expected problem_count=1, got %d", stats[0].ProblemCount)
	}
}

func TestGetWeakest_FindsMin(t *testing.T) {
	now := time.Now()
	svc := NewCategoryService(&mockCategoryRepo{
		getRawScoresFn: func(_ context.Context, _ int) ([]models.CategoryRawScore, error) {
			return []models.CategoryRawScore{
				{Category: "array", Score: 100, SolvedAt: now},
				{Category: "dp", Score: 40, SolvedAt: now},
			}, nil
		},
	})

	result, err := svc.GetWeakest(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Category != "dp" {
		t.Errorf("expected weakest=dp, got %s", result.Category)
	}
	if len(result.Recommendations) == 0 {
		t.Error("expected non-empty recommendations for dp")
	}
}

func TestGetWeakest_ErrNoProblems(t *testing.T) {
	svc := NewCategoryService(&mockCategoryRepo{
		getRawScoresFn: func(_ context.Context, _ int) ([]models.CategoryRawScore, error) {
			return nil, nil
		},
	})
	_, err := svc.GetWeakest(context.Background(), 1)
	if !errors.Is(err, ErrNoProblems) {
		t.Fatalf("expected ErrNoProblems, got %v", err)
	}
}

