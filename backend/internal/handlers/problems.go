package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/apgupta3091/interview-iq/internal/middleware"
	"github.com/apgupta3091/interview-iq/internal/models"
)

type ProblemHandler struct {
	DB *sql.DB
}

type logProblemRequest struct {
	Name             string `json:"name"`
	Category         string `json:"category"`
	Difficulty       string `json:"difficulty"`
	Attempts         int    `json:"attempts"`
	LookedAtSolution bool   `json:"looked_at_solution"`
	TimeTakenMins    int    `json:"time_taken_mins"`
}

type problemResponse struct {
	ID               int       `json:"id"`
	Name             string    `json:"name"`
	Category         string    `json:"category"`
	Difficulty       string    `json:"difficulty"`
	Attempts         int       `json:"attempts"`
	LookedAtSolution bool      `json:"looked_at_solution"`
	TimeTakenMins    int       `json:"time_taken_mins"`
	Score            int       `json:"score"`
	DecayedScore     float64   `json:"decayed_score"`
	SolvedAt         time.Time `json:"solved_at"`
	CreatedAt        time.Time `json:"created_at"`
}

var validDifficulties = map[string]bool{
	"easy": true, "medium": true, "hard": true,
}

var validCategories = map[string]bool{
	"array": true, "string": true, "hash-map": true, "two-pointers": true,
	"sliding-window": true, "binary-search": true, "stack": true, "queue": true,
	"linked-list": true, "tree": true, "graph": true, "heap": true,
	"dp": true, "backtracking": true, "greedy": true, "math": true, "other": true,
}

func (h *ProblemHandler) LogProblem(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())

	var req logProblemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	req.Name = strings.TrimSpace(req.Name)
	req.Category = strings.ToLower(strings.TrimSpace(req.Category))
	req.Difficulty = strings.ToLower(strings.TrimSpace(req.Difficulty))

	if req.Name == "" {
		writeError(w, http.StatusBadRequest, "name is required")
		return
	}
	if !validCategories[req.Category] {
		writeError(w, http.StatusBadRequest, "invalid category")
		return
	}
	if !validDifficulties[req.Difficulty] {
		writeError(w, http.StatusBadRequest, "difficulty must be easy, medium, or hard")
		return
	}
	if req.Attempts < 1 {
		req.Attempts = 1
	}

	score := models.ComputeScore(req.Attempts, req.LookedAtSolution)
	solvedAt := time.Now()

	var p problemResponse
	err := h.DB.QueryRowContext(r.Context(), `
		INSERT INTO problems
			(user_id, name, category, difficulty, attempts, looked_at_solution, time_taken_mins, score, solved_at)
		VALUES
			($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, name, category, difficulty, attempts, looked_at_solution, time_taken_mins, score, solved_at, created_at`,
		userID, req.Name, req.Category, req.Difficulty,
		req.Attempts, req.LookedAtSolution, req.TimeTakenMins,
		score, solvedAt,
	).Scan(
		&p.ID, &p.Name, &p.Category, &p.Difficulty,
		&p.Attempts, &p.LookedAtSolution, &p.TimeTakenMins,
		&p.Score, &p.SolvedAt, &p.CreatedAt,
	)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to log problem")
		return
	}

	p.DecayedScore = models.ApplyDecay(p.Score, p.SolvedAt)
	writeJSON(w, http.StatusCreated, p)
}
