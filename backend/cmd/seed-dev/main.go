// Command seed-dev inserts a deterministic set of ~75 problem attempts for a
// dev seed user (clerk_user_id = "dev_seed_user") so the UI has realistic data
// without needing a real Clerk session.
//
// The script is idempotent: if the seed user already has problems it exits
// early with a message rather than inserting duplicates.
//
// Usage:
//
//	go run ./cmd/seed-dev          # from backend/
//	make seed-dev                  # from repo root
package main

import (
	"context"
	"log"
	"math/rand"
	"time"

	"github.com/apgupta3091/interview-iq/internal/db"
	"github.com/apgupta3091/interview-iq/internal/models"
	"github.com/apgupta3091/interview-iq/internal/repository"
)

const devClerkUserID = "dev_seed_user"

func main() {
	database, err := db.Connect()
	if err != nil {
		log.Fatalf("db connect: %v", err)
	}
	defer database.Close()

	if err := db.RunMigrations(database, "migrations"); err != nil {
		log.Fatalf("migrations: %v", err)
	}

	ctx := context.Background()
	userRepo := repository.NewUserRepo(database)
	problemRepo := repository.NewProblemRepo(database)

	// Upsert the dev seed user — safe to run repeatedly.
	userID, err := userRepo.GetOrCreateByClerkID(ctx, devClerkUserID)
	if err != nil {
		log.Fatalf("upsert dev user: %v", err)
	}
	log.Printf("dev seed user id = %d", userID)

	// Skip seeding if the user already has problems to stay idempotent.
	existing, err := problemRepo.ListByUser(ctx, userID)
	if err != nil {
		log.Fatalf("list existing problems: %v", err)
	}
	if len(existing) > 0 {
		log.Printf("seed user already has %d problems — skipping insert", len(existing))
		return
	}

	// Use a fixed seed so the generated data is deterministic across runs.
	rng := rand.New(rand.NewSource(42))

	now := time.Now()
	inserted := 0
	for _, p := range seedProblems {
		// Spread solved_at over the past 90 days to produce realistic decay.
		daysAgo := rng.Intn(90) + 1
		solvedAt := now.AddDate(0, 0, -daysAgo)

		score := models.ComputeScore(p.attempts, p.lookedAtSolution, p.solutionType)

		_, err := problemRepo.Insert(ctx, repository.InsertProblemParams{
			UserID:           userID,
			Name:             p.name,
			Categories:       p.categories,
			Difficulty:       p.difficulty,
			Attempts:         p.attempts,
			LookedAtSolution: p.lookedAtSolution,
			SolutionType:     p.solutionType,
			TimeTakenMins:    p.timeTakenMins,
			Score:            score,
			SolvedAt:         solvedAt,
		})
		if err != nil {
			log.Fatalf("insert %q: %v", p.name, err)
		}
		inserted++
	}

	log.Printf("seeded %d problems for dev user (id=%d)", inserted, userID)
}

// seedProblem holds the per-row data for each seeded problem attempt.
type seedProblem struct {
	name             string
	categories       []string
	difficulty       string
	attempts         int
	lookedAtSolution bool
	solutionType     string // "none" | "brute_force" | "optimal"
	timeTakenMins    int
}

