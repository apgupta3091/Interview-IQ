package service

import "errors"

var (
	ErrNoProblems = errors.New("no problems logged yet")
)

// ValidationError carries a user-facing validation message.
type ValidationError struct {
	Message string
}

func (e ValidationError) Error() string {
	return e.Message
}
