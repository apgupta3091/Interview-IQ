package service

import "errors"

var (
	ErrNoProblems = errors.New("no problems logged yet")
	ErrNotFound   = errors.New("not found")
)

// ValidationError carries a user-facing validation message.
type ValidationError struct {
	Message string
}

func (e ValidationError) Error() string {
	return e.Message
}
