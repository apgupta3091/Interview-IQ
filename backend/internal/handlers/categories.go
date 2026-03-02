package handlers

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/apgupta3091/interview-iq/internal/middleware"
	"github.com/apgupta3091/interview-iq/internal/models"
)

type CategoryHandler struct {
	DB *sql.DB
}

type categoryStats struct {
	Category     string  `json:"category"`
	Strength     float64 `json:"strength"`
	ProblemCount int     `json:"problem_count"`
}

func (h *CategoryHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())

	rows, err := h.DB.QueryContext(r.Context(), `
		SELECT category, score, solved_at
		FROM problems
		WHERE user_id = $1`,
		userID,
	)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to fetch category stats")
		return
	}
	defer rows.Close()

	// accumulate decayed scores per category
	type accumulator struct {
		total float64
		count int
	}
	acc := map[string]*accumulator{}

	for rows.Next() {
		var category string
		var score int
		var solvedAt time.Time
		if err := rows.Scan(&category, &score, &solvedAt); err != nil {
			writeError(w, http.StatusInternalServerError, "failed to read stats")
			return
		}
		if _, ok := acc[category]; !ok {
			acc[category] = &accumulator{}
		}
		acc[category].total += models.ApplyDecay(score, solvedAt)
		acc[category].count++
	}
	if err := rows.Err(); err != nil {
		writeError(w, http.StatusInternalServerError, "error iterating stats")
		return
	}

	stats := []categoryStats{}
	for cat, a := range acc {
		avg := a.total / float64(a.count)
		// round to 1 decimal and cap at 100
		avg = float64(int(avg*10+0.5)) / 10
		if avg > 100 {
			avg = 100
		}
		stats = append(stats, categoryStats{
			Category:     cat,
			Strength:     avg,
			ProblemCount: a.count,
		})
	}

	writeJSON(w, http.StatusOK, stats)
}
