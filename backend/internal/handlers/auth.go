package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/apgupta3091/interview-iq/internal/service"
)

type AuthHandler struct {
	Service service.AuthService
}

type authRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type authResponse struct {
	Token string `json:"token"`
	Email string `json:"email"`
}

// Register godoc
// @Summary      Register a new user
// @Description  Creates a new user account with email and password. Returns a JWT token on success.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body  body      authRequest   true  "Registration credentials (password min 8 chars)"
// @Success      201   {object}  authResponse  "JWT token and email"
// @Failure      400   {object}  errorResponse "Invalid input or password too short"
// @Failure      409   {object}  errorResponse "Email already registered"
// @Failure      500   {object}  errorResponse "Internal server error"
// @Router       /auth/register [post]
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req authRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	token, email, err := h.Service.Register(r.Context(), req.Email, req.Password)
	if err != nil {
		var ve service.ValidationError
		if errors.As(err, &ve) {
			writeError(w, http.StatusBadRequest, ve.Message)
			return
		}
		if errors.Is(err, service.ErrEmailTaken) {
			writeError(w, http.StatusConflict, "email already registered")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to register")
		return
	}

	writeJSON(w, http.StatusCreated, authResponse{Token: token, Email: email})
}

// Login godoc
// @Summary      Login with existing credentials
// @Description  Authenticates a user and returns a JWT token valid for 7 days.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body  body      authRequest   true  "Login credentials"
// @Success      200   {object}  authResponse  "JWT token and email"
// @Failure      400   {object}  errorResponse "Invalid request body"
// @Failure      401   {object}  errorResponse "Invalid email or password"
// @Failure      500   {object}  errorResponse "Internal server error"
// @Router       /auth/login [post]
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req authRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	token, email, err := h.Service.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			writeError(w, http.StatusUnauthorized, "invalid email or password")
			return
		}
		writeError(w, http.StatusInternalServerError, "login failed")
		return
	}

	writeJSON(w, http.StatusOK, authResponse{Token: token, Email: email})
}
