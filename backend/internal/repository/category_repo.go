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
	// UNNEST expands the categories array so each category gets its own row with
	// the problem's full score — multi-category problems contribute to every category.
	rows, err := r.db.QueryContext(ctx, `
		SELECT UNNEST(categories) AS category, score, solved_at
		FROM problems
		WHERE user_id = $1`,
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
