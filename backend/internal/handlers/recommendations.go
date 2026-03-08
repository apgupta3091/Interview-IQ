package handlers

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/apgupta3091/interview-iq/internal/middleware"
	"github.com/apgupta3091/interview-iq/internal/service"
)

// RecommendationHandler handles AI-powered problem recommendation requests.
type RecommendationHandler struct {
	Service service.RecommendationService
}

// GetRecommendations returns AI-generated problem recommendations for the authenticated user.
//
// Query params:
//   - category[]  — one or more category slugs (repeatable); omit to auto-select weak categories
//   - from        — ISO date (YYYY-MM-DD); scopes practice history sent to the AI
//   - to          — ISO date (YYYY-MM-DD)
//   - limit       — problems per category, 1–10 (default 3)
func (h *RecommendationHandler) GetRecommendations(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())

	params := service.RecommendationParams{}

	// Parse repeated category params (supports both ?category=dp&category=tree and ?category=dp,tree).
	for _, raw := range r.URL.Query()["category"] {
		for _, cat := range strings.Split(raw, ",") {
			cat = strings.TrimSpace(cat)
			if cat != "" {
				params.Categories = append(params.Categories, cat)
			}
		}
	}

	if fromStr := r.URL.Query().Get("from"); fromStr != "" {
		if t, err := time.Parse("2006-01-02", fromStr); err == nil {
			params.DateFrom = &t
		}
	}

	if toStr := r.URL.Query().Get("to"); toStr != "" {
		if t, err := time.Parse("2006-01-02", toStr); err == nil {
			// Include the full end date by advancing to 23:59:59 of that day.
			end := t.Add(24*time.Hour - time.Second)
			params.DateTo = &end
		}
	}

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l >= 1 && l <= 10 {
			params.Limit = l
		}
	}

	result, err := h.Service.GetRecommendations(r.Context(), userID, params)
	if err != nil {
		var ve service.ValidationError
		if errors.As(err, &ve) {
			writeError(w, http.StatusBadRequest, ve.Message)
			return
		}
		if errors.Is(err, service.ErrNoProblems) {
			writeError(w, http.StatusNotFound, "no problems logged yet")
			return
		}
		slog.ErrorContext(r.Context(), "recommendations failed", "user_id", userID, "error", err)
		writeError(w, http.StatusInternalServerError, "failed to get recommendations")
		return
	}

	writeJSON(w, http.StatusOK, result)
}
