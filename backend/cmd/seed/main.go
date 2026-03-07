// Command seed fetches the public LeetCode problem list, applies the hardcoded
// NeetCode category map, and bulk-upserts into the leetcode_problems table.
//
// Usage:
//
//	go run ./cmd/seed
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/apgupta3091/interview-iq/internal/db"
	"github.com/apgupta3091/interview-iq/internal/models"
	"github.com/apgupta3091/interview-iq/internal/repository"
)

func main() {
	database, err := db.Connect()
	if err != nil {
		log.Fatalf("db connect: %v", err)
	}
	defer database.Close()

	if err := db.RunMigrations(database, "migrations"); err != nil {
		log.Fatalf("migrations: %v", err)
	}

	log.Println("fetching LeetCode problem list…")
	lcProblems, err := fetchLeetCodeProblems()
	if err != nil {
		log.Fatalf("fetch: %v", err)
	}
	log.Printf("fetched %d problems", len(lcProblems))

	repo := repository.NewLeetCodeProblemRepo(database)
	ctx := context.Background()

	if err := repo.BulkUpsert(ctx, lcProblems); err != nil {
		log.Fatalf("upsert: %v", err)
	}
	log.Printf("upserted %d problems into leetcode_problems", len(lcProblems))
}

// lcAPIResponse mirrors the top-level shape of https://leetcode.com/api/problems/all/
type lcAPIResponse struct {
	StatStatusPairs []struct {
		Stat struct {
			FrontendQuestionID int    `json:"frontend_question_id"`
			QuestionTitle      string `json:"question__title"`
			QuestionTitleSlug  string `json:"question__title_slug"`
		} `json:"stat"`
		Difficulty struct {
			Level int `json:"level"` // 1=Easy, 2=Medium, 3=Hard
		} `json:"difficulty"`
		PaidOnly bool `json:"paid_only"`
	} `json:"stat_status_pairs"`
}

func fetchLeetCodeProblems() ([]models.LeetCodeProblem, error) {
	apiURL := "https://leetcode.com/api/problems/all/"
	req, err := http.NewRequest(http.MethodGet, apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("User-Agent", "interview-iq-seed/1.0")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http get: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d", resp.StatusCode)
	}

	var apiResp lcAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("decode: %w", err)
	}

	diffMap := map[int]string{1: "easy", 2: "medium", 3: "hard"}

	problems := make([]models.LeetCodeProblem, 0, len(apiResp.StatStatusPairs))
	for _, p := range apiResp.StatStatusPairs {
		diff, ok := diffMap[p.Difficulty.Level]
		if !ok {
			diff = "medium"
		}
		tags := neetcodeTagMap[p.Stat.QuestionTitleSlug]
		if tags == nil {
			tags = []string{}
		}
		problems = append(problems, models.LeetCodeProblem{
			LcID:       p.Stat.FrontendQuestionID,
			Title:      p.Stat.QuestionTitle,
			Slug:       p.Stat.QuestionTitleSlug,
			Difficulty: diff,
			Tags:       tags,
			PaidOnly:   p.PaidOnly,
		})
	}
	return problems, nil
}

