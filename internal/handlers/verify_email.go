package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/Ashoke15/AuthX/internal/auth"
	"github.com/Ashoke15/AuthX/internal/repository"
	"github.com/Ashoke15/AuthX/internal/validation"
)

type VerifyEmailHandler struct {
	userRepo   repository.UserReposerty
	verifyRepo repository.EmailVerificationRepo
}

func NewVerifyEmailHandler(userrepo repository.UserReposerty, verifyRepo repository.EmailVerificationRepo) *VerifyEmailHandler {
	return &VerifyEmailHandler{userRepo: userrepo, verifyRepo: verifyRepo}
}

type verifyEmailRequest struct {
	Email string `json:"email"`
	Code  string `json:"code"`
}

func (h *VerifyEmailHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var req verifyEmailRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := validation.ValidateEmail(req.Email); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	if req.Code == "" {
		writeError(w, http.StatusBadRequest, "code is require")
		return
	}

	user, err := h.userRepo.GetByEmail(req.Email)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "invalid email or code")
		return
	}

	if user.EmailVerified {
		writeError(w, http.StatusConflict, "email alredy verified")
		return
	}

	verification, err := h.verifyRepo.GetLatestbyUserId(user.Id)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "invalid email or code")
		return
	}

	if verification.UsedAt != nil {
		writeError(w, http.StatusUnauthorized, "code alredy used, request a new one")
		return
	}

	if time.Now().After(verification.ExpiresAt) {
		writeError(w, http.StatusUnauthorized, "code has expaire, request a new one")
		return
	}

	if verification.Attempts >= auth.MaxOtpAttempts {
		writeError(w, http.StatusTooManyRequests, "too mant attempts request a new code")
		return
	}

	if auth.HashOTP(req.Code) != verification.CodeHash {
		h.verifyRepo.IncrementsAttempts(verification.ID)
		writeError(w, http.StatusUnauthorized, "invalid email or code")
		return
	}

	if err := h.verifyRepo.MarkUsed(verification.ID); err != nil {
		writeError(w, http.StatusInternalServerError, "could not complete verification")
		return
	}

	if err := h.verifyRepo.MarkEmailVerified(user.Id); err != nil {
		writeError(w, http.StatusInternalServerError, "could not complete verification")
		return
	}

	writeJson(w, http.StatusOK, map[string]string{"message": "email verified successfully"})
}
