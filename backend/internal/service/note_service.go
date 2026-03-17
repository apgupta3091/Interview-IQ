package service

import (
	"context"
	"errors"
	"strings"

	"github.com/apgupta3091/interview-iq/internal/models"
	"github.com/apgupta3091/interview-iq/internal/repository"
)

// maxNoteLen is the maximum allowed length for a note in characters.
const maxNoteLen = 5000

type NoteService interface {
	// List returns all notes for the given user and problem name.
	List(ctx context.Context, userID int, problemName string) ([]models.Note, error)
	// Create adds a new note. Returns ValidationError for empty or over-length content.
	Create(ctx context.Context, userID int, problemName, content string) (models.Note, error)
	// Update replaces a note's content. Returns ErrNotFound or ValidationError.
	Update(ctx context.Context, noteID, userID int, content string) (models.Note, error)
	// Delete removes a note. Returns ErrNotFound if it doesn't exist or belongs to another user.
	Delete(ctx context.Context, noteID, userID int) error
}

type noteService struct {
	notes repository.NoteRepository
}

func NewNoteService(notes repository.NoteRepository) NoteService {
	return &noteService{notes: notes}
}

func (s *noteService) List(ctx context.Context, userID int, problemName string) ([]models.Note, error) {
	return s.notes.ListByProblemName(ctx, userID, problemName)
}

func (s *noteService) Create(ctx context.Context, userID int, problemName, content string) (models.Note, error) {
	content = strings.TrimSpace(content)
	if content == "" {
		return models.Note{}, ValidationError{Message: "note content cannot be empty"}
	}
	if len(content) > maxNoteLen {
		return models.Note{}, ValidationError{Message: "note must be 5000 characters or fewer"}
	}
	return s.notes.Insert(ctx, userID, problemName, content)
}

func (s *noteService) Update(ctx context.Context, noteID, userID int, content string) (models.Note, error) {
	content = strings.TrimSpace(content)
	if content == "" {
		return models.Note{}, ValidationError{Message: "note content cannot be empty"}
	}
	if len(content) > maxNoteLen {
		return models.Note{}, ValidationError{Message: "note must be 5000 characters or fewer"}
	}
	n, err := s.notes.Update(ctx, noteID, userID, content)
	if errors.Is(err, repository.ErrNotFound) {
		return models.Note{}, ErrNotFound
	}
	return n, err
}

func (s *noteService) Delete(ctx context.Context, noteID, userID int) error {
	err := s.notes.Delete(ctx, noteID, userID)
	if errors.Is(err, repository.ErrNotFound) {
		return ErrNotFound
	}
	return err
}
