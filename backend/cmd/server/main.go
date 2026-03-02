package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"

	"github.com/apgupta3091/interview-iq/internal/db"
	"github.com/apgupta3091/interview-iq/internal/handlers"
	"github.com/apgupta3091/interview-iq/internal/middleware"
	"github.com/apgupta3091/interview-iq/internal/repository"
	"github.com/apgupta3091/interview-iq/internal/service"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

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

	authSvc := service.NewAuthService(userRepo)
	problemSvc := service.NewProblemService(problemRepo)
	categorySvc := service.NewCategoryService(categoryRepo)

	authHandler := &handlers.AuthHandler{Service: authSvc}
	problemHandler := &handlers.ProblemHandler{Service: problemSvc}
	categoryHandler := &handlers.CategoryHandler{Service: categorySvc}

	r := chi.NewRouter()

	r.Use(chimiddleware.Logger)
	r.Use(chimiddleware.Recoverer)
	r.Use(chimiddleware.RequestID)

	r.Get("/health", healthHandler)

	r.Route("/api", func(r chi.Router) {
		// public routes — no auth required
		r.Post("/auth/register", authHandler.Register)
		r.Post("/auth/login", authHandler.Login)

		// protected routes — JWT required
		r.Group(func(r chi.Router) {
			r.Use(middleware.Authenticate)
			r.Get("/problems", problemHandler.ListProblems)
			r.Post("/problems", problemHandler.LogProblem)
			r.Get("/categories/stats", categoryHandler.GetStats)
			r.Get("/categories/weakest", categoryHandler.GetWeakest)
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
