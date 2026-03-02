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
	insertFn     func(ctx context.Context, p repository.InsertProblemParams) (models.Problem, error)
	listByUserFn func(ctx context.Context, userID int) ([]models.Problem, error)
}

func (m *mockProblemRepo) Insert(ctx context.Context, p repository.InsertProblemParams) (models.Problem, error) {
	return m.insertFn(ctx, p)
}

func (m *mockProblemRepo) ListByUser(ctx context.Context, userID int) ([]models.Problem, error) {
	return m.listByUserFn(ctx, userID)
}

func TestLog_EmptyName(t *testing.T) {
	svc := NewProblemService(&mockProblemRepo{})
	_, err := svc.Log(context.Background(), 1, LogProblemInput{
		Name: "", Category: "array", Difficulty: "easy",
	})
	var ve ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("expected ValidationError, got %v", err)
	}
}

func TestLog_InvalidCategory(t *testing.T) {
	svc := NewProblemService(&mockProblemRepo{})
	_, err := svc.Log(context.Background(), 1, LogProblemInput{
		Name: "Two Sum", Category: "invalid", Difficulty: "easy",
	})
	var ve ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("expected ValidationError, got %v", err)
	}
}

func TestLog_InvalidDifficulty(t *testing.T) {
	svc := NewProblemService(&mockProblemRepo{})
	_, err := svc.Log(context.Background(), 1, LogProblemInput{
		Name: "Two Sum", Category: "array", Difficulty: "extreme",
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
		Name: "Two Sum", Category: "array", Difficulty: "easy", Attempts: 0,
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
		Name: "Two Sum", Category: "array", Difficulty: "easy",
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

func TestLog_DecayedScoreSet(t *testing.T) {
	solvedAt := time.Now().Add(-10 * 24 * time.Hour)
	svc := NewProblemService(&mockProblemRepo{
		insertFn: func(_ context.Context, p repository.InsertProblemParams) (models.Problem, error) {
			return models.Problem{Score: 100, SolvedAt: solvedAt}, nil
		},
	})
	p, err := svc.Log(context.Background(), 1, LogProblemInput{
		Name: "Two Sum", Category: "array", Difficulty: "easy", Attempts: 1,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// After 10 days: 100 - (7 * 2) = 86; not full score
	if p.DecayedScore >= 100 {
		t.Errorf("expected decayed score < 100, got %f", p.DecayedScore)
	}
}

func TestList_AppliesDecay(t *testing.T) {
	solvedAt := time.Now().Add(-10 * 24 * time.Hour)
	svc := NewProblemService(&mockProblemRepo{
		listByUserFn: func(_ context.Context, _ int) ([]models.Problem, error) {
			return []models.Problem{
				{ID: 1, Score: 100, SolvedAt: solvedAt},
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
	if problems[0].DecayedScore >= 100 {
		t.Errorf("expected decayed score < 100, got %f", problems[0].DecayedScore)
	}
}
