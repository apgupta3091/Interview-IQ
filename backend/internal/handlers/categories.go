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
}

type weakestResponse struct {
	Category        string   `json:"category"`
	Strength        float64  `json:"strength"`
	Recommendations []string `json:"recommendations"`
}

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
		}
	}
	writeJSON(w, http.StatusOK, resp)
}

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
