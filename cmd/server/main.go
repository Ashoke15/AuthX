package main

import (
	"log"
	"net/http"

	"github.com/Ashoke15/AuthX/internal/config"
	"github.com/Ashoke15/AuthX/internal/db"
	"github.com/Ashoke15/AuthX/internal/handlers"
	"github.com/Ashoke15/AuthX/internal/repository"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

func main() {
	cfg := config.Load()

	conn, err := db.New(cfg.DatabaseUrl)
	if err != nil {
		log.Fatalf("failed to connect db: %v", err)
	}
	defer conn.Close()

	userRepo := repository.NPURepository(conn)
	registerHandeler := handlers.NewRegisterHandeler(userRepo)
	loginHandler := handlers.NewLoginHandler(userRepo, cfg.JWTSecret)

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Post("/register", registerHandeler.ServeHTTP)
	r.Post("/login", loginHandler.ServeHTTP)

	log.Printf("auth-service listening on :%s", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, r); err != nil {
		log.Fatal(err)
	}
}
