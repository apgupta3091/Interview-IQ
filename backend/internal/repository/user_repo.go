package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
)

type UserRepository interface {
	Create(ctx context.Context, email, passwordHash string) (int, error)
	GetByEmail(ctx context.Context, email string) (userID int, passwordHash string, err error)
	// GetOrCreateByClerkID returns the internal integer user ID for a Clerk user,
	// creating a new row on first sign-in (upsert).
	GetOrCreateByClerkID(ctx context.Context, clerkUserID string) (int, error)
}

type sqlUserRepo struct {
	db *sql.DB
}

func NewUserRepo(db *sql.DB) UserRepository {
	return &sqlUserRepo{db: db}
}

func (r *sqlUserRepo) Create(ctx context.Context, email, passwordHash string) (int, error) {
	var id int
	err := r.db.QueryRowContext(ctx,
		`INSERT INTO users (email, password_hash) VALUES ($1, $2) RETURNING id`,
		email, passwordHash,
	).Scan(&id)
	if err != nil {
		if strings.Contains(err.Error(), "unique") {
			return 0, ErrDuplicate
		}
		return 0, err
	}
	return id, nil
}

func (r *sqlUserRepo) GetByEmail(ctx context.Context, email string) (int, string, error) {
	var id int
	var hash string
	err := r.db.QueryRowContext(ctx,
		`SELECT id, password_hash FROM users WHERE email = $1`,
		email,
	).Scan(&id, &hash)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, "", ErrNotFound
	}
	if err != nil {
		return 0, "", err
	}
	return id, hash, nil
}

// GetOrCreateByClerkID upserts a user row keyed by clerk_user_id and returns
// the internal integer id in a single round-trip.
func (r *sqlUserRepo) GetOrCreateByClerkID(ctx context.Context, clerkUserID string) (int, error) {
	var id int
	err := r.db.QueryRowContext(ctx,
		`INSERT INTO users (clerk_user_id)
		 VALUES ($1)
		 ON CONFLICT (clerk_user_id) DO UPDATE SET clerk_user_id = EXCLUDED.clerk_user_id
		 RETURNING id`,
		clerkUserID,
	).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("user_repo.GetOrCreateByClerkID: %w", err)
	}
	return id, nil
}
