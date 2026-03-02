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

	r := chi.NewRouter()

	r.Use(chimiddleware.Logger)
	r.Use(chimiddleware.Recoverer)
	r.Use(chimiddleware.RequestID)

	authHandler    := &handlers.AuthHandler{DB: database}
	problemHandler := &handlers.ProblemHandler{DB: database}

	r.Get("/health", healthHandler)

	r.Route("/api", func(r chi.Router) {
		// public routes — no auth required
		r.Post("/auth/register", authHandler.Register)
		r.Post("/auth/login", authHandler.Login)

		// protected routes — JWT required
		r.Group(func(r chi.Router) {
			r.Use(middleware.Authenticate)
			r.Post("/problems", problemHandler.LogProblem)
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
