package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/apgupta3091/interview-iq/internal/middleware"
	"github.com/apgupta3091/interview-iq/internal/models"
	"github.com/apgupta3091/interview-iq/internal/service"
)

type ProblemHandler struct {
	Service service.ProblemService
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

func toProblemResponse(p models.Problem) problemResponse {
	return problemResponse{
		ID:               p.ID,
		Name:             p.Name,
		Category:         p.Category,
		Difficulty:       p.Difficulty,
		Attempts:         p.Attempts,
		LookedAtSolution: p.LookedAtSolution,
		TimeTakenMins:    p.TimeTakenMins,
		Score:            p.Score,
		DecayedScore:     p.DecayedScore,
		SolvedAt:         p.SolvedAt,
		CreatedAt:        p.CreatedAt,
	}
}

func (h *ProblemHandler) ListProblems(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())

	problems, err := h.Service.List(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to fetch problems")
		return
	}

	resp := make([]problemResponse, len(problems))
	for i, p := range problems {
		resp[i] = toProblemResponse(p)
	}
	writeJSON(w, http.StatusOK, resp)
}

func (h *ProblemHandler) LogProblem(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())

	var req logProblemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	p, err := h.Service.Log(r.Context(), userID, service.LogProblemInput{
		Name:             req.Name,
		Category:         req.Category,
		Difficulty:       req.Difficulty,
		Attempts:         req.Attempts,
		LookedAtSolution: req.LookedAtSolution,
		TimeTakenMins:    req.TimeTakenMins,
	})
	if err != nil {
		var ve service.ValidationError
		if errors.As(err, &ve) {
			writeError(w, http.StatusBadRequest, ve.Message)
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to log problem")
		return
	}

	writeJSON(w, http.StatusCreated, toProblemResponse(p))
}
