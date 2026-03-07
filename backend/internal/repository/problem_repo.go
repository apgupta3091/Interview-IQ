package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/lib/pq"

	"github.com/apgupta3091/interview-iq/internal/models"
)

type InsertProblemParams struct {
	UserID, Attempts, TimeTakenMins, Score int
	Name, Difficulty, SolutionType         string
	Categories                             []string
	LookedAtSolution                       bool
	SolvedAt                               time.Time
}

type ProblemRepository interface {
	Insert(ctx context.Context, p InsertProblemParams) (models.Problem, error)
	ListByUser(ctx context.Context, userID int) ([]models.Problem, error)
}

type sqlProblemRepo struct {
	db *sql.DB
}

func NewProblemRepo(db *sql.DB) ProblemRepository {
	return &sqlProblemRepo{db: db}
}

func (r *sqlProblemRepo) Insert(ctx context.Context, p InsertProblemParams) (models.Problem, error) {
	var prob models.Problem
	err := r.db.QueryRowContext(ctx, `
		INSERT INTO problems
			(user_id, name, categories, difficulty, attempts, looked_at_solution, time_taken_mins, score, solved_at, solution_type)
		VALUES
			($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id, name, categories, difficulty, attempts, looked_at_solution, time_taken_mins, score, solved_at, created_at, solution_type`,
		p.UserID, p.Name, pq.Array(p.Categories), p.Difficulty,
		p.Attempts, p.LookedAtSolution, p.TimeTakenMins,
		p.Score, p.SolvedAt, p.SolutionType,
	).Scan(
		&prob.ID, &prob.Name, pq.Array(&prob.Categories), &prob.Difficulty,
		&prob.Attempts, &prob.LookedAtSolution, &prob.TimeTakenMins,
		&prob.Score, &prob.SolvedAt, &prob.CreatedAt, &prob.SolutionType,
	)
	if err != nil {
		return models.Problem{}, err
	}
	prob.UserID = p.UserID
	return prob, nil
}

func (r *sqlProblemRepo) ListByUser(ctx context.Context, userID int) ([]models.Problem, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, name, categories, difficulty, attempts, looked_at_solution,
		       time_taken_mins, score, solved_at, created_at, solution_type
		FROM problems
		WHERE user_id = $1
		ORDER BY created_at DESC`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var problems []models.Problem
	for rows.Next() {
		var p models.Problem
		p.UserID = userID
		if err := rows.Scan(
			&p.ID, &p.Name, pq.Array(&p.Categories), &p.Difficulty,
			&p.Attempts, &p.LookedAtSolution, &p.TimeTakenMins,
			&p.Score, &p.SolvedAt, &p.CreatedAt, &p.SolutionType,
		); err != nil {
			return nil, err
		}
		problems = append(problems, p)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return problems, nil
}
