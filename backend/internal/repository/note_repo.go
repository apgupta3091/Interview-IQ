package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/apgupta3091/interview-iq/internal/models"
)

type NoteRepository interface {
	// ListByProblemName returns all notes for (userID, problemName) sorted newest first.
	ListByProblemName(ctx context.Context, userID int, problemName string) ([]models.Note, error)
	// Insert creates a new note and returns it.
	Insert(ctx context.Context, userID int, problemName, content string) (models.Note, error)
	// Update replaces the content of an existing note owned by userID.
	// Returns ErrNotFound when no matching row exists.
	Update(ctx context.Context, noteID, userID int, content string) (models.Note, error)
	// Delete removes a note owned by userID.
	// Returns ErrNotFound when no matching row exists.
	Delete(ctx context.Context, noteID, userID int) error
}

type sqlNoteRepo struct {
	db *sql.DB
}

func NewNoteRepo(db *sql.DB) NoteRepository {
	return &sqlNoteRepo{db: db}
}

func (r *sqlNoteRepo) ListByProblemName(ctx context.Context, userID int, problemName string) ([]models.Note, error) {
	// Normalise the problem name the same way Insert does.
	name := strings.ToLower(strings.TrimSpace(problemName))

	rows, err := r.db.QueryContext(ctx, `
		SELECT id, user_id, problem_name, content, created_at, updated_at
		FROM problem_notes
		WHERE user_id = $1 AND problem_name = $2
		ORDER BY created_at ASC`,
		userID, name,
	)
	if err != nil {
		return nil, fmt.Errorf("ListByProblemName: %w", err)
	}
	defer rows.Close()

	var notes []models.Note
	for rows.Next() {
		var n models.Note
		if err := rows.Scan(&n.ID, &n.UserID, &n.ProblemName, &n.Content, &n.CreatedAt, &n.UpdatedAt); err != nil {
			return nil, fmt.Errorf("ListByProblemName scan: %w", err)
		}
		notes = append(notes, n)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("ListByProblemName rows: %w", err)
	}
	if notes == nil {
		notes = []models.Note{}
	}
	return notes, nil
}

func (r *sqlNoteRepo) Insert(ctx context.Context, userID int, problemName, content string) (models.Note, error) {
	// Normalise problem name so "Two Sum" and "two sum" resolve to the same bucket.
	name := strings.ToLower(strings.TrimSpace(problemName))

	var n models.Note
	err := r.db.QueryRowContext(ctx, `
		INSERT INTO problem_notes (user_id, problem_name, content)
		VALUES ($1, $2, $3)
		RETURNING id, user_id, problem_name, content, created_at, updated_at`,
		userID, name, content,
	).Scan(&n.ID, &n.UserID, &n.ProblemName, &n.Content, &n.CreatedAt, &n.UpdatedAt)
	if err != nil {
		return models.Note{}, fmt.Errorf("Insert note: %w", err)
	}
	return n, nil
}

func (r *sqlNoteRepo) Update(ctx context.Context, noteID, userID int, content string) (models.Note, error) {
	var n models.Note
	err := r.db.QueryRowContext(ctx, `
		UPDATE problem_notes
		SET content = $1, updated_at = NOW()
		WHERE id = $2 AND user_id = $3
		RETURNING id, user_id, problem_name, content, created_at, updated_at`,
		content, noteID, userID,
	).Scan(&n.ID, &n.UserID, &n.ProblemName, &n.Content, &n.CreatedAt, &n.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return models.Note{}, ErrNotFound
	}
	if err != nil {
		return models.Note{}, fmt.Errorf("Update note: %w", err)
	}
	return n, nil
}

func (r *sqlNoteRepo) Delete(ctx context.Context, noteID, userID int) error {
	result, err := r.db.ExecContext(ctx, `
		DELETE FROM problem_notes WHERE id = $1 AND user_id = $2`,
		noteID, userID,
	)
	if err != nil {
		return fmt.Errorf("Delete note: %w", err)
	}
	n, _ := result.RowsAffected()
	if n == 0 {
		return ErrNotFound
	}
	return nil
}
