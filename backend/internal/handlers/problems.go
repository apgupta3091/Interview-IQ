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
	Name             string   `json:"name"`
	Categories       []string `json:"categories"`
	Difficulty       string   `json:"difficulty"`
	Attempts         int      `json:"attempts"`
	LookedAtSolution bool     `json:"looked_at_solution"`
	TimeTakenMins    int      `json:"time_taken_mins"`
}

type problemResponse struct {
	ID               int       `json:"id"`
	Name             string    `json:"name"`
	Categories       []string  `json:"categories"`
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
	categories := p.Categories
	if categories == nil {
		categories = []string{}
	}
	return problemResponse{
		ID:               p.ID,
		Name:             p.Name,
		Categories:       categories,
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

// ListProblems godoc
// @Summary      List all logged problems
// @Description  Returns all problems logged by the authenticated user, ordered newest first. Each problem includes its raw score and live decayed score computed at request time.
// @Tags         problems
// @Produce      json
// @Security     BearerAuth
// @Success      200  {array}   problemResponse "List of problems with decayed scores"
// @Failure      401  {object}  errorResponse   "Missing or invalid JWT token"
// @Failure      500  {object}  errorResponse   "Internal server error"
// @Router       /problems [get]
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

// LogProblem godoc
// @Summary      Log a new problem attempt
// @Description  Records a LeetCode problem attempt and computes a score. Score formula: base 100, -10 per extra attempt (capped -40), -25 if solution viewed, floor 5. Decay applies at read time: 3-day grace period, then -2pts/day, floor at 30% of raw score.
// @Tags         problems
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body      logProblemRequest  true  "Problem attempt details"
// @Success      201   {object}  problemResponse    "Created problem with computed score"
// @Failure      400   {object}  errorResponse      "Invalid input — bad category, difficulty, or missing name"
// @Failure      401   {object}  errorResponse      "Missing or invalid JWT token"
// @Failure      500   {object}  errorResponse      "Internal server error"
// @Router       /problems [post]
func (h *ProblemHandler) LogProblem(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())

	var req logProblemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	p, err := h.Service.Log(r.Context(), userID, service.LogProblemInput{
		Name:             req.Name,
		Categories:       req.Categories,
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
