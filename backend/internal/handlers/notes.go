package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/apgupta3091/interview-iq/internal/middleware"
	"github.com/apgupta3091/interview-iq/internal/models"
	"github.com/apgupta3091/interview-iq/internal/service"
)

// NoteHandler handles CRUD for problem notes.
// Notes are keyed by (userID, problemName) — aggregated across all attempts.
type NoteHandler struct {
	Notes    service.NoteService
	Problems service.ProblemService
}

type noteResponse struct {
	ID          int    `json:"id"`
	ProblemName string `json:"problem_name"`
	Content     string `json:"content"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

type listNotesResponse struct {
	Notes []noteResponse `json:"notes"`
}

func toNoteResponse(n models.Note) noteResponse {
	return noteResponse{
		ID:          n.ID,
		ProblemName: n.ProblemName,
		Content:     n.Content,
		CreatedAt:   n.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:   n.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

// ListNotes returns all notes for the problem identified by :problemID.
// Notes are scoped by the problem's name so all attempts share the same pool.
func (h *NoteHandler) ListNotes(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())

	problemID, err := parseProblemID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid problem ID")
		return
	}

	// Look up the problem to get its canonical name.
	p, err := h.Problems.GetByID(r.Context(), userID, problemID)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			writeError(w, http.StatusNotFound, "problem not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to fetch problem")
		return
	}

	notes, err := h.Notes.List(r.Context(), userID, p.Name)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to fetch notes")
		return
	}

	resp := make([]noteResponse, len(notes))
	for i, n := range notes {
		resp[i] = toNoteResponse(n)
	}
	writeJSON(w, http.StatusOK, listNotesResponse{Notes: resp})
}

// CreateNote adds a new note to the problem identified by :problemID.
func (h *NoteHandler) CreateNote(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())

	problemID, err := parseProblemID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid problem ID")
		return
	}

	var req struct {
		Content string `json:"content"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	p, err := h.Problems.GetByID(r.Context(), userID, problemID)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			writeError(w, http.StatusNotFound, "problem not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to fetch problem")
		return
	}

	note, err := h.Notes.Create(r.Context(), userID, p.Name, req.Content)
	if err != nil {
		var ve service.ValidationError
		if errors.As(err, &ve) {
			writeError(w, http.StatusBadRequest, ve.Message)
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to create note")
		return
	}

	writeJSON(w, http.StatusCreated, toNoteResponse(note))
}

// UpdateNote replaces the content of the note identified by :noteID.
func (h *NoteHandler) UpdateNote(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())

	noteIDStr := chi.URLParam(r, "noteID")
	noteID, err := strconv.Atoi(noteIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid note ID")
		return
	}

	var req struct {
		Content string `json:"content"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	note, err := h.Notes.Update(r.Context(), noteID, userID, req.Content)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			writeError(w, http.StatusNotFound, "note not found")
			return
		}
		var ve service.ValidationError
		if errors.As(err, &ve) {
			writeError(w, http.StatusBadRequest, ve.Message)
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to update note")
		return
	}

	writeJSON(w, http.StatusOK, toNoteResponse(note))
}

// DeleteNote removes the note identified by :noteID.
func (h *NoteHandler) DeleteNote(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())

	noteIDStr := chi.URLParam(r, "noteID")
	noteID, err := strconv.Atoi(noteIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid note ID")
		return
	}

	if err := h.Notes.Delete(r.Context(), noteID, userID); err != nil {
		if errors.Is(err, service.ErrNotFound) {
			writeError(w, http.StatusNotFound, "note not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to delete note")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func parseProblemID(r *http.Request) (int, error) {
	return strconv.Atoi(chi.URLParam(r, "problemID"))
}
