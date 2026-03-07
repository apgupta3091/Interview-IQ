package handlers

import (
	"net/http"
	"strconv"

	"github.com/apgupta3091/interview-iq/internal/repository"
)

type LeetCodeHandler struct {
	Repo repository.LeetCodeProblemRepository
}

type leetCodeProblemSuggestion struct {
	LcID       int      `json:"lc_id"`
	Title      string   `json:"title"`
	Slug       string   `json:"slug"`
	Difficulty string   `json:"difficulty"`
	Tags       []string `json:"tags"`
}

// Search handles GET /api/leetcode-problems/search?q=...&limit=10
func (h *LeetCodeHandler) Search(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	if q == "" {
		writeJSON(w, http.StatusOK, []leetCodeProblemSuggestion{})
		return
	}

	limit := 10
	if l := r.URL.Query().Get("limit"); l != "" {
		if n, err := strconv.Atoi(l); err == nil && n > 0 && n <= 50 {
			limit = n
		}
	}

	problems, err := h.Repo.Search(r.Context(), q, limit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "search failed")
		return
	}

	resp := make([]leetCodeProblemSuggestion, len(problems))
	for i, p := range problems {
		tags := p.Tags
		if tags == nil {
			tags = []string{}
		}
		resp[i] = leetCodeProblemSuggestion{
			LcID:       p.LcID,
			Title:      p.Title,
			Slug:       p.Slug,
			Difficulty: p.Difficulty,
			Tags:       tags,
		}
	}
	writeJSON(w, http.StatusOK, resp)
}
