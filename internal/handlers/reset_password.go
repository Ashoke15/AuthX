package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/Ashoke15/AuthX/internal/auth"
	"github.com/Ashoke15/AuthX/internal/repository"
	"github.com/Ashoke15/AuthX/internal/validation"
)

type ResetPasswordHandler struct {
	userRepo    repository.UserReposerty
	resetRepo   repository.PRRepo
	refreshRepo repository.RefreshTokenRepository
}

func NewResetPasswordHander(userRepo repository.UserReposerty, resetRepo repository.PRRepo, refreshRepo repository.RefreshTokenRepository) *ResetPasswordHandler {
	return &ResetPasswordHandler{userRepo: userRepo, resetRepo: resetRepo, refreshRepo: refreshRepo}
}

type resetPasswordRequest struct {
	Email string `json:"email"`
	Code string `json:"code"`
	NewPassword string `json:"new_password"`
}

func (h *ResetPasswordHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var req resetPasswordRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := validation.ValidateEmail(req.Email); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := validation.ValidatePassword(req.NewPassword); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	if req.Code == "" {
		writeError(w, http.StatusBadRequest, "code is required")
		return
	}

	user, err := h.userRepo.GetByEmail(req.Email)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "invalid email or code")
		return
	}

	reset, err := h.resetRepo.GetLatestByUserId(user.Id)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "invalid email or code")
		return
	}

	if reset.UsedAt != nil {
		writeError(w, http.StatusUnauthorized, "code alredy used request a new one")
		return
	}

	if time.Now().After(reset.ExpiresAt) {
		writeError(w, http.StatusUnauthorized, "code has expired request a new one")
		return
	}

	if reset.Attempts >= auth.MaxOtpAttempts {
		writeError(w, http.StatusTooManyRequests, "too many attempts, request a new code")
		return
	}

	if auth.HashOTP(req.Code) != reset.CodeHash {
		h.resetRepo.IncrementsAttempts(reset.ID)
		writeError(w, http.StatusUnauthorized, "Invalid email or code")
		return
	}

	newHash, err := auth.HashPassword(req.NewPassword)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not process new passwored")
		return
	}

	if err := h.userRepo.UpdatePasswored(user.Id, newHash); err != nil {
		writeError(w, http.StatusInternalServerError, "could not reset passwored")
		return
	}

	if err := h.resetRepo.MarkUsed(reset.ID); err != nil {
		writeError(w, http.StatusInternalServerError, "could not complete reset")
		return
	}

	_ = h.refreshRepo.RevokedAllByUserId(user.Id)

	writeJson(w, http.StatusOK, map[string]string{"message": "passwored reset successful"})
}
