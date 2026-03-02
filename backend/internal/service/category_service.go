package service

import (
	"context"
	"math"

	"github.com/apgupta3091/interview-iq/internal/models"
	"github.com/apgupta3091/interview-iq/internal/repository"
)

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

type CategoryService interface {
	GetStats(ctx context.Context, userID int) ([]models.CategoryStats, error)
	GetWeakest(ctx context.Context, userID int) (models.WeakestResult, error)
}

type categoryService struct {
	categories repository.CategoryRepository
}

func NewCategoryService(categories repository.CategoryRepository) CategoryService {
	return &categoryService{categories: categories}
}

func (s *categoryService) GetStats(ctx context.Context, userID int) ([]models.CategoryStats, error) {
	rawScores, err := s.categories.GetRawScoresByUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	type acc struct {
		total float64
		count int
	}
	buckets := map[string]*acc{}
	for _, rs := range rawScores {
		if _, ok := buckets[rs.Category]; !ok {
			buckets[rs.Category] = &acc{}
		}
		buckets[rs.Category].total += models.ApplyDecay(rs.Score, rs.SolvedAt)
		buckets[rs.Category].count++
	}

	stats := make([]models.CategoryStats, 0, len(buckets))
	for cat, a := range buckets {
		avg := min(math.Round((a.total/float64(a.count))*10)/10, 100)
		stats = append(stats, models.CategoryStats{
			Category:     cat,
			Strength:     avg,
			ProblemCount: a.count,
		})
	}
	return stats, nil
}

func (s *categoryService) GetWeakest(ctx context.Context, userID int) (models.WeakestResult, error) {
	stats, err := s.GetStats(ctx, userID)
	if err != nil {
		return models.WeakestResult{}, err
	}

	if len(stats) == 0 {
		return models.WeakestResult{}, ErrNoProblems
	}

	weakest := stats[0]
	for _, st := range stats[1:] {
		if st.Strength < weakest.Strength {
			weakest = st
		}
	}

	recs := recommendations[weakest.Category]
	if recs == nil {
		recs = []string{}
	}

	return models.WeakestResult{
		Category:        weakest.Category,
		Strength:        weakest.Strength,
		Recommendations: recs,
	}, nil
}
