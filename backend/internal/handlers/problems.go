package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
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
	// SolutionType indicates how the user solved the problem.
	// Accepted values: "none" (default), "brute_force", "optimal".
	// "brute_force" applies a -15 point penalty; the others carry no penalty.
	SolutionType string `json:"solution_type"`
}

type problemResponse struct {
	ID               int       `json:"id"`
	Name             string    `json:"name"`
	Categories       []string  `json:"categories"`
	Difficulty       string    `json:"difficulty"`
	Attempts         int       `json:"attempts"`
	LookedAtSolution bool      `json:"looked_at_solution"`
	TimeTakenMins    int       `json:"time_taken_mins"`
	SolutionType     string    `json:"solution_type"`
	Score            int       `json:"score"`
	OriginalScore    int       `json:"original_score"`
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
		SolutionType:     p.SolutionType,
		Score:            p.Score,
		OriginalScore:    p.OriginalScore,
		SolvedAt:         p.SolvedAt,
		CreatedAt:        p.CreatedAt,
	}
}

// listProblemsResponse is the paginated envelope returned by ListProblems.
type listProblemsResponse struct {
	Problems []problemResponse `json:"problems"`
	Total    int               `json:"total"`
	Limit    int               `json:"limit"`
	Offset   int               `json:"offset"`
}

// ListProblems godoc
// @Summary      List logged problems with filtering and pagination
// @Description  Returns a paginated, filtered list of problems for the authenticated user, always sorted by created_at DESC. Supports name search, date range, category/difficulty/score filters, and offset-based pagination.
// @Tags         problems
// @Produce      json
// @Security     BearerAuth
// @Param        q           query  string  false  "Name search (case-insensitive partial match)"
// @Param        category    query  []string false  "Category filter (repeatable; matches any)"
// @Param        difficulty  query  []string false  "Difficulty filter (repeatable: easy|medium|hard)"
// @Param        score_min   query  int     false  "Minimum raw score (inclusive)"
// @Param        score_max   query  int     false  "Maximum raw score (inclusive)"
// @Param        from        query  string  false  "Solved-on start date YYYY-MM-DD (inclusive)"
// @Param        to          query  string  false  "Solved-on end date YYYY-MM-DD (inclusive)"
// @Param        limit       query  int     false  "Page size (default 20)"
// @Param        offset      query  int     false  "Record offset (default 0)"
// @Success      200  {object}  listProblemsResponse "Paginated problem list"
// @Failure      401  {object}  errorResponse        "Missing or invalid JWT token"
// @Failure      500  {object}  errorResponse        "Internal server error"
// @Router       /problems [get]
func (h *ProblemHandler) ListProblems(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())
	q := r.URL.Query()

	params := service.ListProblemsParams{
		NameSearch:   q.Get("q"),
		Categories:   q["category"],
		Difficulties: q["difficulty"],
		Limit:        20, // default page size
	}

	// Parse optional score bounds.
	if v := q.Get("score_min"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			params.ScoreMin = &n
		}
	}
	if v := q.Get("score_max"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			params.ScoreMax = &n
		}
	}

	// Parse optional date range (YYYY-MM-DD).
	// 'from' is inclusive; 'to' is made exclusive by adding 1 day so the
	// selected end date is fully included.
	if v := q.Get("from"); v != "" {
		if t, err := time.Parse("2006-01-02", v); err == nil {
			params.DateFrom = &t
		}
	}
	if v := q.Get("to"); v != "" {
		if t, err := time.Parse("2006-01-02", v); err == nil {
			next := t.AddDate(0, 0, 1)
			params.DateTo = &next
		}
	}

	// Parse optional pagination overrides.
	if v := q.Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			params.Limit = n
		}
	}
	if v := q.Get("offset"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			params.Offset = n
		}
	}

	result, err := h.Service.ListFiltered(r.Context(), userID, params)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to fetch problems")
		return
	}

	problems := make([]problemResponse, len(result.Problems))
	for i, p := range result.Problems {
		problems[i] = toProblemResponse(p)
	}
	writeJSON(w, http.StatusOK, listProblemsResponse{
		Problems: problems,
		Total:    result.Total,
		Limit:    params.Limit,
		Offset:   params.Offset,
	})
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
		SolutionType:     req.SolutionType,
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
