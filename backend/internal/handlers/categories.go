package handlers

import (
	"context"
	"database/sql"
	"math"
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

type weakestResponse struct {
	Category        string   `json:"category"`
	Strength        float64  `json:"strength"`
	Recommendations []string `json:"recommendations"`
}

// recommendations is a static map of category → 3 suggested LeetCode problems.
var recommendations = map[string][]string{
	"array":          {"Best Time to Buy and Sell Stock", "Product of Array Except Self", "Maximum Subarray"},
	"string":         {"Longest Substring Without Repeating Characters", "Valid Anagram", "Minimum Window Substring"},
	"hash-map":       {"Group Anagrams", "Top K Frequent Elements", "LRU Cache"},
	"two-pointers":   {"Container With Most Water", "3Sum", "Trapping Rain Water"},
	"sliding-window": {"Longest Repeating Character Replacement", "Permutation in String", "Minimum Size Subarray Sum"},
	"binary-search":  {"Find Minimum in Rotated Sorted Array", "Search in Rotated Sorted Array", "Koko Eating Bananas"},
	"stack":          {"Min Stack", "Daily Temperatures", "Largest Rectangle in Histogram"},
	"queue":          {"Sliding Window Maximum", "Design Circular Queue", "Task Scheduler"},
	"linked-list":    {"Reverse Linked List", "Merge Two Sorted Lists", "Linked List Cycle II"},
	"tree":           {"Binary Tree Level Order Traversal", "Validate Binary Search Tree", "Serialize and Deserialize Binary Tree"},
	"graph":          {"Number of Islands", "Clone Graph", "Course Schedule II"},
	"heap":           {"Find Median from Data Stream", "Merge K Sorted Lists", "Task Scheduler"},
	"dp":             {"Climbing Stairs", "Coin Change", "Longest Increasing Subsequence"},
	"backtracking":   {"Combination Sum", "Permutations", "N-Queens"},
	"greedy":         {"Jump Game", "Gas Station", "Partition Labels"},
	"math":           {"Reverse Integer", "Pow(x,n)", "Sieve of Eratosthenes"},
	"other":          {"LRU Cache", "Design Twitter", "Insert Delete GetRandom O(1)"},
}

// aggregateCategoryStrengths queries all problems for a user and returns
// a map of category → categoryStats with averaged decayed scores.
func aggregateCategoryStrengths(ctx context.Context, db *sql.DB, userID int) (map[string]categoryStats, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT category, score, solved_at
		FROM problems
		WHERE user_id = $1`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	type acc struct {
		total float64
		count int
	}
	buckets := map[string]*acc{}

	for rows.Next() {
		var category string
		var score int
		var solvedAt time.Time
		if err := rows.Scan(&category, &score, &solvedAt); err != nil {
			return nil, err
		}
		if _, ok := buckets[category]; !ok {
			buckets[category] = &acc{}
		}
		buckets[category].total += models.ApplyDecay(score, solvedAt)
		buckets[category].count++
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	result := make(map[string]categoryStats, len(buckets))
	for cat, a := range buckets {
		avg := min(math.Round((a.total/float64(a.count))*10)/10, 100)
		result[cat] = categoryStats{
			Category:     cat,
			Strength:     avg,
			ProblemCount: a.count,
		}
	}
	return result, nil
}

func (h *CategoryHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())

	statsMap, err := aggregateCategoryStrengths(r.Context(), h.DB, userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to fetch category stats")
		return
	}

	stats := make([]categoryStats, 0, len(statsMap))
	for _, s := range statsMap {
		stats = append(stats, s)
	}

	writeJSON(w, http.StatusOK, stats)
}

func (h *CategoryHandler) GetWeakest(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())

	statsMap, err := aggregateCategoryStrengths(r.Context(), h.DB, userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to fetch category stats")
		return
	}

	if len(statsMap) == 0 {
		writeError(w, http.StatusNotFound, "no problems logged yet")
		return
	}

	// find the category with the lowest strength
	var weakest categoryStats
	first := true
	for _, s := range statsMap {
		if first || s.Strength < weakest.Strength {
			weakest = s
			first = false
		}
	}

	recs := recommendations[weakest.Category]
	if recs == nil {
		recs = []string{}
	}

	writeJSON(w, http.StatusOK, weakestResponse{
		Category:        weakest.Category,
		Strength:        weakest.Strength,
		Recommendations: recs,
	})
}
