package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/apgupta3091/interview-iq/internal/models"
	"github.com/apgupta3091/interview-iq/internal/repository"
)

// maxNameLen is the maximum allowed length for a problem name in characters.
const maxNameLen = 200

// freeTierProblemLimit is the maximum number of problems a free-tier user can log.
const freeTierProblemLimit = 20

// escapeLikePattern escapes LIKE/ILIKE metacharacters so a raw user string
// can be safely used as a substring pattern (repo wraps it in % wildcards).
// The repository must pair this with ESCAPE '\' in the SQL clause.
func escapeLikePattern(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `%`, `\%`)
	s = strings.ReplaceAll(s, `_`, `\_`)
	return s
}

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
	Notes                          string
	// Tier is the user's subscription tier ("free" or "pro"), set by the handler
	// from the request context. Used to enforce the free-tier problem cap.
	Tier                    string
	Categories              []string
	Attempts, TimeTakenMins int
	LookedAtSolution        bool
}

// ListProblemsParams mirrors repository.ListProblemsFilter but lives at the service layer
// so handlers depend only on the service package.
type ListProblemsParams struct {
	NameSearch   string
	Categories   []string
	Difficulties []string
	ScoreMin     *int
	ScoreMax     *int
	DateFrom     *time.Time
	DateTo       *time.Time
	Limit        int
	Offset       int
}

// ListResult holds a page of problems plus the total matching count.
type ListResult struct {
	Problems []models.Problem
	Total    int
}

type ProblemService interface {
	Log(ctx context.Context, userID int, req LogProblemInput) (models.Problem, error)
	GetByID(ctx context.Context, userID, id int) (models.Problem, error)
	List(ctx context.Context, userID int) ([]models.Problem, error)
	ListFiltered(ctx context.Context, userID int, params ListProblemsParams) (ListResult, error)
	// Count returns the total number of problems logged by the user.
	Count(ctx context.Context, userID int) (int, error)
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
	if len(req.Name) > maxNameLen {
		return models.Problem{}, ValidationError{Message: "name must be 200 characters or fewer"}
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
	// Clamp time_taken_mins to a valid positive value.
	if req.TimeTakenMins < 1 {
		req.TimeTakenMins = 1
	}

	// Normalise solution_type; default to "none" if unrecognised.
	validSolutionTypes := map[string]bool{"none": true, "brute_force": true, "optimal": true}
	if !validSolutionTypes[req.SolutionType] {
		req.SolutionType = "none"
	}

	// Enforce the free-tier problem cap before inserting.
	if req.Tier == "" || req.Tier == "free" {
		count, err := s.problems.CountByUser(ctx, userID)
		if err != nil {
			return models.Problem{}, fmt.Errorf("Log: check free tier cap: %w", err)
		}
		if count >= freeTierProblemLimit {
			return models.Problem{}, ErrFreeTierLimitReached
		}
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
		Notes:            req.Notes,
		Score:            score,
		OriginalScore:    score, // new problems start with score == original_score (no decay yet)
		SolvedAt:         solvedAt,
	})
	if err != nil {
		return models.Problem{}, err
	}
	return p, nil
}

// GetByID returns a single problem owned by userID. Translates repository.ErrNotFound
// to service.ErrNotFound so the handler does not need to import the repository package.
func (s *problemService) GetByID(ctx context.Context, userID, id int) (models.Problem, error) {
	p, err := s.problems.GetByID(ctx, id, userID)
	if errors.Is(err, repository.ErrNotFound) {
		return models.Problem{}, ErrNotFound
	}
	if err != nil {
		return models.Problem{}, err
	}
	return p, nil
}

func (s *problemService) List(ctx context.Context, userID int) ([]models.Problem, error) {
	return s.problems.ListByUser(ctx, userID)
}

func (s *problemService) ListFiltered(ctx context.Context, userID int, params ListProblemsParams) (ListResult, error) {
	// Sanitize the name search: trim whitespace, enforce length cap, and escape
	// LIKE metacharacters (% and _) so the user's literal text is matched, not
	// interpreted as a pattern. The repository pairs this with ESCAPE '\'.
	params.NameSearch = strings.TrimSpace(params.NameSearch)
	if len(params.NameSearch) > maxNameLen {
		params.NameSearch = params.NameSearch[:maxNameLen]
	}
	params.NameSearch = escapeLikePattern(params.NameSearch)

	result, err := s.problems.ListByUserFiltered(ctx, userID, repository.ListProblemsFilter{
		NameSearch:   params.NameSearch,
		Categories:   params.Categories,
		Difficulties: params.Difficulties,
		ScoreMin:     params.ScoreMin,
		ScoreMax:     params.ScoreMax,
		DateFrom:     params.DateFrom,
		DateTo:       params.DateTo,
		Limit:        params.Limit,
		Offset:       params.Offset,
	})
	if err != nil {
		return ListResult{}, err
	}
	return ListResult{Problems: result.Problems, Total: result.Total}, nil
}

// Count returns the total number of problems logged by the user.
func (s *problemService) Count(ctx context.Context, userID int) (int, error) {
	return s.problems.CountByUser(ctx, userID)
}
