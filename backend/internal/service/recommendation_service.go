package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/apgupta3091/interview-iq/internal/models"
)

// RecommendationParams controls which categories and history window to use.
type RecommendationParams struct {
	// Categories to generate recommendations for.
	// Empty means auto-select categories with strength < 60 (or the weakest one).
	Categories []string
	DateFrom   *time.Time
	DateTo     *time.Time
	// Limit is the number of problems per category; defaults to 3.
	Limit int
}

// ProblemRec is a single AI-recommended LeetCode problem.
type ProblemRec struct {
	Name        string `json:"name"`
	Difficulty  string `json:"difficulty"`
	Description string `json:"description"`
	// Reason explains why this problem was chosen based on the user's history and scores.
	Reason string `json:"reason"`
}

// CategoryRec groups AI recommendations for one category.
type CategoryRec struct {
	Category  string       `json:"category"`
	Strength  float64      `json:"strength"`
	FocusNote string       `json:"focus_note"`
	Problems  []ProblemRec `json:"problems"`
}

// RecommendationResult is the top-level AI recommendations response.
type RecommendationResult struct {
	Categories []CategoryRec `json:"categories"`
}

// RecommendationService generates AI-powered problem recommendations.
type RecommendationService interface {
	GetRecommendations(ctx context.Context, userID int, params RecommendationParams) (RecommendationResult, error)
	// InvalidateCache drops all cached results for userID so the next request
	// triggers a fresh OpenAI call. Call this whenever a user logs a new problem.
	InvalidateCache(userID int)
}

type recCacheEntry struct {
	result RecommendationResult
}

type recommendationService struct {
	categories CategoryService
	problems   ProblemService
	apiKey     string

	// cacheMu guards cache. The lock is not held during the OpenAI call so
	// concurrent requests for distinct keys can proceed in parallel.
	cacheMu sync.Mutex
	cache   map[string]recCacheEntry
}

// NewRecommendationService creates a RecommendationService backed by OpenAI gpt-4o-mini.
// apiKey must be a valid OpenAI API key; passing an empty string will cause all calls to fail
// with a clear error rather than panicking.
func NewRecommendationService(categories CategoryService, problems ProblemService, apiKey string) RecommendationService {
	return &recommendationService{
		categories: categories,
		problems:   problems,
		apiKey:     apiKey,
		cache:      make(map[string]recCacheEntry),
	}
}

// recCacheKey produces a deterministic string key for the given (userID, params) pair.
// Categories are sorted so order does not affect cache hits.
func recCacheKey(userID int, params RecommendationParams) string {
	cats := make([]string, len(params.Categories))
	copy(cats, params.Categories)
	sort.Strings(cats)

	from, to := "", ""
	if params.DateFrom != nil {
		from = params.DateFrom.UTC().Format(time.RFC3339)
	}
	if params.DateTo != nil {
		to = params.DateTo.UTC().Format(time.RFC3339)
	}
	return fmt.Sprintf("%d|%s|%d|%s|%s", userID, strings.Join(cats, ","), params.Limit, from, to)
}

// InvalidateCache removes all cached recommendation entries for the given user.
func (s *recommendationService) InvalidateCache(userID int) {
	prefix := fmt.Sprintf("%d|", userID)
	s.cacheMu.Lock()
	defer s.cacheMu.Unlock()
	for k := range s.cache {
		if strings.HasPrefix(k, prefix) {
			delete(s.cache, k)
		}
	}
}

