package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
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
}

type recommendationService struct {
	categories CategoryService
	problems   ProblemService
	apiKey     string
}

// NewRecommendationService creates a RecommendationService backed by OpenAI gpt-4o-mini.
// apiKey must be a valid OpenAI API key; passing an empty string will cause all calls to fail
// with a clear error rather than panicking.
func NewRecommendationService(categories CategoryService, problems ProblemService, apiKey string) RecommendationService {
	return &recommendationService{
		categories: categories,
		problems:   problems,
		apiKey:     apiKey,
	}
}

func (s *recommendationService) GetRecommendations(ctx context.Context, userID int, params RecommendationParams) (RecommendationResult, error) {
	if s.apiKey == "" {
		return RecommendationResult{}, fmt.Errorf("recommendation service: OPENAI_API_KEY is not configured")
	}

	if params.Limit <= 0 {
		params.Limit = 3
	}

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

	// 3. Fetch the user's practice history so the AI can avoid over-recommending mastered problems.
	listResult, err := s.problems.ListFiltered(ctx, userID, ListProblemsParams{
		DateFrom: params.DateFrom,
		DateTo:   params.DateTo,
		Limit:    1000, // realistic per-user volume is well below this cap
	})
	if err != nil {
		return RecommendationResult{}, fmt.Errorf("recommendation service: list problems: %w", err)
	}

	// 4. Build a structured prompt and call OpenAI.
	prompt := buildRecommendationPrompt(targetCategories, statsByCategory, listResult.Problems, params.Limit)
	rawJSON, err := s.callOpenAI(ctx, prompt)
	if err != nil {
		return RecommendationResult{}, fmt.Errorf("recommendation service: %w", err)
	}

	// 5. Parse the AI JSON response into our typed struct.
	var result RecommendationResult
	if err := json.Unmarshal([]byte(rawJSON), &result); err != nil {
		return RecommendationResult{}, fmt.Errorf("recommendation service: parse AI response: %w", err)
	}

	// Back-fill strength from real data — the AI should not be trusted for exact numbers.
	for i, catRec := range result.Categories {
		if strength, ok := statsByCategory[catRec.Category]; ok {
			result.Categories[i].Strength = strength
		}
	}

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
func buildRecommendationPrompt(
	categories []string,
	statsByCategory map[string]float64,
	problems []models.Problem,
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

	if len(problems) > 0 {
		sb.WriteString("\nUser's practice history (problem name → decayed score):\n")
		for _, p := range problems {
			sb.WriteString(fmt.Sprintf("  - %s → %d\n", p.Name, int(p.DecayedScore)))
		}
	}

	sb.WriteString("\n")
	sb.WriteString("Rules:\n")
	sb.WriteString("  - Only include categories from the list above — no extras.\n")
	sb.WriteString("  - Prefer problems the user hasn't attempted yet.\n")
	sb.WriteString("  - Never recommend a problem the user has already attempted with a decayed score above 80.\n")
	sb.WriteString(fmt.Sprintf("  - Recommend exactly %d problems per category.\n", limit))
	sb.WriteString("  - Each focus_note should be 2–3 sentences explaining what patterns to practise.\n")
	sb.WriteString("  - Each problem description should be 1 sentence explaining its value.\n\n")
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
	sb.WriteString("          \"description\": \"...\"\n")
	sb.WriteString("        }\n")
	sb.WriteString("      ]\n")
	sb.WriteString("    }\n")
	sb.WriteString("  ]\n")
	sb.WriteString("}\n")

	return sb.String()
}
