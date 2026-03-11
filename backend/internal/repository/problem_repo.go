package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/lib/pq"

	"github.com/apgupta3091/interview-iq/internal/models"
)

type InsertProblemParams struct {
	UserID, Attempts, TimeTakenMins, Score, OriginalScore int
	Name, Difficulty, SolutionType                        string
	Notes                                                 string
	Categories                                            []string
	LookedAtSolution                                      bool
	SolvedAt                                              time.Time
}

// ListProblemsFilter holds optional server-side filter criteria for ListByUserFiltered.
// Nil pointer fields are ignored (not applied to the query).
type ListProblemsFilter struct {
	NameSearch   string     // ILIKE match on problem name
	Categories   []string   // overlap match: problem must have at least one of these categories
	Difficulties []string   // exact match: difficulty must be one of these values
	ScoreMin     *int       // inclusive lower bound on raw score
	ScoreMax     *int       // inclusive upper bound on raw score
	DateFrom     *time.Time // inclusive lower bound on created_at
	DateTo       *time.Time // exclusive upper bound on created_at (caller adds 1 day for inclusive end-date)
	Limit        int        // page size; defaults to 20 if ≤ 0
	Offset       int        // zero-based record offset
}

// ListProblemsResult holds a page of problems along with the total unfiltered-but-user-scoped count.
type ListProblemsResult struct {
	Problems []models.Problem
	Total    int
}

type ProblemRepository interface {
	Insert(ctx context.Context, p InsertProblemParams) (models.Problem, error)
	GetByID(ctx context.Context, id, userID int) (models.Problem, error)
	ListByUser(ctx context.Context, userID int) ([]models.Problem, error)
	ListByUserFiltered(ctx context.Context, userID int, f ListProblemsFilter) (ListProblemsResult, error)
	// DecayAllProblems updates score = ApplyDecay(original_score, solved_at) for every row.
	// Uses now as the reference time so the caller controls when "today" is.
	DecayAllProblems(ctx context.Context, now time.Time) (int64, error)
}

type sqlProblemRepo struct {
	db *sql.DB
}

func NewProblemRepo(db *sql.DB) ProblemRepository {
	return &sqlProblemRepo{db: db}
}

func (r *sqlProblemRepo) Insert(ctx context.Context, p InsertProblemParams) (models.Problem, error) {
	var prob models.Problem
	var notes sql.NullString
	notesVal := sql.NullString{String: p.Notes, Valid: p.Notes != ""}
	err := r.db.QueryRowContext(ctx, `
		INSERT INTO problems
			(user_id, name, categories, difficulty, attempts, looked_at_solution, time_taken_mins, score, original_score, solved_at, solution_type, notes)
		VALUES
			($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		RETURNING id, name, categories, difficulty, attempts, looked_at_solution, time_taken_mins, score, original_score, solved_at, created_at, solution_type, notes`,
		p.UserID, p.Name, pq.Array(p.Categories), p.Difficulty,
		p.Attempts, p.LookedAtSolution, p.TimeTakenMins,
		p.Score, p.OriginalScore, p.SolvedAt, p.SolutionType, notesVal,
	).Scan(
		&prob.ID, &prob.Name, pq.Array(&prob.Categories), &prob.Difficulty,
		&prob.Attempts, &prob.LookedAtSolution, &prob.TimeTakenMins,
		&prob.Score, &prob.OriginalScore, &prob.SolvedAt, &prob.CreatedAt, &prob.SolutionType,
		&notes,
	)
	if err != nil {
		return models.Problem{}, err
	}
	prob.UserID = p.UserID
	prob.Notes = notes.String
	return prob, nil
}

// GetByID fetches a single problem by its ID, scoped to the given user.
// Returns ErrNotFound when no matching row exists.
func (r *sqlProblemRepo) GetByID(ctx context.Context, id, userID int) (models.Problem, error) {
	var prob models.Problem
	var notes sql.NullString
	err := r.db.QueryRowContext(ctx, `
		SELECT id, name, categories, difficulty, attempts, looked_at_solution,
		       time_taken_mins, score, original_score, solved_at, created_at, solution_type, notes
		FROM problems
		WHERE id = $1 AND user_id = $2`,
		id, userID,
	).Scan(
		&prob.ID, &prob.Name, pq.Array(&prob.Categories), &prob.Difficulty,
		&prob.Attempts, &prob.LookedAtSolution, &prob.TimeTakenMins,
		&prob.Score, &prob.OriginalScore, &prob.SolvedAt, &prob.CreatedAt, &prob.SolutionType,
		&notes,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return models.Problem{}, ErrNotFound
	}
	if err != nil {
		return models.Problem{}, fmt.Errorf("GetByID: %w", err)
	}
	prob.UserID = userID
	prob.Notes = notes.String
	return prob, nil
}

func (r *sqlProblemRepo) ListByUser(ctx context.Context, userID int) ([]models.Problem, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, name, categories, difficulty, attempts, looked_at_solution,
		       time_taken_mins, score, original_score, solved_at, created_at, solution_type, notes
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
		var notes sql.NullString
		p.UserID = userID
		if err := rows.Scan(
			&p.ID, &p.Name, pq.Array(&p.Categories), &p.Difficulty,
			&p.Attempts, &p.LookedAtSolution, &p.TimeTakenMins,
			&p.Score, &p.OriginalScore, &p.SolvedAt, &p.CreatedAt, &p.SolutionType,
			&notes,
		); err != nil {
			return nil, err
		}
		p.Notes = notes.String
		problems = append(problems, p)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return problems, nil
}