// seedProblems is a hand-curated list of 75 real NeetCode problems covering
// all 17 supported categories with a realistic distribution of scores.
// Attempts, looked_at_solution, solution_type, and time_taken_mins are varied
// to produce a wide score range (5–100) across the dataset.
var seedProblems = []seedProblem{
	// ── Arrays ──────────────────────────────────────────────────────────────
	{"Two Sum", []string{"array", "hash-map"}, "easy", 1, false, "optimal", 12},
	{"Best Time to Buy and Sell Stock", []string{"array", "sliding-window"}, "easy", 1, false, "optimal", 10},
	{"Contains Duplicate", []string{"array", "hash-map"}, "easy", 1, false, "optimal", 8},
	{"Product of Array Except Self", []string{"array"}, "medium", 2, false, "optimal", 25},
	{"Maximum Subarray", []string{"array", "greedy"}, "medium", 2, false, "brute_force", 30},
	{"Maximum Product Subarray", []string{"array", "dp"}, "medium", 3, false, "brute_force", 35},
	{"Find Minimum in Rotated Sorted Array", []string{"array", "binary-search"}, "medium", 2, false, "optimal", 20},
	{"Search in Rotated Sorted Array", []string{"array", "binary-search"}, "medium", 3, true, "optimal", 40},
	{"3Sum", []string{"array", "two-pointers"}, "medium", 4, false, "optimal", 45},
	{"Container With Most Water", []string{"array", "two-pointers"}, "medium", 2, false, "optimal", 22},

	// ── Strings ─────────────────────────────────────────────────────────────
	{"Valid Anagram", []string{"string", "hash-map"}, "easy", 1, false, "optimal", 10},
	{"Valid Palindrome", []string{"string", "two-pointers"}, "easy", 1, false, "optimal", 8},
	{"Longest Substring Without Repeating Characters", []string{"string", "sliding-window"}, "medium", 2, false, "optimal", 25},
	{"Longest Repeating Character Replacement", []string{"string", "sliding-window"}, "medium", 3, false, "brute_force", 35},
	{"Minimum Window Substring", []string{"string", "sliding-window"}, "hard", 5, true, "optimal", 60},
	{"Group Anagrams", []string{"string", "hash-map"}, "medium", 2, false, "optimal", 20},
	{"Encode and Decode Strings", []string{"string"}, "medium", 2, false, "optimal", 18},
	{"Palindromic Substrings", []string{"string", "dp"}, "medium", 3, false, "brute_force", 30},
	{"Longest Palindromic Substring", []string{"string", "dp"}, "medium", 4, true, "optimal", 50},
	{"Valid Parentheses", []string{"string", "stack"}, "easy", 1, false, "optimal", 10},

	// ── Hash Map ─────────────────────────────────────────────────────────────
	{"Top K Frequent Elements", []string{"array", "hash-map", "heap"}, "medium", 2, false, "optimal", 22},
	{"Valid Sudoku", []string{"array", "hash-map"}, "medium", 3, false, "brute_force", 40},
	{"Longest Consecutive Sequence", []string{"array", "hash-map"}, "medium", 2, false, "optimal", 28},
	{"LRU Cache", []string{"linked-list", "hash-map"}, "medium", 4, true, "optimal", 55},

	// ── Two Pointers ─────────────────────────────────────────────────────────
	{"Two Sum II - Input Array Is Sorted", []string{"array", "two-pointers"}, "medium", 1, false, "optimal", 12},
	{"Trapping Rain Water", []string{"array", "two-pointers"}, "hard", 4, false, "optimal", 50},

	// ── Sliding Window ───────────────────────────────────────────────────────
	{"Permutation in String", []string{"string", "sliding-window"}, "medium", 3, false, "optimal", 30},
	{"Sliding Window Maximum", []string{"array", "sliding-window"}, "hard", 5, true, "brute_force", 65},

	// ── Binary Search ────────────────────────────────────────────────────────
	{"Binary Search", []string{"binary-search"}, "easy", 1, false, "optimal", 8},
	{"Search a 2D Matrix", []string{"binary-search"}, "medium", 2, false, "optimal", 18},
	{"Koko Eating Bananas", []string{"binary-search"}, "medium", 3, false, "brute_force", 30},
	{"Time Based Key-Value Store", []string{"binary-search"}, "medium", 3, false, "optimal", 28},
	{"Median of Two Sorted Arrays", []string{"binary-search"}, "hard", 5, true, "optimal", 70},

	// ── Stack ────────────────────────────────────────────────────────────────
	{"Min Stack", []string{"stack"}, "medium", 2, false, "optimal", 15},
	{"Evaluate Reverse Polish Notation", []string{"stack"}, "medium", 2, false, "optimal", 20},
	{"Generate Parentheses", []string{"stack", "backtracking"}, "medium", 3, false, "brute_force", 35},
	{"Daily Temperatures", []string{"stack"}, "medium", 2, false, "optimal", 22},
	{"Largest Rectangle in Histogram", []string{"stack"}, "hard", 5, true, "optimal", 75},

	// ── Linked List ──────────────────────────────────────────────────────────
	{"Reverse Linked List", []string{"linked-list"}, "easy", 1, false, "optimal", 10},
	{"Merge Two Sorted Lists", []string{"linked-list"}, "easy", 1, false, "optimal", 12},
	{"Reorder List", []string{"linked-list"}, "medium", 3, false, "optimal", 30},
	{"Remove Nth Node From End of List", []string{"linked-list"}, "medium", 2, false, "optimal", 20},
	{"Linked List Cycle", []string{"linked-list"}, "easy", 1, false, "optimal", 8},
	{"Merge k Sorted Lists", []string{"linked-list", "heap"}, "hard", 4, true, "optimal", 55},
	{"Reverse Nodes in K-Group", []string{"linked-list"}, "hard", 5, false, "brute_force", 60},

	// ── Tree ─────────────────────────────────────────────────────────────────
	{"Invert Binary Tree", []string{"tree"}, "easy", 1, false, "optimal", 8},
	{"Maximum Depth of Binary Tree", []string{"tree"}, "easy", 1, false, "optimal", 8},
	{"Balanced Binary Tree", []string{"tree"}, "easy", 2, false, "optimal", 15},
	{"Same Tree", []string{"tree"}, "easy", 1, false, "optimal", 10},
	{"Binary Tree Level Order Traversal", []string{"tree"}, "medium", 2, false, "optimal", 20},
	{"Validate Binary Search Tree", []string{"tree"}, "medium", 3, false, "optimal", 30},
	{"Kth Smallest Element in a BST", []string{"tree"}, "medium", 2, false, "optimal", 22},
	{"Lowest Common Ancestor of a BST", []string{"tree"}, "medium", 2, false, "optimal", 20},
	{"Binary Tree Maximum Path Sum", []string{"tree"}, "hard", 5, true, "optimal", 65},
	{"Serialize and Deserialize Binary Tree", []string{"tree"}, "hard", 4, true, "brute_force", 60},
	{"Construct Binary Tree from Preorder and Inorder Traversal", []string{"tree"}, "medium", 4, false, "optimal", 45},

	// ── Graph ────────────────────────────────────────────────────────────────
	{"Number of Islands", []string{"graph"}, "medium", 1, false, "optimal", 18},
	{"Clone Graph", []string{"graph"}, "medium", 2, false, "optimal", 25},
	{"Course Schedule", []string{"graph"}, "medium", 3, false, "optimal", 32},
	{"Pacific Atlantic Water Flow", []string{"graph"}, "medium", 4, false, "brute_force", 45},
	{"Word Ladder", []string{"graph"}, "hard", 5, true, "optimal", 70},

	// ── Heap ─────────────────────────────────────────────────────────────────
	{"Kth Largest Element in an Array", []string{"heap"}, "medium", 1, false, "optimal", 15},
	{"Task Scheduler", []string{"heap", "greedy"}, "medium", 3, false, "brute_force", 35},
	{"Find Median from Data Stream", []string{"heap"}, "hard", 4, true, "optimal", 55},

	// ── Dynamic Programming ──────────────────────────────────────────────────
	{"Climbing Stairs", []string{"dp"}, "easy", 1, false, "optimal", 10},
	{"House Robber", []string{"dp"}, "medium", 2, false, "optimal", 18},
	{"House Robber II", []string{"dp"}, "medium", 2, false, "optimal", 20},
	{"Coin Change", []string{"dp"}, "medium", 3, false, "brute_force", 35},
	{"Longest Increasing Subsequence", []string{"dp"}, "medium", 3, false, "optimal", 30},
	{"Word Break", []string{"dp"}, "medium", 4, true, "optimal", 45},
	{"Partition Equal Subset Sum", []string{"dp"}, "medium", 4, false, "brute_force", 40},

	// ── Backtracking ─────────────────────────────────────────────────────────
	{"Subsets", []string{"backtracking"}, "medium", 1, false, "optimal", 18},
	{"Combination Sum", []string{"backtracking"}, "medium", 2, false, "optimal", 25},
	{"Permutations", []string{"backtracking"}, "medium", 2, false, "optimal", 22},
	{"N-Queens", []string{"backtracking"}, "hard", 5, true, "brute_force", 75},

	// ── Greedy ───────────────────────────────────────────────────────────────
	{"Jump Game", []string{"greedy"}, "medium", 2, false, "optimal", 20},
	{"Jump Game II", []string{"greedy"}, "medium", 3, false, "optimal", 28},

	// ── Math ─────────────────────────────────────────────────────────────────
	{"Rotate Image", []string{"math"}, "medium", 2, false, "optimal", 20},
	{"Happy Number", []string{"math"}, "easy", 1, false, "optimal", 12},
}
