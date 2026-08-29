package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/Ashoke15/AuthX/internal/auth"
	"github.com/Ashoke15/AuthX/internal/mailer"
	"github.com/Ashoke15/AuthX/internal/models"
	"github.com/Ashoke15/AuthX/internal/ratelimit"
	"github.com/Ashoke15/AuthX/internal/repository"
	"github.com/Ashoke15/AuthX/internal/validation"
	"github.com/google/uuid"
)

type ForgetPasswordHandler struct {
	userRepo  repository.UserReposerty
	resetRepo repository.PRRepo
	mailer    mailer.Mailer
	emailLimit *ratelimit.Limiter
}

func NewForgetPasswordHandeler(userRepo repository.UserReposerty, resetRepo repository.PRRepo, m mailer.Mailer, emailLimit *ratelimit.Limiter) *ForgetPasswordHandler {
	return &ForgetPasswordHandler{userRepo: userRepo, resetRepo: resetRepo, mailer: m, emailLimit: emailLimit}
}

type forgetpasswordRequest struct {
	Email string `json:"email"`
}

func (h *ForgetPasswordHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var req forgetpasswordRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := validation.ValidateEmail(req.Email); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	if !h.emailLimit.Allow(req.Email) {
		writeError(w, http.StatusTooManyRequests, "too many request, please try again later")
		return
	}

	genericResponce := map[string]string{
		"message": "if the email is registered, a reset code has been sent",
	}

	user, err := h.userRepo.GetByEmail(req.Email)
	if err != nil {
		writeJson(w, http.StatusOK, genericResponce)
		return
	}

	code, err := auth.GenerateOTP()
	if err != nil {
		log.Printf("failed to generate reset code: %v", err)
		writeJson(w, http.StatusOK, genericResponce)
		return
	}

	reset := &models.PasswordReset{
		ID: uuid.NewString(),
		UserId: user.Id,
		CodeHash: auth.HashOTP(code),
		ExpiresAt: time.Now().Add(auth.OTPTTL),
	}

	if err := h.resetRepo.Create(reset); err != nil {
		log.Printf("failed to store reset code for %s: %v", user.Email, err)
		writeJson(w, http.StatusOK, genericResponce)
		return
	}

	if err := h.mailer.SendPasswordResetEmail(user.Email, code); err != nil {
		log.Printf("failed to sent reset email to %s: %v", user.Email, err)
	}

	writeJson(w, http.StatusOK, genericResponce)
}
