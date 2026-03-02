package service

import (
	"context"
	"errors"
	"testing"

	"golang.org/x/crypto/bcrypt"

	"github.com/apgupta3091/interview-iq/internal/repository"
)

type mockUserRepo struct {
	createFn     func(ctx context.Context, email, passwordHash string) (int, error)
	getByEmailFn func(ctx context.Context, email string) (int, string, error)
}

func (m *mockUserRepo) Create(ctx context.Context, email, passwordHash string) (int, error) {
	return m.createFn(ctx, email, passwordHash)
}

func (m *mockUserRepo) GetByEmail(ctx context.Context, email string) (int, string, error) {
	return m.getByEmailFn(ctx, email)
}

func TestRegister_ShortPassword(t *testing.T) {
	svc := NewAuthService(&mockUserRepo{})
	_, _, err := svc.Register(context.Background(), "a@b.com", "short")
	var ve ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("expected ValidationError, got %v", err)
	}
}

func TestRegister_Success(t *testing.T) {
	svc := NewAuthService(&mockUserRepo{
		createFn: func(_ context.Context, _, _ string) (int, error) { return 1, nil },
	})
	token, email, err := svc.Register(context.Background(), "Test@Example.COM", "password123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if email != "test@example.com" {
		t.Errorf("expected normalized email test@example.com, got %s", email)
	}
	if token == "" {
		t.Error("expected non-empty token")
	}
}

func TestRegister_DuplicateEmail(t *testing.T) {
	svc := NewAuthService(&mockUserRepo{
		createFn: func(_ context.Context, _, _ string) (int, error) {
			return 0, repository.ErrDuplicate
		},
	})
	_, _, err := svc.Register(context.Background(), "a@b.com", "password123")
	if !errors.Is(err, ErrEmailTaken) {
		t.Fatalf("expected ErrEmailTaken, got %v", err)
	}
}

func TestLogin_NotFound(t *testing.T) {
	svc := NewAuthService(&mockUserRepo{
		getByEmailFn: func(_ context.Context, _ string) (int, string, error) {
			return 0, "", repository.ErrNotFound
		},
	})
	_, _, err := svc.Login(context.Background(), "a@b.com", "password123")
	if !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("expected ErrInvalidCredentials, got %v", err)
	}
}

func TestLogin_WrongPassword(t *testing.T) {
	hash, _ := bcrypt.GenerateFromPassword([]byte("correctPassword"), bcrypt.DefaultCost)
	svc := NewAuthService(&mockUserRepo{
		getByEmailFn: func(_ context.Context, _ string) (int, string, error) {
			return 1, string(hash), nil
		},
	})
	_, _, err := svc.Login(context.Background(), "a@b.com", "wrongPassword")
	if !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("expected ErrInvalidCredentials, got %v", err)
	}
}

func TestLogin_Success(t *testing.T) {
	hash, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	svc := NewAuthService(&mockUserRepo{
		getByEmailFn: func(_ context.Context, _ string) (int, string, error) {
			return 1, string(hash), nil
		},
	})
	token, email, err := svc.Login(context.Background(), "A@B.COM", "password123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if email != "a@b.com" {
		t.Errorf("expected normalized email a@b.com, got %s", email)
	}
	if token == "" {
		t.Error("expected non-empty token")
	}
}