// ListByUserFiltered returns a paginated, filtered page of problems for a user.
// All filter fields are optional; omitted fields match all rows.
func (r *sqlProblemRepo) ListByUserFiltered(ctx context.Context, userID int, f ListProblemsFilter) (ListProblemsResult, error) {
	// Build WHERE clause dynamically to avoid unnecessary conditions.
	args := []any{userID}
	conditions := []string{"user_id = $1"}
	argIdx := 2

	if f.NameSearch != "" {
		// ESCAPE '\' pairs with the service-layer escapeLikePattern call so that
		// literal % and _ in the search term are not treated as LIKE wildcards.
		conditions = append(conditions, fmt.Sprintf(`name ILIKE $%d ESCAPE '\'`, argIdx))
		args = append(args, "%"+f.NameSearch+"%")
		argIdx++
	}
	if len(f.Difficulties) > 0 {
		// Match any of the requested difficulties.
		conditions = append(conditions, fmt.Sprintf("difficulty = ANY($%d)", argIdx))
		args = append(args, pq.Array(f.Difficulties))
		argIdx++
	}
	if len(f.Categories) > 0 {
		// Overlap operator (&&): problem must share at least one category with the filter set.
		conditions = append(conditions, fmt.Sprintf("categories && $%d", argIdx))
		args = append(args, pq.Array(f.Categories))
		argIdx++
	}
	if f.ScoreMin != nil {
		conditions = append(conditions, fmt.Sprintf("score >= $%d", argIdx))
		args = append(args, *f.ScoreMin)
		argIdx++
	}
	if f.ScoreMax != nil {
		conditions = append(conditions, fmt.Sprintf("score <= $%d", argIdx))
		args = append(args, *f.ScoreMax)
		argIdx++
	}
	if f.DateFrom != nil {
		conditions = append(conditions, fmt.Sprintf("created_at >= $%d", argIdx))
		args = append(args, *f.DateFrom)
		argIdx++
	}
	if f.DateTo != nil {
		// DateTo is exclusive — the caller adds 1 day to make the UI end-date inclusive.
		conditions = append(conditions, fmt.Sprintf("created_at < $%d", argIdx))
		args = append(args, *f.DateTo)
		argIdx++
	}

	where := "WHERE " + strings.Join(conditions, " AND ")

	// Count total matching rows (for pagination metadata).
	var total int
	if err := r.db.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM problems "+where, args...,
	).Scan(&total); err != nil {
		return ListProblemsResult{}, fmt.Errorf("ListByUserFiltered count: %w", err)
	}

	limit := f.Limit
	if limit <= 0 {
		limit = 20
	}
	offset := f.Offset
	if offset < 0 {
		offset = 0
	}

	// Append LIMIT and OFFSET as the last two positional parameters.
	pageArgs := make([]any, len(args)+2)
	copy(pageArgs, args)
	pageArgs[len(args)] = limit
	pageArgs[len(args)+1] = offset

	rows, err := r.db.QueryContext(ctx, `
		SELECT id, name, categories, difficulty, attempts, looked_at_solution,
		       time_taken_mins, score, original_score, solved_at, created_at, solution_type, notes
		FROM problems
		`+where+`
		ORDER BY created_at DESC
		LIMIT $`+fmt.Sprint(argIdx)+` OFFSET $`+fmt.Sprint(argIdx+1),
		pageArgs...,
	)
	if err != nil {
		return ListProblemsResult{}, fmt.Errorf("ListByUserFiltered query: %w", err)
	}
	defer rows.Close()

	var problems []models.Problem
	for rows.Next() {
		var p models.Problem
		var notes sql.NullString
		p.UserID = userID
		if err := rows.Scan(
			&p.ID, &p.Name, pq.Array(&p.Categories), &p.Difficulty,
			&p.Attempts, &p.LookedAtSolution, &p.TimeTakenMins,
			&p.Score, &p.OriginalScore, &p.SolvedAt, &p.CreatedAt, &p.SolutionType,
			&notes,
		); err != nil {
			return ListProblemsResult{}, fmt.Errorf("ListByUserFiltered scan: %w", err)
		}
		p.Notes = notes.String
		problems = append(problems, p)
	}
	if err := rows.Err(); err != nil {
		return ListProblemsResult{}, fmt.Errorf("ListByUserFiltered rows: %w", err)
	}

	if problems == nil {
		problems = []models.Problem{}
	}
	return ListProblemsResult{Problems: problems, Total: total}, nil
}

// DecayAllProblems updates score = ApplyDecay(original_score, solved_at) for every problem row.
// now is the reference timestamp used as "today" so callers control when decay is evaluated.
// Constants match models/score.go: 7-day grace, 1.0 pt/day, 0.40 floor.
func (r *sqlProblemRepo) DecayAllProblems(ctx context.Context, now time.Time) (int64, error) {
	// Only decay the latest submission per (user_id, name) — older submissions
	// are excluded from category stats and the retry queue, so there's no value
	// in updating their score column.
	result, err := r.db.ExecContext(ctx, `
		UPDATE problems p
		SET score = GREATEST(
		    CEIL(original_score * 0.40),
		    ROUND(
		        original_score::numeric
		        - GREATEST(
		            EXTRACT(EPOCH FROM ($1::timestamptz - p.solved_at)) / 86400.0 - 7,
		            0
		          ) * 1.0
		    )
		)::integer
		WHERE p.id IN (
		    SELECT DISTINCT ON (user_id, name) id
		    FROM problems
		    ORDER BY user_id, name, created_at DESC
		)`,
		now,
	)
	if err != nil {
		return 0, fmt.Errorf("DecayAllProblems: %w", err)
	}
	n, _ := result.RowsAffected()
	return n, nil
}
