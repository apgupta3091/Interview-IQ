package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/apgupta3091/interview-iq/internal/models"
	"github.com/apgupta3091/interview-iq/internal/repository"
)

type mockProblemRepo struct {
	insertFn             func(ctx context.Context, p repository.InsertProblemParams) (models.Problem, error)
	listByUserFn         func(ctx context.Context, userID int) ([]models.Problem, error)
	listByUserFilteredFn func(ctx context.Context, userID int, f repository.ListProblemsFilter) (repository.ListProblemsResult, error)
}

func (m *mockProblemRepo) Insert(ctx context.Context, p repository.InsertProblemParams) (models.Problem, error) {
	return m.insertFn(ctx, p)
}

func (m *mockProblemRepo) ListByUser(ctx context.Context, userID int) ([]models.Problem, error) {
	return m.listByUserFn(ctx, userID)
}

func (m *mockProblemRepo) ListByUserFiltered(ctx context.Context, userID int, f repository.ListProblemsFilter) (repository.ListProblemsResult, error) {
	if m.listByUserFilteredFn != nil {
		return m.listByUserFilteredFn(ctx, userID, f)
	}
	return repository.ListProblemsResult{}, nil
}

func (m *mockProblemRepo) GetByID(_ context.Context, _, _ int) (models.Problem, error) {
	return models.Problem{}, nil
}

func (m *mockProblemRepo) DecayAllProblems(_ context.Context, _ time.Time) (int64, error) {
	return 0, nil
}

func TestLog_EmptyName(t *testing.T) {
	svc := NewProblemService(&mockProblemRepo{})
	_, err := svc.Log(context.Background(), 1, LogProblemInput{
		Name: "", Categories: []string{"array"}, Difficulty: "easy",
	})
	var ve ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("expected ValidationError, got %v", err)
	}
}

func TestLog_InvalidCategory(t *testing.T) {
	svc := NewProblemService(&mockProblemRepo{})
	_, err := svc.Log(context.Background(), 1, LogProblemInput{
		Name: "Two Sum", Categories: []string{"invalid"}, Difficulty: "easy",
	})
	var ve ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("expected ValidationError, got %v", err)
	}
}

func TestLog_InvalidDifficulty(t *testing.T) {
	svc := NewProblemService(&mockProblemRepo{})
	_, err := svc.Log(context.Background(), 1, LogProblemInput{
		Name: "Two Sum", Categories: []string{"array"}, Difficulty: "extreme",
	})
	var ve ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("expected ValidationError, got %v", err)
	}
}

func TestLog_DefaultAttempts(t *testing.T) {
	var gotParams repository.InsertProblemParams
	svc := NewProblemService(&mockProblemRepo{
		insertFn: func(_ context.Context, p repository.InsertProblemParams) (models.Problem, error) {
			gotParams = p
			return models.Problem{Score: p.Score, SolvedAt: time.Now()}, nil
		},
	})
	_, err := svc.Log(context.Background(), 1, LogProblemInput{
		Name: "Two Sum", Categories: []string{"array"}, Difficulty: "easy", Attempts: 0,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotParams.Attempts != 1 {
		t.Errorf("expected attempts=1 (defaulted), got %d", gotParams.Attempts)
	}
}

func TestLog_ScoreComputed(t *testing.T) {
	var gotParams repository.InsertProblemParams
	svc := NewProblemService(&mockProblemRepo{
		insertFn: func(_ context.Context, p repository.InsertProblemParams) (models.Problem, error) {
			gotParams = p
			return models.Problem{Score: p.Score, SolvedAt: time.Now()}, nil
		},
	})
	_, err := svc.Log(context.Background(), 1, LogProblemInput{
		Name: "Two Sum", Categories: []string{"array"}, Difficulty: "easy",
		Attempts: 1, LookedAtSolution: false,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// 1 attempt, no solution → score = 100
	if gotParams.Score != 100 {
		t.Errorf("expected score=100, got %d", gotParams.Score)
	}
}

func TestLog_OriginalScoreMatchesScore(t *testing.T) {
	// When a new problem is logged, Score and OriginalScore must be equal
	// (no decay has occurred yet; the cron will update Score later).
	var gotParams repository.InsertProblemParams
	svc := NewProblemService(&mockProblemRepo{
		insertFn: func(_ context.Context, p repository.InsertProblemParams) (models.Problem, error) {
			gotParams = p
			return models.Problem{Score: p.Score, OriginalScore: p.OriginalScore}, nil
		},
	})
	_, err := svc.Log(context.Background(), 1, LogProblemInput{
		Name: "Two Sum", Categories: []string{"array"}, Difficulty: "easy", Attempts: 1,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotParams.Score != gotParams.OriginalScore {
		t.Errorf("Score (%d) != OriginalScore (%d) on insert", gotParams.Score, gotParams.OriginalScore)
	}
}

func TestList_ReturnsProblemsDirect(t *testing.T) {
	// List now returns problems as-is from the repo; score is already decayed in the DB.
	svc := NewProblemService(&mockProblemRepo{
		listByUserFn: func(_ context.Context, _ int) ([]models.Problem, error) {
			return []models.Problem{
				{ID: 1, Score: 93, OriginalScore: 100},
			}, nil
		},
	})
	problems, err := svc.List(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(problems) != 1 {
		t.Fatalf("expected 1 problem, got %d", len(problems))
	}
	if problems[0].Score != 93 {
		t.Errorf("expected Score=93 (already decayed from DB), got %d", problems[0].Score)
	}
}
