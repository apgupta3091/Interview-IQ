package service

import "errors"

var (
	ErrNoProblems           = errors.New("no problems logged yet")
	ErrNotFound             = errors.New("not found")
	ErrFreeTierLimitReached = errors.New("free tier problem limit reached")
)

// ValidationError carries a user-facing validation message.
type ValidationError struct {
	Message string
}

func (e ValidationError) Error() string {
	return e.Message
}