func (s *recommendationService) GetRecommendations(ctx context.Context, userID int, params RecommendationParams) (RecommendationResult, error) {
	if s.apiKey == "" {
		return RecommendationResult{}, fmt.Errorf("recommendation service: OPENAI_API_KEY is not configured")
	}

	if params.Limit <= 0 {
		params.Limit = 3
	}

	// Return cached result if available. The cache is invalidated when the user
	// logs a new problem so recommendations stay fresh after score changes.
	cacheKey := recCacheKey(userID, params)
	s.cacheMu.Lock()
	if entry, ok := s.cache[cacheKey]; ok {
		s.cacheMu.Unlock()
		return entry.result, nil
	}
	s.cacheMu.Unlock()

	// 1. Load all category strengths so we can build meaningful context for the AI.
	stats, err := s.categories.GetStats(ctx, userID)
	if err != nil {
		return RecommendationResult{}, fmt.Errorf("recommendation service: get stats: %w", err)
	}

	statsByCategory := make(map[string]float64, len(stats))
	for _, st := range stats {
		statsByCategory[st.Category] = st.Strength
	}

	// 2. Determine which categories to target.
	targetCategories := params.Categories
	if len(targetCategories) == 0 {
		// Auto-select every category below the 60-point mastery threshold.
		for _, st := range stats {
			if st.Strength < 60 {
				targetCategories = append(targetCategories, st.Category)
			}
		}
		// If everything is strong (or there is no data), fall back to the weakest one.
		if len(targetCategories) == 0 && len(stats) > 0 {
			weakest := stats[0]
			for _, st := range stats[1:] {
				if st.Strength < weakest.Strength {
					weakest = st
				}
			}
			targetCategories = []string{weakest.Category}
		}
	}

	if len(targetCategories) == 0 {
		return RecommendationResult{}, ErrNoProblems
	}

	// 3. Fetch the user's practice history filtered to only the target categories.
	// This keeps the AI prompt focused and avoids leaking irrelevant history.
	listResult, err := s.problems.ListFiltered(ctx, userID, ListProblemsParams{
		Categories: targetCategories,
		DateFrom:   params.DateFrom,
		DateTo:     params.DateTo,
		Limit:      1000, // realistic per-user volume is well below this cap
	})
	if err != nil {
		return RecommendationResult{}, fmt.Errorf("recommendation service: list problems: %w", err)
	}

	// 4. Build a set of ALL attempted problem names for post-filtering.
	// Normalise to lowercase for case-insensitive matching.
	// All previously attempted problems are excluded so the user only sees fresh ones.
	attemptedNames := make(map[string]struct{}, len(listResult.Problems))
	for _, p := range listResult.Problems {
		attemptedNames[strings.ToLower(p.Name)] = struct{}{}
	}

	// 5. Build a structured prompt and call OpenAI.
	prompt := buildRecommendationPrompt(targetCategories, statsByCategory, listResult.Problems, attemptedNames, params.Limit)
	rawJSON, err := s.callOpenAI(ctx, prompt)
	if err != nil {
		return RecommendationResult{}, fmt.Errorf("recommendation service: %w", err)
	}

	// 6. Parse the AI JSON response into our typed struct.
	var result RecommendationResult
	if err := json.Unmarshal([]byte(rawJSON), &result); err != nil {
		return RecommendationResult{}, fmt.Errorf("recommendation service: parse AI response: %w", err)
	}

	// 7. Post-filter: remove any AI-recommended problem that the user has already
	// mastered (decayed score ≥ 75). The AI prompt instructs the same rule but
	// LLMs are not perfectly reliable, so we enforce it deterministically here.
	for i, catRec := range result.Categories {
		filtered := catRec.Problems[:0]
		for _, p := range catRec.Problems {
			if _, mastered := attemptedNames[strings.ToLower(p.Name)]; !mastered {
				filtered = append(filtered, p)
			}
		}
		result.Categories[i].Problems = filtered
	}

	// Back-fill strength from real data — the AI should not be trusted for exact numbers.
	for i, catRec := range result.Categories {
		if strength, ok := statsByCategory[catRec.Category]; ok {
			result.Categories[i].Strength = strength
		}
	}

	// Store in cache so subsequent requests skip the OpenAI call until the user
	// logs a new problem (which calls InvalidateCache).
	s.cacheMu.Lock()
	s.cache[cacheKey] = recCacheEntry{result: result}
	s.cacheMu.Unlock()

	return result, nil
}

