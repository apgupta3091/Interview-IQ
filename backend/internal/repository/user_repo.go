package repository

import (
	"context"
	"database/sql"
	"errors"
	"strings"
)

type UserRepository interface {
	Create(ctx context.Context, email, passwordHash string) (int, error)
	GetByEmail(ctx context.Context, email string) (userID int, passwordHash string, err error)
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
