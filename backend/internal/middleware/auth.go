package middleware

import (
	"context"
	"net/http"
	"strings"

	clerkjwt "github.com/clerk/clerk-sdk-go/v2/jwt"

	"github.com/apgupta3091/interview-iq/internal/repository"
)

type contextKey string

const (
	UserIDKey contextKey = "userID"
	TierKey   contextKey = "tier"
)

// ClerkAuthenticate verifies Clerk-issued JWTs, auto-provisions an internal
// integer user ID on first sign-in, and injects both the user ID and
// subscription tier into the request context.
func ClerkAuthenticate(users repository.UserRepository) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
				http.Error(w, `{"error":"missing or invalid authorization header"}`, http.StatusUnauthorized)
				return
			}

			tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
			// Verify validates the Clerk RS256 JWT against Clerk's published JWKS.
			claims, err := clerkjwt.Verify(r.Context(), &clerkjwt.VerifyParams{Token: tokenStr})
			if err != nil {
				http.Error(w, `{"error":"invalid or expired token"}`, http.StatusUnauthorized)
				return
			}

			// claims.Subject is the Clerk user ID (e.g. "user_abc123").
			// GetOrCreateByClerkID maps it to our internal integer user_id and also
			// returns the subscription tier so both are available in a single query.
			userID, tier, err := users.GetOrCreateByClerkID(r.Context(), claims.Subject)
			if err != nil {
				http.Error(w, `{"error":"failed to resolve user"}`, http.StatusInternalServerError)
				return
			}

			ctx := context.WithValue(r.Context(), UserIDKey, userID)
			ctx = context.WithValue(ctx, TierKey, tier)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// UserIDFromContext extracts the authenticated user's ID from the request context.
func UserIDFromContext(ctx context.Context) int {
	id, _ := ctx.Value(UserIDKey).(int)
	return id
}

// TierFromContext extracts the user's subscription tier from the request context.
// Returns "free" when no tier is present (safe default).
func TierFromContext(ctx context.Context) string {
	tier, _ := ctx.Value(TierKey).(string)
	if tier == "" {
		return "free"
	}
	return tier
}
