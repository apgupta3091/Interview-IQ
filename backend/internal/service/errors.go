package service

import "errors"

var (
	ErrNoProblems = errors.New("no problems logged yet")
	ErrNotFound   = errors.New("not found")
	// ErrFreeTierLimitReached = errors.New("free tier problem limit reached")
	// Payments removed — all users have unlimited access. Re-enable when billing is added back.
)

// ValidationError carries a user-facing validation message.
type ValidationError struct {
	Message string
}

func (e ValidationError) Error() string {
	return e.Message
}
