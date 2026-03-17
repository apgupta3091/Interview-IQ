package handlers

// Billing/Stripe handler logic is preserved here but disabled.
// Payments are removed for now — all users get full access.
// Re-enable by uncommenting this file and wiring it back into main.go.

/*
import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/apgupta3091/interview-iq/internal/middleware"
	"github.com/apgupta3091/interview-iq/internal/service"
)

// BillingHandler handles Stripe checkout, portal, status, and webhook endpoints.
type BillingHandler struct {
	Service       service.BillingService
	Problems      service.ProblemService
	WebhookSecret string
	PriceMonthly  string // Stripe Price ID for the $7/mo plan
	PriceAnnual   string // Stripe Price ID for the $60/yr plan
	FrontendURL   string // Base URL of the frontend app (for redirect URLs)
}

// CreateCheckoutSession godoc
// POST /api/billing/checkout
// Creates a Stripe Checkout Session and returns its hosted URL.
// Body: {"plan": "monthly"|"annual"}
func (h *BillingHandler) CreateCheckoutSession(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())

	var req struct {
		Plan string `json:"plan"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	priceID := h.PriceMonthly
	if req.Plan == "annual" {
		priceID = h.PriceAnnual
	}
	if priceID == "" {
		writeError(w, http.StatusInternalServerError, "billing not configured")
		return
	}

	successURL := h.FrontendURL + "/dashboard?upgraded=true"
	cancelURL := h.FrontendURL + "/pricing"

	url, err := h.Service.CreateCheckoutSession(r.Context(), userID, priceID, successURL, cancelURL)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create checkout session")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"url": url})
}

// CreatePortalSession godoc
// POST /api/billing/portal
// Creates a Stripe Billing Portal session and returns its URL.
// The user can manage or cancel their subscription from the portal.
func (h *BillingHandler) CreatePortalSession(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())
	returnURL := h.FrontendURL + "/dashboard"

	url, err := h.Service.CreatePortalSession(r.Context(), userID, returnURL)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create portal session")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"url": url})
}

// GetStatus godoc
// GET /api/billing/status
// Returns the user's subscription tier, problem count, and limit (0 = unlimited).
func (h *BillingHandler) GetStatus(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())
	tier := middleware.TierFromContext(r.Context())

	count, err := h.Problems.Count(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to fetch problem count")
		return
	}

	limit := 0 // 0 = unlimited (Pro)
	if tier == "free" {
		limit = 20
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"tier":          tier,
		"problem_count": count,
		"problem_limit": limit,
	})
}

// HandleWebhook godoc
// POST /api/webhooks/stripe
// Receives and processes Stripe webhook events.
// Must be registered WITHOUT the auth middleware.
// Reads the raw body before any parsing so the Stripe signature remains valid.
func (h *BillingHandler) HandleWebhook(w http.ResponseWriter, r *http.Request) {
	payload, err := io.ReadAll(r.Body)
	if err != nil {
		writeError(w, http.StatusBadRequest, "failed to read request body")
		return
	}

	sigHeader := r.Header.Get("Stripe-Signature")
	if err := h.Service.HandleWebhook(r.Context(), payload, sigHeader, h.WebhookSecret); err != nil {
		writeError(w, http.StatusBadRequest, "webhook error")
		return
	}
	w.WriteHeader(http.StatusOK)
}
*/