// neetcodeTagMap maps LeetCode problem slugs to our category labels.
// Covers the ~150 most common NeetCode problems.
var neetcodeTagMap = map[string][]string{
	// Arrays & Hashing
	"contains-duplicate":                {"array", "hash-map"},
	"valid-anagram":                     {"string", "hash-map"},
	"two-sum":                           {"array", "hash-map"},
	"group-anagrams":                    {"string", "hash-map"},
	"top-k-frequent-elements":           {"array", "hash-map", "heap"},
	"encode-and-decode-strings":         {"string"},
	"product-of-array-except-self":      {"array"},
	"valid-sudoku":                      {"array", "hash-map"},
	"longest-consecutive-sequence":      {"array", "hash-map"},

	// Two Pointers
	"valid-palindrome":                  {"string", "two-pointers"},
	"two-sum-ii-input-array-is-sorted":  {"array", "two-pointers"},
	"3sum":                              {"array", "two-pointers"},
	"container-with-most-water":         {"array", "two-pointers"},
	"trapping-rain-water":               {"array", "two-pointers"},

	// Sliding Window
	"best-time-to-buy-and-sell-stock":          {"array", "sliding-window"},
	"longest-substring-without-repeating-characters": {"string", "sliding-window"},
	"longest-repeating-character-replacement":  {"string", "sliding-window"},
	"permutation-in-string":                    {"string", "sliding-window"},
	"minimum-window-substring":                 {"string", "sliding-window"},
	"sliding-window-maximum":                   {"array", "sliding-window"},

	// Stack
	"valid-parentheses":                 {"stack"},
	"min-stack":                         {"stack"},
	"evaluate-reverse-polish-notation":  {"stack"},
	"generate-parentheses":              {"stack", "backtracking"},
	"daily-temperatures":                {"stack"},
	"car-fleet":                         {"stack"},
	"largest-rectangle-in-histogram":    {"stack"},

	// Binary Search
	"binary-search":                     {"binary-search"},
	"search-a-2d-matrix":               {"binary-search"},
	"koko-eating-bananas":              {"binary-search"},
	"find-minimum-in-rotated-sorted-array": {"binary-search"},
	"search-in-rotated-sorted-array":   {"binary-search"},
	"time-based-key-value-store":        {"binary-search"},
	"median-of-two-sorted-arrays":       {"binary-search"},

	// Linked List
	"reverse-linked-list":              {"linked-list"},
	"merge-two-sorted-lists":           {"linked-list"},
	"reorder-list":                     {"linked-list"},
	"remove-nth-node-from-end-of-list": {"linked-list"},
	"copy-list-with-random-pointer":    {"linked-list"},
	"add-two-numbers":                  {"linked-list", "math"},
	"linked-list-cycle":                {"linked-list"},
	"find-the-duplicate-number":        {"array", "linked-list"},
	"lru-cache":                        {"linked-list", "hash-map"},
	"merge-k-sorted-lists":             {"linked-list", "heap"},
	"reverse-nodes-in-k-group":         {"linked-list"},

	// Trees
	"invert-binary-tree":               {"tree"},
	"maximum-depth-of-binary-tree":     {"tree"},
	"diameter-of-binary-tree":          {"tree"},
	"balanced-binary-tree":             {"tree"},
	"same-tree":                        {"tree"},
	"subtree-of-another-tree":          {"tree"},
	"lowest-common-ancestor-of-a-binary-search-tree": {"tree"},
	"binary-tree-level-order-traversal": {"tree"},
	"binary-tree-right-side-view":      {"tree"},
	"count-good-nodes-in-binary-tree":  {"tree"},
	"validate-binary-search-tree":      {"tree"},
	"kth-smallest-element-in-a-bst":    {"tree"},
	"construct-binary-tree-from-preorder-and-inorder-traversal": {"tree"},
	"binary-tree-maximum-path-sum":     {"tree"},
	"serialize-and-deserialize-binary-tree": {"tree"},

	// Trie
	"implement-trie-prefix-tree":       {"trie"},
	"design-add-and-search-words-data-structure": {"trie"},
	"word-search-ii":                   {"trie", "backtracking"},

	// Heap / Priority Queue
	"kth-largest-element-in-a-stream":  {"heap"},
	"last-stone-weight":                {"heap"},
	"k-closest-points-to-origin":       {"heap"},
	"kth-largest-element-in-an-array":  {"heap"},
	"task-scheduler":                   {"heap", "greedy"},
	"design-twitter":                   {"heap", "linked-list"},
	"find-median-from-data-stream":     {"heap"},

	// Backtracking
	"subsets":                          {"backtracking"},
	"combination-sum":                  {"backtracking"},
	"permutations":                     {"backtracking"},
	"subsets-ii":                       {"backtracking"},
	"combination-sum-ii":               {"backtracking"},
	"word-search":                      {"backtracking", "graph"},
	"palindrome-partitioning":          {"backtracking", "dp"},
	"letter-combinations-of-a-phone-number": {"backtracking"},
	"n-queens":                         {"backtracking"},

	// Graphs
	"number-of-islands":                {"graph"},
	"clone-graph":                      {"graph"},
	"max-area-of-island":               {"graph"},
	"pacific-atlantic-water-flow":      {"graph"},
	"surrounded-regions":               {"graph"},
	"rotting-oranges":                  {"graph"},
	"walls-and-gates":                  {"graph"},
	"course-schedule":                  {"graph"},
	"course-schedule-ii":               {"graph"},
	"redundant-connection":             {"graph"},
	"number-of-connected-components-in-an-undirected-graph": {"graph"},
	"graph-valid-tree":                 {"graph"},
	"word-ladder":                      {"graph"},

	// Advanced Graphs
	"reconstruct-itinerary":            {"advanced-graphs"},
	"min-cost-to-connect-all-points":   {"advanced-graphs"},
	"network-delay-time":               {"advanced-graphs"},
	"swim-in-rising-water":             {"advanced-graphs"},
	"alien-dictionary":                 {"advanced-graphs"},
	"cheapest-flights-within-k-stops":  {"advanced-graphs"},

	// Dynamic Programming 1D
	"climbing-stairs":                  {"dp"},
	"min-cost-climbing-stairs":         {"dp"},
	"house-robber":                     {"dp"},
	"house-robber-ii":                  {"dp"},
	"longest-palindromic-substring":    {"dp", "string"},
	"palindromic-substrings":           {"dp", "string"},
	"decode-ways":                      {"dp"},
	"coin-change":                      {"dp"},
	"maximum-product-subarray":         {"dp", "array"},
	"word-break":                       {"dp"},
	"longest-increasing-subsequence":   {"dp"},
	"partition-equal-subset-sum":       {"dp"},

	// Dynamic Programming 2D
	"unique-paths":                     {"dp-2d"},
	"longest-common-subsequence":       {"dp-2d", "string"},
	"best-time-to-buy-and-sell-stock-with-cooldown": {"dp-2d"},
	"coin-change-ii":                   {"dp-2d"},
	"target-sum":                       {"dp-2d"},
	"interleaving-string":              {"dp-2d", "string"},
	"longest-increasing-path-in-a-matrix": {"dp-2d", "graph"},
	"distinct-subsequences":            {"dp-2d", "string"},
	"edit-distance":                    {"dp-2d", "string"},
	"burst-balloons":                   {"dp-2d"},
	"regular-expression-matching":      {"dp-2d", "string"},

	// Greedy
	"maximum-subarray":                 {"greedy", "dp"},
	"jump-game":                        {"greedy"},
	"jump-game-ii":                     {"greedy"},
	"gas-station":                      {"greedy"},
	"hand-of-straights":                {"greedy"},
	"merge-triplets-to-form-target-triplet": {"greedy"},
	"partition-labels":                 {"greedy", "string"},
	"valid-parenthesis-string":         {"greedy", "stack"},

	// Intervals
	"insert-interval":                  {"intervals"},
	"merge-intervals":                  {"intervals"},
	"non-overlapping-intervals":        {"intervals"},
	"meeting-rooms":                    {"intervals"},
	"meeting-rooms-ii":                 {"intervals"},
	"minimum-interval-to-include-each-query": {"intervals"},

	// Math & Geometry
	"rotate-image":                     {"math"},
	"spiral-matrix":                    {"math"},
	"set-matrix-zeroes":                {"math"},
	"happy-number":                     {"math"},
	"plus-one":                         {"math"},
	"pow-x-n":                          {"math"},
	"multiply-strings":                 {"math", "string"},
	"detect-squares":                   {"math"},

	// Bit Manipulation
	"single-number":                    {"bit-manipulation"},
	"number-of-1-bits":                 {"bit-manipulation"},
	"counting-bits":                    {"bit-manipulation", "dp"},
	"reverse-bits":                     {"bit-manipulation"},
	"missing-number":                   {"bit-manipulation"},
	"sum-of-two-integers":              {"bit-manipulation"},
	"reverse-integer":                  {"bit-manipulation", "math"},
}

// init validates that all tag values in neetcodeTagMap are known categories.
// This runs at startup and panics if someone adds an unknown category slug.
func init() {
	known := map[string]bool{
		"array": true, "string": true, "hash-map": true, "two-pointers": true,
		"sliding-window": true, "binary-search": true, "stack": true, "queue": true,
		"linked-list": true, "tree": true, "trie": true, "graph": true,
		"advanced-graphs": true, "heap": true, "dp": true, "dp-2d": true,
		"backtracking": true, "greedy": true, "intervals": true,
		"math": true, "bit-manipulation": true, "other": true,
	}
	for slug, tags := range neetcodeTagMap {
		for _, t := range tags {
			if !known[t] {
				_, _ = fmt.Fprintf(os.Stderr, "seed: unknown category %q for slug %q\n", t, slug)
				os.Exit(1)
			}
		}
	}
}
