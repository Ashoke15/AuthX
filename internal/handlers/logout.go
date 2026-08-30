package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/Ashoke15/AuthX/internal/auth"
	"github.com/Ashoke15/AuthX/internal/repository"
)

type LogoutHandler struct {
	refreshRepo repository.RefreshTokenRepository
}

func NewLogoutHandler(refreshRepo repository.RefreshTokenRepository) *LogoutHandler {
	return &LogoutHandler{refreshRepo: refreshRepo}
}

type logoutRequest struct {
	RefreshToken string `json:"refresh_token"`
}

func (h *LogoutHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var req logoutRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.RefreshToken == "" {
		writeError(w, http.StatusBadRequest, "refresh_token is require")
		return
	}

	stored, err := h.refreshRepo.GetByHash(auth.HashRefreashToken(req.RefreshToken))
	if err == nil {
		_ = h.refreshRepo.Revoked(stored.Id)
	}

	writeJson(w, http.StatusOK, map[string]string{"message": "logged out"})
}
