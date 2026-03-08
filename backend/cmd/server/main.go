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
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/joho/godotenv"
	httpSwagger "github.com/swaggo/http-swagger"

	clerk "github.com/clerk/clerk-sdk-go/v2"
	"golang.org/x/time/rate"

	_ "github.com/apgupta3091/interview-iq/docs"
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

	problemHandler := &handlers.ProblemHandler{Service: problemSvc}
	categoryHandler := &handlers.CategoryHandler{Service: categorySvc}
	recHandler := &handlers.RecommendationHandler{Service: recSvc}
	lcHandler := &handlers.LeetCodeHandler{Repo: lcRepo}

	r := chi.NewRouter()

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:5173"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Authorization", "Content-Type"},
		AllowCredentials: false,
		MaxAge:           300,
	}))
	r.Use(chimiddleware.Logger)
	r.Use(chimiddleware.Recoverer)
	r.Use(chimiddleware.RequestID)
	// Global per-IP rate limit: 30 req/min sustained, burst of 10.
	// Applied before auth so unauthenticated abuse is stopped early.
	r.Use(middleware.RateLimitByIP(rate.Every(2*time.Second), 10))

	r.Get("/health", healthHandler)
	r.Get("/swagger/*", httpSwagger.WrapHandler)

	r.Route("/api", func(r chi.Router) {
		// All API routes require a valid Clerk JWT.
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
			r.Get("/categories/stats", categoryHandler.GetStats)
			r.Get("/categories/weakest", categoryHandler.GetWeakest)
			r.Get("/leetcode-problems/search", lcHandler.Search)
			r.Get("/recommendations", recHandler.GetRecommendations)
		})
	})

	log.Printf("server starting on :%s", port)
	if err := http.ListenAndServe(":"+port, r); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status": "ok",
	})
}
