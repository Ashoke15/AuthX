package handlers

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/Ashoke15/AuthX/internal/auth"
	"github.com/Ashoke15/AuthX/internal/mailer"
	"github.com/Ashoke15/AuthX/internal/models"
	"github.com/Ashoke15/AuthX/internal/repository"
	"github.com/Ashoke15/AuthX/internal/validation"
	"github.com/google/uuid"
)

type RegisterHandeler struct {
	repo repository.UserReposerty
	verifyRepo repository.EmailVerificationRepo
	mailer *mailer.SMTPMailer
}

func NewRegisterHandeler(repo repository.UserReposerty, verifyRepo repository.EmailVerificationRepo, mailer *mailer.SMTPMailer) *RegisterHandeler {
	return &RegisterHandeler{repo: repo, verifyRepo: verifyRepo, mailer: mailer}
}

type registerRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type registerResponce struct {
	ID            string    `json:"id"`
	Email         string    `json:"email"`
	EmailVerified bool      `json:"email_verified"`
	CreatedAt     time.Time `json:"created_at"`
}

func (h *RegisterHandeler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var req registerRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	req.Email = strings.TrimSpace(strings.ToLower(req.Email))

	if err := validateRegisterRequest(req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	hash, err := auth.HashPassword(req.Password)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not process password")
		return
	}

	user := &models.User{
		Id:           uuid.NewString(),
		Email:        req.Email,
		PasswordHash: hash,
	}

	if err := h.repo.Create(user); err != nil {
		if errors.Is(err, repository.ErrEmailTaken) {
			writeError(w, http.StatusConflict, "email alredy registerd")
			return
		}
		writeError(w, http.StatusInternalServerError, "could not create user")
		return
	}

	code, err := auth.GenerateOTP()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not generate verification code")
		return
	}

	verification := &models.EmailVerification{
		ID: uuid.NewString(),
		UserId: user.Id,
		CodeHash: auth.HashOTP(code),
		ExpiresAt: time.Now().Add(auth.OTPTTL),
	}

	if err := h.verifyRepo.Create(verification); err != nil {
		writeError(w, http.StatusInternalServerError, "could not store verfication code")
		return
	}

	if err := h.mailer.SendVerificationEmail(user.Email, code); err != nil {
		log.Printf("sent verification emailto %s : %v", user.Email, err)
	}

	writeJson(w, http.StatusCreated, registerResponce{
		ID:            user.Id,
		Email:         user.Email,
		CreatedAt:     user.CreatedAt,
		EmailVerified: user.EmailVerified,
	})
}

func validateRegisterRequest(req registerRequest) error {
	if err := validation.ValidateEmail(req.Email); err != nil {
		return err
	}

	if err := validation.ValidatePassword(req.Password); err != nil {
		return err
	}

	return nil
}

func writeJson(w http.ResponseWriter, status int, paylod any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(paylod)
}

func writeError(w http.ResponseWriter, status int, massege string) {
	writeJson(w, status, map[string]string{"error": massege})
}
