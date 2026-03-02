package service

import "errors"

var (
	ErrEmailTaken         = errors.New("email already taken")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrNoProblems         = errors.New("no problems logged yet")
)

// ValidationError carries a user-facing validation message.
type ValidationError struct {
	Message string
}

func (e ValidationError) Error() string {
	return e.Message
}
