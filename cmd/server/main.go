package main

import (
	"log"
	"net/http"

	"github.com/Ashoke15/AuthX/internal/mailer"
	authmw "github.com/Ashoke15/AuthX/internal/middleware"

	"github.com/Ashoke15/AuthX/internal/config"
	"github.com/Ashoke15/AuthX/internal/db"
	"github.com/Ashoke15/AuthX/internal/handlers"
	"github.com/Ashoke15/AuthX/internal/repository"
	"github.com/go-chi/chi"
	chimw "github.com/go-chi/chi/middleware"
)

func main() {
	cfg := config.Load()

	conn, err := db.New(cfg.DatabaseUrl)
	if err != nil {
		log.Fatalf("failed to connect db: %v", err)
	}
	defer conn.Close()

	userRepo := repository.NPURepository(conn)
	refreshRepo := repository.NPRTRepositry(conn)
	verifyRepo := repository.NewPGEVRepo(conn)
	emailMailer, err := mailer.NewSmtpMailer(cfg.SMTPHost, cfg.SMTPPort, cfg.SMTPFrom, cfg.AppName)
	if err != nil {
		log.Fatalf("failed to load email templates: %v", err)
	}

	registerHandeler := handlers.NewRegisterHandeler(userRepo, verifyRepo, emailMailer)
	loginHandler := handlers.NewLoginHandler(userRepo, refreshRepo, cfg.JWTSecret)
	refreshHandler := handlers.NewRefreshHandeler(refreshRepo, cfg.JWTSecret)
	verifyEmailHandler := handlers.NewVerifyEmailHandler(userRepo, verifyRepo)
	meHandler := handlers.NewMeHandler(userRepo)

	r := chi.NewRouter()
	r.Use(chimw.Logger)
	r.Use(chimw.Recoverer)

	r.Post("/register", registerHandeler.ServeHTTP)
	r.Post("/login", loginHandler.ServeHTTP)
	r.Post("/refresh", refreshHandler.ServeHTTP)
	r.Post("/verify-email", verifyEmailHandler.ServeHTTP)

	r.Group(func(protect chi.Router) {
		protect.Use(authmw.RequireAuth(cfg.JWTSecret))
		protect.Get("/me", meHandler.ServeHTTP)
	})

	log.Printf("auth-service listening on :%s", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, r); err != nil {
		log.Fatal(err)
	}
}
