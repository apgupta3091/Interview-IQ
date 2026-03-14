package handlers

import (
	"errors"
	"net/http"

	"github.com/apgupta3091/interview-iq/internal/middleware"
	"github.com/apgupta3091/interview-iq/internal/service"
)

type CategoryHandler struct {
	Service service.CategoryService
}

type categoryStatsResponse struct {
	Category     string  `json:"category"`
	Strength     float64 `json:"strength"`
	ProblemCount int     `json:"problem_count"`
	// ScoreReady is false when the category has fewer than 3 problems and the
	// strength value should not be treated as a reliable score yet.
	ScoreReady bool `json:"score_ready"`
}

type weakestResponse struct {
	Category        string   `json:"category"`
	Strength        float64  `json:"strength"`
	Recommendations []string `json:"recommendations"`
}

// GetStats godoc
// @Summary      Get category strength scores
// @Description  Returns the average decayed score per category for the authenticated user. Strength is a 0–100 value representing current mastery. Categories with no problems are omitted.
// @Tags         categories
// @Produce      json
// @Security     BearerAuth
// @Success      200  {array}   categoryStatsResponse "Per-category strength scores"
// @Failure      401  {object}  errorResponse         "Missing or invalid JWT token"
// @Failure      500  {object}  errorResponse         "Internal server error"
// @Router       /categories/stats [get]
func (h *CategoryHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())

	stats, err := h.Service.GetStats(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to fetch category stats")
		return
	}

	resp := make([]categoryStatsResponse, len(stats))
	for i, s := range stats {
		resp[i] = categoryStatsResponse{
			Category:     s.Category,
			Strength:     s.Strength,
			ProblemCount: s.ProblemCount,
			ScoreReady:   s.ScoreReady,
		}
	}
	writeJSON(w, http.StatusOK, resp)
}

// GetWeakest godoc
// @Summary      Get weakest category with recommendations
// @Description  Identifies the category with the lowest average decayed score and returns 3 curated LeetCode problems to practice. Used to drive the dashboard weakness banner.
// @Tags         categories
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  weakestResponse "Weakest category with 3 recommended problems"
// @Failure      401  {object}  errorResponse   "Missing or invalid JWT token"
// @Failure      404  {object}  errorResponse   "No problems logged yet"
// @Failure      500  {object}  errorResponse   "Internal server error"
// @Router       /categories/weakest [get]
func (h *CategoryHandler) GetWeakest(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())

	result, err := h.Service.GetWeakest(r.Context(), userID)
	if err != nil {
		if errors.Is(err, service.ErrNoProblems) {
			writeError(w, http.StatusNotFound, "no problems logged yet")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to fetch weakest category")
		return
	}

	writeJSON(w, http.StatusOK, weakestResponse{
		Category:        result.Category,
		Strength:        result.Strength,
		Recommendations: result.Recommendations,
	})
}
