package service

import (
	"context"
	"strings"
	"time"

	"github.com/apgupta3091/interview-iq/internal/models"
	"github.com/apgupta3091/interview-iq/internal/repository"
)

var validCategories = map[string]bool{
	"array": true, "string": true, "hash-map": true, "two-pointers": true,
	"sliding-window": true, "binary-search": true, "stack": true, "queue": true,
	"linked-list": true, "tree": true, "trie": true, "graph": true,
	"advanced-graphs": true, "heap": true, "dp": true, "dp-2d": true,
	"backtracking": true, "greedy": true, "intervals": true,
	"math": true, "bit-manipulation": true, "other": true,
}

var validDifficulties = map[string]bool{
	"easy": true, "medium": true, "hard": true,
}

type LogProblemInput struct {
	Name, Category, Difficulty string
	Attempts, TimeTakenMins    int
	LookedAtSolution           bool
}

type ProblemService interface {
	Log(ctx context.Context, userID int, req LogProblemInput) (models.Problem, error)
	List(ctx context.Context, userID int) ([]models.Problem, error)
}

type problemService struct {
	problems repository.ProblemRepository
}

func NewProblemService(problems repository.ProblemRepository) ProblemService {
	return &problemService{problems: problems}
}

func (s *problemService) Log(ctx context.Context, userID int, req LogProblemInput) (models.Problem, error) {
	req.Name = strings.TrimSpace(req.Name)
	req.Category = strings.ToLower(strings.TrimSpace(req.Category))
	req.Difficulty = strings.ToLower(strings.TrimSpace(req.Difficulty))

	if req.Name == "" {
		return models.Problem{}, ValidationError{Message: "name is required"}
	}
	if !validCategories[req.Category] {
		return models.Problem{}, ValidationError{Message: "invalid category"}
	}
	if !validDifficulties[req.Difficulty] {
		return models.Problem{}, ValidationError{Message: "difficulty must be easy, medium, or hard"}
	}
	if req.Attempts < 1 {
		req.Attempts = 1
	}

	score := models.ComputeScore(req.Attempts, req.LookedAtSolution)
	solvedAt := time.Now()

	p, err := s.problems.Insert(ctx, repository.InsertProblemParams{
		UserID:           userID,
		Name:             req.Name,
		Category:         req.Category,
		Difficulty:       req.Difficulty,
		Attempts:         req.Attempts,
		TimeTakenMins:    req.TimeTakenMins,
		LookedAtSolution: req.LookedAtSolution,
		Score:            score,
		SolvedAt:         solvedAt,
	})
	if err != nil {
		return models.Problem{}, err
	}

	p.DecayedScore = models.ApplyDecay(p.Score, p.SolvedAt)
	return p, nil
}

func (s *problemService) List(ctx context.Context, userID int) ([]models.Problem, error) {
	problems, err := s.problems.ListByUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	for i, p := range problems {
		problems[i].DecayedScore = models.ApplyDecay(p.Score, p.SolvedAt)
	}
	return problems, nil
}
