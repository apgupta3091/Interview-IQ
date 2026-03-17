// @title           Interview Skill Radar API
// @version         1.0
// @description     Backend API for tracking LeetCode problem attempts, computing skill scores with time decay, and surfacing category weaknesses.
// @host            localhost:8080
// @BasePath        /api
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Enter your Clerk session token as: Bearer <token>
package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"
	_ "time/tzdata" // embed IANA tz data so America/New_York works inside Docker

	"github.com/getsentry/sentry-go"
	sentryhttp "github.com/getsentry/sentry-go/http"
	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/joho/godotenv"
	httpSwagger "github.com/swaggo/http-swagger"

	clerk "github.com/clerk/clerk-sdk-go/v2"
	"golang.org/x/time/rate"

	_ "github.com/apgupta3091/interview-iq/docs"
	"github.com/apgupta3091/interview-iq/internal/cron"
	"github.com/apgupta3091/interview-iq/internal/db"
	"github.com/apgupta3091/interview-iq/internal/handlers"
	"github.com/apgupta3091/interview-iq/internal/middleware"
	"github.com/apgupta3091/interview-iq/internal/repository"
	"github.com/apgupta3091/interview-iq/internal/service"
)

func main() {
	// Load .env if present (dev convenience; ignored if the file doesn't exist).
	_ = godotenv.Load()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Initialise Sentry — captures panics and errors in production.
	// APP_ENV should be "production" or "development".
	if err := sentry.Init(sentry.ClientOptions{
		Dsn:              os.Getenv("SENTRY_DSN"),
		Environment:      os.Getenv("APP_ENV"),
		TracesSampleRate: 0.1,
	}); err != nil {
		log.Printf("sentry init: %v", err)
	}
	defer sentry.Flush(2 * time.Second)

	// Initialise Clerk SDK — must be called before any token verification.
	clerk.SetKey(os.Getenv("CLERK_SECRET_KEY"))

	database, err := db.Connect()
	if err != nil {
		log.Fatalf("db connect: %v", err)
	}
	defer database.Close()

	if err := db.RunMigrations(database, "migrations"); err != nil {
		log.Fatalf("migrations: %v", err)
	}

	userRepo := repository.NewUserRepo(database)
	problemRepo := repository.NewProblemRepo(database)
	categoryRepo := repository.NewCategoryRepo(database)
	lcRepo := repository.NewLeetCodeProblemRepo(database)

	openaiKey := os.Getenv("OPENAI_API_KEY")

	problemSvc := service.NewProblemService(problemRepo)
	categorySvc := service.NewCategoryService(categoryRepo)
	recSvc := service.NewRecommendationService(categorySvc, problemSvc, openaiKey)
	billingSvc := service.NewBillingService(userRepo, os.Getenv("STRIPE_SECRET_KEY"))

	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL == "" {
		frontendURL = "http://localhost:5173"
	}

	problemHandler := &handlers.ProblemHandler{Service: problemSvc, RecCache: recSvc}
	categoryHandler := &handlers.CategoryHandler{Service: categorySvc}
	recHandler := &handlers.RecommendationHandler{Service: recSvc}
	lcHandler := &handlers.LeetCodeHandler{Repo: lcRepo}
	billingHandler := &handlers.BillingHandler{
		Service:       billingSvc,
		Problems:      problemSvc,
		WebhookSecret: os.Getenv("STRIPE_WEBHOOK_SECRET"),
		PriceMonthly:  os.Getenv("STRIPE_PRICE_MONTHLY"),
		PriceAnnual:   os.Getenv("STRIPE_PRICE_ANNUAL"),
		FrontendURL:   frontendURL,
	}

	r := chi.NewRouter()

	// Build allowed origins from env + always include local dev.
	allowedOrigins := []string{"http://localhost:5173"}
	if frontendURL != "" && frontendURL != "http://localhost:5173" {
		allowedOrigins = append(allowedOrigins, frontendURL)
	}

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   allowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Authorization", "Content-Type"},
		AllowCredentials: false,
		MaxAge:           300,
	}))
	r.Use(chimiddleware.Logger)
	r.Use(chimiddleware.Recoverer)
	r.Use(chimiddleware.RequestID)
	// Sentry middleware: captures panics and reports them before Recoverer swallows them.
	sentryHandler := sentryhttp.New(sentryhttp.Options{Repanic: true})
	r.Use(sentryHandler.Handle)
	// Global per-IP rate limit: 60 req/min sustained, burst of 20.
	// Applied before auth so unauthenticated abuse is stopped early.
	r.Use(middleware.RateLimitByIP(rate.Every(time.Second), 20))

	r.Get("/health", healthHandler(database))
	r.Get("/swagger/*", httpSwagger.WrapHandler)

	r.Route("/api", func(r chi.Router) {
		// Stripe webhook: unauthenticated — Stripe verifies via signature header.
		r.Post("/webhooks/stripe", billingHandler.HandleWebhook)

		// All other API routes require a valid Clerk JWT.
		r.Group(func(r chi.Router) {
			r.Use(middleware.ClerkAuthenticate(userRepo))
			// Per-user rate limit: 120 req/min sustained, burst of 40.
			// Applied after auth so we limit by resolved internal user ID.
			// Higher burst is needed because the Dashboard fires ~4 parallel
			// requests on load and the typeahead search fires on each debounced
			// keystroke shortly after.
			r.Use(middleware.RateLimitByUser(rate.Every(500*time.Millisecond), 40))
			r.Get("/problems", problemHandler.ListProblems)
			r.Post("/problems", problemHandler.LogProblem)
			r.Get("/problems/{problemID}", problemHandler.GetProblem)
			r.Get("/categories/stats", categoryHandler.GetStats)
			r.Get("/categories/weakest", categoryHandler.GetWeakest)
			r.Get("/leetcode-problems/search", lcHandler.Search)
			r.Get("/recommendations", recHandler.GetRecommendations)
			r.Get("/billing/status", billingHandler.GetStatus)
			r.Post("/billing/checkout", billingHandler.CreateCheckoutSession)
			r.Post("/billing/portal", billingHandler.CreatePortalSession)
		})
	})

	// ctx is cancelled on SIGINT/SIGTERM, which stops the decay cron and the HTTP server.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	// Start the nightly score-decay cron (10pm EST daily).
	cron.RunDecayCron(ctx, problemRepo)

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}

	// Shut down gracefully when the signal context is cancelled.
	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := srv.Shutdown(shutdownCtx); err != nil {
			log.Printf("server shutdown: %v", err)
		}
	}()

	log.Printf("server starting on :%s", port)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server failed: %v", err)
	}
}

// healthHandler returns a closure that checks DB connectivity before responding.
func healthHandler(database *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := database.PingContext(r.Context()); err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusServiceUnavailable)
			json.NewEncoder(w).Encode(map[string]string{"error": "db unavailable"})
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}
}
