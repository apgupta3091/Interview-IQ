package service

import (
	"context"
	"errors"
	"strings"

	"golang.org/x/crypto/bcrypt"

	"github.com/apgupta3091/interview-iq/internal/auth"
	"github.com/apgupta3091/interview-iq/internal/repository"
)

type AuthService interface {
	Register(ctx context.Context, email, password string) (string, string, error)
	Login(ctx context.Context, email, password string) (string, string, error)
}

type authService struct {
	users repository.UserRepository
}

func NewAuthService(users repository.UserRepository) AuthService {
	return &authService{users: users}
}

func (s *authService) Register(ctx context.Context, email, password string) (string, string, error) {
	email = strings.TrimSpace(strings.ToLower(email))
	if email == "" || len(password) < 8 {
		return "", "", ValidationError{Message: "email required and password must be at least 8 characters"}
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", "", err
	}

	userID, err := s.users.Create(ctx, email, string(hash))
	if err != nil {
		if errors.Is(err, repository.ErrDuplicate) {
			return "", "", ErrEmailTaken
		}
		return "", "", err
	}

	token, err := auth.GenerateToken(userID)
	if err != nil {
		return "", "", err
	}
	return token, email, nil
}

func (s *authService) Login(ctx context.Context, email, password string) (string, string, error) {
	email = strings.TrimSpace(strings.ToLower(email))

	userID, hash, err := s.users.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return "", "", ErrInvalidCredentials
		}
		return "", "", err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)); err != nil {
		return "", "", ErrInvalidCredentials
	}

	token, err := auth.GenerateToken(userID)
	if err != nil {
		return "", "", err
	}
	return token, email, nil
}
