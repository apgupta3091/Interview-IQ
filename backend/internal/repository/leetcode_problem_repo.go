package repository

import (
	"context"
	"database/sql"

	"github.com/lib/pq"

	"github.com/apgupta3091/interview-iq/internal/models"
)

type LeetCodeProblemRepository interface {
	// Search returns up to limit problems whose title matches q (full-text or ILIKE).
	Search(ctx context.Context, q string, limit int) ([]models.LeetCodeProblem, error)
	// BulkUpsert inserts or updates problems by lc_id (used by the seed command).
	BulkUpsert(ctx context.Context, problems []models.LeetCodeProblem) error
}

type sqlLeetCodeProblemRepo struct {
	db *sql.DB
}

func NewLeetCodeProblemRepo(db *sql.DB) LeetCodeProblemRepository {
	return &sqlLeetCodeProblemRepo{db: db}
}

func (r *sqlLeetCodeProblemRepo) Search(ctx context.Context, q string, limit int) ([]models.LeetCodeProblem, error) {
	// Use full-text search for whole-word matches and ILIKE as a fallback for
	// partial-prefix matches (e.g. "two" matches "Two Sum").
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, lc_id, title, slug, difficulty, tags, paid_only
		FROM leetcode_problems
		WHERE
			to_tsvector('english', title) @@ plainto_tsquery('english', $1)
			OR title ILIKE '%' || $1 || '%'
		ORDER BY
			(to_tsvector('english', title) @@ plainto_tsquery('english', $1)) DESC,
			lc_id ASC
		LIMIT $2`,
		q, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []models.LeetCodeProblem
	for rows.Next() {
		var p models.LeetCodeProblem
		if err := rows.Scan(
			&p.ID, &p.LcID, &p.Title, &p.Slug,
			&p.Difficulty, pq.Array(&p.Tags), &p.PaidOnly,
		); err != nil {
			return nil, err
		}
		results = append(results, p)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return results, nil
}

func (r *sqlLeetCodeProblemRepo) BulkUpsert(ctx context.Context, problems []models.LeetCodeProblem) error {
	for _, p := range problems {
		_, err := r.db.ExecContext(ctx, `
			INSERT INTO leetcode_problems (lc_id, title, slug, difficulty, tags, paid_only)
			VALUES ($1, $2, $3, $4, $5, $6)
			ON CONFLICT (lc_id) DO UPDATE SET
				title      = EXCLUDED.title,
				slug       = EXCLUDED.slug,
				difficulty = EXCLUDED.difficulty,
				tags       = EXCLUDED.tags,
				paid_only  = EXCLUDED.paid_only`,
			p.LcID, p.Title, p.Slug, p.Difficulty, pq.Array(p.Tags), p.PaidOnly,
		)
		if err != nil {
			return err
		}
	}
	return nil
}
