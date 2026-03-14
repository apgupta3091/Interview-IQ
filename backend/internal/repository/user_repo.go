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
	// GetOrCreateByClerkID returns the internal integer user ID and subscription tier
	// for a Clerk user, creating a new row on first sign-in (upsert).
	GetOrCreateByClerkID(ctx context.Context, clerkUserID string) (userID int, tier string, err error)
	// GetBilling returns the Stripe customer ID and subscription tier for the given user.
	// stripeCustomerID is empty string when no Stripe customer has been created yet.
	GetBilling(ctx context.Context, userID int) (stripeCustomerID string, tier string, err error)
	// UpdateBilling persists a new Stripe customer ID and subscription tier.
	UpdateBilling(ctx context.Context, userID int, stripeCustomerID, tier string) error
	// GetUserIDByStripeCustomerID returns the internal user ID for a Stripe customer ID.
	// Returns ErrNotFound when no matching row exists.
	GetUserIDByStripeCustomerID(ctx context.Context, stripeCustomerID string) (int, error)
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

// GetOrCreateByClerkID upserts a user row keyed by clerk_user_id and returns the
// internal integer id and subscription tier in a single round-trip.
func (r *sqlUserRepo) GetOrCreateByClerkID(ctx context.Context, clerkUserID string) (int, string, error) {
	var id int
	var tier string
	err := r.db.QueryRowContext(ctx,
		`INSERT INTO users (clerk_user_id)
		 VALUES ($1)
		 ON CONFLICT (clerk_user_id) DO UPDATE SET clerk_user_id = EXCLUDED.clerk_user_id
		 RETURNING id, subscription_tier`,
		clerkUserID,
	).Scan(&id, &tier)
	if err != nil {
		return 0, "free", fmt.Errorf("user_repo.GetOrCreateByClerkID: %w", err)
	}
	return id, tier, nil
}

// GetBilling returns the Stripe customer ID and subscription tier for the given user.
func (r *sqlUserRepo) GetBilling(ctx context.Context, userID int) (string, string, error) {
	var customerID sql.NullString
	var tier string
	err := r.db.QueryRowContext(ctx,
		`SELECT stripe_customer_id, subscription_tier FROM users WHERE id = $1`,
		userID,
	).Scan(&customerID, &tier)
	if errors.Is(err, sql.ErrNoRows) {
		return "", "free", ErrNotFound
	}
	if err != nil {
		return "", "free", fmt.Errorf("GetBilling: %w", err)
	}
	return customerID.String, tier, nil
}

// UpdateBilling persists a new Stripe customer ID and subscription tier for the user.
func (r *sqlUserRepo) UpdateBilling(ctx context.Context, userID int, stripeCustomerID, tier string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE users SET stripe_customer_id = $1, subscription_tier = $2 WHERE id = $3`,
		stripeCustomerID, tier, userID,
	)
	if err != nil {
		return fmt.Errorf("UpdateBilling: %w", err)
	}
	return nil
}

// GetUserIDByStripeCustomerID looks up the internal user ID by Stripe customer ID.
func (r *sqlUserRepo) GetUserIDByStripeCustomerID(ctx context.Context, stripeCustomerID string) (int, error) {
	var id int
	err := r.db.QueryRowContext(ctx,
		`SELECT id FROM users WHERE stripe_customer_id = $1`,
		stripeCustomerID,
	).Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, ErrNotFound
	}
	if err != nil {
		return 0, fmt.Errorf("GetUserIDByStripeCustomerID: %w", err)
	}
	return id, nil
}
