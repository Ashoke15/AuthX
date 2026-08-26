package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/Ashoke15/AuthX/internal/auth"
	"github.com/Ashoke15/AuthX/internal/repository"
)

type RefreshHandeler struct {
	refreshRepo repository.RefreshTokenRepository
	jwtSecret   string
}

func NewRefreshHandeler(refreshRepo repository.RefreshTokenRepository, jwtSecret string) *RefreshHandeler {
	return &RefreshHandeler{refreshRepo: refreshRepo, jwtSecret: jwtSecret}
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type RefreshResponce struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expaires_in"`
}

func (h *RefreshHandeler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var req RefreshRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.RefreshToken == "" {
		writeError(w, http.StatusBadRequest, "Refresh Token is Require")
		return
	}

	stored, err := h.refreshRepo.GetByHash(auth.HashRefreashToken(req.RefreshToken))
	if err != nil {
		if errors.Is(err, repository.ErrRefreshTokenNotFound) {
			writeError(w, http.StatusUnauthorized, "invalid refresh token")
			return
		}

		writeError(w, http.StatusInternalServerError, "could not validate refresh token")
		return
	}

	if stored.RevokedAt != nil {
		writeError(w, http.StatusUnauthorized, "refresh token has been revoked")
		return
	}

	if time.Now().After(stored.ExpairesAt) {
		writeError(w, http.StatusUnauthorized, "refresh token has expaire")
		return
	}

	accessToken, err := auth.GenerateAccessToken(stored.UserId, h.jwtSecret)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not generate access token")
		return
	}

	writeJson(w, http.StatusOK, RefreshResponce{
		AccessToken: accessToken,
		ExpiresIn:   int(auth.AccessTokenTTL.Seconds()),
	})
}
