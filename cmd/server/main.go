package main

import (
	"log"
	"net/http"
	"time"

	"github.com/Ashoke15/AuthX/internal/mailer"
	authmw "github.com/Ashoke15/AuthX/internal/middleware"
	"github.com/Ashoke15/AuthX/internal/ratelimit"
	"github.com/Ashoke15/AuthX/internal/sms"
	"golang.org/x/time/rate"

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
	resetRepo := repository.NPgPRRepo(conn)
	phoneRepo := repository.NewPGPVRepo(conn)

	emailMailer, err := mailer.NewSmtpMailer(cfg.SMTPHost, cfg.SMTPPort, cfg.SMTPFrom, cfg.AppName)
	if err != nil {
		log.Fatalf("failed to load email templates: %v", err)
	}

	var smsSender sms.Sender
	switch cfg.SMSProvider {
	case "msg91":
		smsSender = sms.NewMSG91Sender(cfg.MSG91AuthKey, cfg.MSG91FlowId, cfg.MSG91SenderId)
	case "twilio":
		smsSender = sms.NewTwilioSender(cfg.TwilioAccountSID, cfg.TwilioAuthToken, cfg.TwilioFromNumber)
	default:
		smsSender = sms.NewConsolSender()
	}

	iplimiter := ratelimit.New(rate.Every(time.Second), 10, 10*time.Minute)
	loginEmailLimiter := ratelimit.New(rate.Every(time.Minute), 5, 30*time.Minute)
	resetEmailLimiter := ratelimit.New(rate.Every(2*time.Minute), 3, 30*time.Minute)

	registerHandeler := handlers.NewRegisterHandeler(userRepo, verifyRepo, emailMailer)
	loginHandler := handlers.NewLoginHandler(userRepo, refreshRepo, cfg.JWTSecret, loginEmailLimiter)
	refreshHandler := handlers.NewRefreshHandeler(refreshRepo, cfg.JWTSecret)
	verifyEmailHandler := handlers.NewVerifyEmailHandler(userRepo, verifyRepo)
	forgetpasswordHandler := handlers.NewForgetPasswordHandeler(userRepo, resetRepo, emailMailer, resetEmailLimiter)
	resetPassworedHandeler := handlers.NewResetPasswordHander(userRepo, resetRepo, refreshRepo)
	meHandler := handlers.NewMeHandler(userRepo)
	logoutHandler := handlers.NewLogoutHandler(refreshRepo)
	sendPhoneOTPHandler := handlers.NewSentPhoneOTPHandler(userRepo, phoneRepo, smsSender)
	verifyPhoneOTPHandler := handlers.NewVerifyPhoneOTPHandler(userRepo, phoneRepo)

	r := chi.NewRouter()
	r.Use(chimw.Logger)
	r.Use(chimw.Recoverer)

	r.Group(func(public chi.Router) {
		public.Use(authmw.RateLimitByIP(iplimiter))
		public.Post("/register", registerHandeler.ServeHTTP)
		public.Post("/login", loginHandler.ServeHTTP)
		public.Post("/refresh", refreshHandler.ServeHTTP)
		public.Post("/verify-email", verifyEmailHandler.ServeHTTP)
		public.Post("/forget-password", forgetpasswordHandler.ServeHTTP)
		public.Post("/reset-password", resetPassworedHandeler.ServeHTTP)
		public.Post("/logout", logoutHandler.ServeHTTP)
	})

	r.Group(func(protect chi.Router) {
		protect.Use(authmw.RequireAuth(cfg.JWTSecret))
		protect.Get("/me", meHandler.ServeHTTP)
		protect.Post("/phone/send-otp", sendPhoneOTPHandler.ServeHTTP)
		protect.Post("/phone/verify-otp", verifyPhoneOTPHandler.ServeHTTP)
	})

	log.Printf("auth-service listening on :%s", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, r); err != nil {
		log.Fatal(err)
	}
}