// callOpenAI sends a single-turn chat completion to gpt-4o-mini and returns the raw assistant content.
func (s *recommendationService) callOpenAI(ctx context.Context, userPrompt string) (string, error) {
	// Inline struct types keep the OpenAI wire format self-contained within this function.
	type message struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	}
	type responseFormat struct {
		Type string `json:"type"`
	}
	type request struct {
		Model          string         `json:"model"`
		Messages       []message      `json:"messages"`
		Temperature    float64        `json:"temperature"`
		ResponseFormat responseFormat `json:"response_format"`
	}
	type response struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
		Error *struct {
			Message string `json:"message"`
		} `json:"error,omitempty"`
	}

	reqBody := request{
		Model: "gpt-4o-mini",
		Messages: []message{
			{
				Role:    "system",
				Content: "You are an interview prep coach. Return ONLY valid JSON — no markdown, no code fences, no explanation.",
			},
			{Role: "user", Content: userPrompt},
		},
		Temperature:    0.7,
		ResponseFormat: responseFormat{Type: "json_object"},
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("callOpenAI: marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		"https://api.openai.com/v1/chat/completions", bytes.NewReader(bodyBytes))
	if err != nil {
		return "", fmt.Errorf("callOpenAI: create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.apiKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("callOpenAI: http call: %w", err)
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("callOpenAI: read response: %w", err)
	}

	var parsed response
	if err := json.Unmarshal(respBytes, &parsed); err != nil {
		return "", fmt.Errorf("callOpenAI: parse response body: %w", err)
	}

	if parsed.Error != nil {
		return "", fmt.Errorf("callOpenAI: API error: %s", parsed.Error.Message)
	}

	if len(parsed.Choices) == 0 {
		return "", fmt.Errorf("callOpenAI: no choices in response")
	}

	// Strip markdown code fences if present — the model occasionally wraps JSON
	// in ```json...``` even when instructed not to.
	content := parsed.Choices[0].Message.Content
	content = strings.TrimSpace(content)
	if strings.HasPrefix(content, "```") {
		// Remove opening fence (```json or ```)
		if idx := strings.Index(content, "\n"); idx != -1 {
			content = content[idx+1:]
		}
		// Remove closing fence
		if idx := strings.LastIndex(content, "```"); idx != -1 {
			content = strings.TrimSpace(content[:idx])
		}
	}

	return content, nil
}

// buildRecommendationPrompt constructs the structured prompt sent to the AI.
// attemptedNames is a set of ALL lowercase problem names the user has ever attempted;
// the AI is instructed to skip them all and they are post-filtered in GetRecommendations.
func buildRecommendationPrompt(
	categories []string,
	statsByCategory map[string]float64,
	problems []models.Problem,
	attemptedNames map[string]struct{},
	limit int,
) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("Generate %d LeetCode problem recommendations ONLY for the exact categories listed below. Do not add, substitute, or expand to other categories.\n\n", limit))

	if len(categories) == 1 {
		sb.WriteString("You must return recommendations for exactly 1 category:\n")
	} else {
		sb.WriteString(fmt.Sprintf("You must return recommendations for exactly %d categories:\n", len(categories)))
	}
	for _, cat := range categories {
		sb.WriteString(fmt.Sprintf("  - %s: %.1f strength\n", cat, statsByCategory[cat]))
	}

	// Split practice history into mastered vs. still-needs-work so the AI has
	// clear context on what to skip and what gaps remain.
	var mastered, inProgress []models.Problem
	for _, p := range problems {
		if p.Score >= 75 {
			mastered = append(mastered, p)
		} else {
			inProgress = append(inProgress, p)
		}
	}

	// Build a combined "already attempted" list for the prompt, regardless of score.
	// This gives the AI full visibility into what the user has done.
	if len(mastered) > 0 {
		sb.WriteString("\nMastered problems — NEVER recommend any of these (score ≥ 75):\n")
		for _, p := range mastered {
			sb.WriteString(fmt.Sprintf("  - %s (score: %d)\n", p.Name, p.Score))
		}
	}

	if len(inProgress) > 0 {
		sb.WriteString("\nAlready attempted but not yet mastered (score < 75) — DO NOT recommend these either; the user needs NEW problems to discover fresh patterns:\n")
		for _, p := range inProgress {
			sb.WriteString(fmt.Sprintf("  - %s (score: %d)\n", p.Name, p.Score))
		}
	}

	if len(mastered) == 0 && len(inProgress) == 0 {
		sb.WriteString("\nThe user has no practice history in this category yet — recommend beginner-to-intermediate problems.\n")
	}

	sb.WriteString("\n")
	sb.WriteString("Rules:\n")
	sb.WriteString("  - Only include categories from the list above — no extras.\n")
	sb.WriteString("  - NEVER recommend any problem listed in either section above (mastered or already attempted).\n")
	sb.WriteString("  - ALL recommendations must be problems the user has NOT yet attempted.\n")
	sb.WriteString("  - Use the user's attempted problems and their scores as context to understand their skill level and gaps, but recommend only fresh, unattempted problems.\n")
	sb.WriteString(fmt.Sprintf("  - Recommend exactly %d problems per category.\n", limit))
	sb.WriteString("  - Each focus_note should be 2–3 sentences explaining what patterns to practise.\n")
	sb.WriteString("  - Each problem description should be 1 sentence explaining its value.\n")
	sb.WriteString("  - Each problem reason should be 1–2 sentences explaining why this NEW problem is the right next step given the user's history and weak areas. Reference their existing scores for context.\n\n")
	sb.WriteString("Respond with ONLY this JSON structure (no markdown, no code fences):\n")
	sb.WriteString("{\n")
	sb.WriteString("  \"categories\": [\n")
	sb.WriteString("    {\n")
	sb.WriteString("      \"category\": \"category-name\",\n")
	sb.WriteString("      \"focus_note\": \"...\",\n")
	sb.WriteString("      \"problems\": [\n")
	sb.WriteString("        {\n")
	sb.WriteString("          \"name\": \"Exact LeetCode problem title\",\n")
	sb.WriteString("          \"difficulty\": \"easy or medium or hard\",\n")
	sb.WriteString("          \"description\": \"...\",\n")
	sb.WriteString("          \"reason\": \"...\"\n")
	sb.WriteString("        }\n")
	sb.WriteString("      ]\n")
	sb.WriteString("    }\n")
	sb.WriteString("  ]\n")
	sb.WriteString("}\n")

	return sb.String()
}
