package repository

import (
	"context"
	"database/sql"

	"github.com/apgupta3091/interview-iq/internal/models"
)

type CategoryRepository interface {
	GetRawScoresByUser(ctx context.Context, userID int) ([]models.CategoryRawScore, error)
}

type sqlCategoryRepo struct {
	db *sql.DB
}

func NewCategoryRepo(db *sql.DB) CategoryRepository {
	return &sqlCategoryRepo{db: db}
}

func (r *sqlCategoryRepo) GetRawScoresByUser(ctx context.Context, userID int) ([]models.CategoryRawScore, error) {
	// DISTINCT ON (name) picks the most recent log per problem name (ordered by
	// created_at DESC), so a user who retries and improves is scored on their
	// latest attempt rather than an average of all attempts.
	// UNNEST then expands categories so each category gets its own row.
	rows, err := r.db.QueryContext(ctx, `
		SELECT UNNEST(categories) AS category, score, solved_at
		FROM (
			SELECT DISTINCT ON (name) name, categories, score, solved_at
			FROM problems
			WHERE user_id = $1
			ORDER BY name, created_at DESC
		) latest`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var scores []models.CategoryRawScore
	for rows.Next() {
		var s models.CategoryRawScore
		if err := rows.Scan(&s.Category, &s.Score, &s.SolvedAt); err != nil {
			return nil, err
		}
		scores = append(scores, s)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return scores, nil
}
