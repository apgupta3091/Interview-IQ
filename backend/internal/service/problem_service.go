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
	Name, Difficulty, SolutionType string
	Categories                     []string
	Attempts, TimeTakenMins        int
	LookedAtSolution               bool
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
	req.Difficulty = strings.ToLower(strings.TrimSpace(req.Difficulty))

	if req.Name == "" {
		return models.Problem{}, ValidationError{Message: "name is required"}
	}
	if len(req.Categories) == 0 {
		return models.Problem{}, ValidationError{Message: "at least one category is required"}
	}
	// Normalise and validate every category in the list.
	for i, c := range req.Categories {
		req.Categories[i] = strings.ToLower(strings.TrimSpace(c))
		if !validCategories[req.Categories[i]] {
			return models.Problem{}, ValidationError{Message: "invalid category: " + req.Categories[i]}
		}
	}
	if !validDifficulties[req.Difficulty] {
		return models.Problem{}, ValidationError{Message: "difficulty must be easy, medium, or hard"}
	}
	if req.Attempts < 1 {
		req.Attempts = 1
	}

	// Normalise solution_type; default to "none" if unrecognised.
	validSolutionTypes := map[string]bool{"none": true, "brute_force": true, "optimal": true}
	if !validSolutionTypes[req.SolutionType] {
		req.SolutionType = "none"
	}

	score := models.ComputeScore(req.Attempts, req.LookedAtSolution, req.SolutionType)
	solvedAt := time.Now()

	p, err := s.problems.Insert(ctx, repository.InsertProblemParams{
		UserID:           userID,
		Name:             req.Name,
		Categories:       req.Categories,
		Difficulty:       req.Difficulty,
		Attempts:         req.Attempts,
		TimeTakenMins:    req.TimeTakenMins,
		LookedAtSolution: req.LookedAtSolution,
		SolutionType:     req.SolutionType,
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
