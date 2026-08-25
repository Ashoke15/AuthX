package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/Ashoke15/AuthX/internal/auth"
	"github.com/Ashoke15/AuthX/internal/repository"
	"github.com/Ashoke15/AuthX/internal/validation"
)

type LoginHandler struct {
	repo      repository.UserReposerty
	JwtSecret string
}

func NewLoginHandler(repo repository.UserReposerty, jwtSecret string) *LoginHandler {
	return &LoginHandler{repo: repo, JwtSecret: jwtSecret}
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type loginResponce struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
}

func (h *LoginHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var req loginRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	req.Email = strings.TrimSpace(strings.ToLower(req.Email))

	if err := validation.ValidateEmail(req.Email); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	if req.Password == "" {
		writeError(w, http.StatusBadRequest, "password is required")
		return
	}

	user, err := h.repo.GetByEmail(req.Email)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			writeError(w, http.StatusUnauthorized, "invalid email and password")
			return
		}
		writeError(w, http.StatusInternalServerError, "could not process login")
		return
	}

	if !auth.CheckPassword(user.PasswordHash, req.Password) {
		writeError(w, http.StatusUnauthorized, "invalid email and password")
		return
	}

	token, err := auth.GenerateAccessToken(user.Id, h.JwtSecret)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not generate token")
		return
	}

	writeJson(w, http.StatusOK, loginResponce{
		AccessToken: token,
		ExpiresIn:   int(auth.AccessTokenTTL.Seconds()),
	})
}
