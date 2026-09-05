package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/Ashoke15/AuthX/internal/auth"
	"github.com/Ashoke15/AuthX/internal/middleware"
	"github.com/Ashoke15/AuthX/internal/repository"
)

type VerifyPhoneOTPHandler struct {
	userRepo repository.UserReposerty
	phoneRepo repository.PhoneVerificationRepo
}

func NewVerifyPhoneOTPHandler(userRepo repository.UserReposerty, phoneRepo repository.PhoneVerificationRepo) *VerifyPhoneOTPHandler {
	return &VerifyPhoneOTPHandler{userRepo: userRepo, phoneRepo: phoneRepo}
}

type VerifyPhoneOTPRequest struct {
	Code string `json:"code"`
}

func (h *VerifyPhoneOTPHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	userId, ok := middleware.UserIdFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "Not authincated")
		return
	}

	var req VerifyPhoneOTPRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Code == "" {
		writeError(w, http.StatusBadRequest, "code is require")
		return
	}

	verificarion, err := h.phoneRepo.GetLatestbyUserId(userId)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "invalid or expaire code")
		return
	}

	if verificarion.UsedAt != nil {
		writeError(w, http.StatusUnauthorized, "code alredy used, request a new one")
		return
	}

	if time.Now().After(verificarion.ExpiresAt) {
		writeError(w, http.StatusUnauthorized, "code has expaire request a new one")
		return
	}

	if verificarion.Attempts >= auth.MaxOtpAttempts {
		writeError(w, http.StatusTooManyRequests, "too many attempts, request a new code")
		return
	}

	if auth.HashOTP(req.Code) != verificarion.CodeHash {
		h.phoneRepo.IncrementsAttempts(verificarion.ID)
		writeError(w, http.StatusUnauthorized, "invalid or expaire code")
		return
	}

	if err := h.phoneRepo.MarkUsed(verificarion.ID); err != nil {
		writeError(w, http.StatusInternalServerError, "could not complete verification")
		return
	}

	if err := h.userRepo.VerifyPhone(userId); err != nil {
		writeError(w, http.StatusInternalServerError, "could not complete verification")
		return
	}

	writeJson(w, http.StatusOK, map[string]string{"message": "phone verified successfully"})
}