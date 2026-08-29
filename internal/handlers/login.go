package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/Ashoke15/AuthX/internal/auth"
	"github.com/Ashoke15/AuthX/internal/models"
	"github.com/Ashoke15/AuthX/internal/ratelimit"
	"github.com/Ashoke15/AuthX/internal/repository"
	"github.com/Ashoke15/AuthX/internal/validation"
	"github.com/google/uuid"
)

type LoginHandler struct {
	userRepo    repository.UserReposerty
	refreshRepo repository.RefreshTokenRepository
	JwtSecret   string
	emailLimiter *ratelimit.Limiter
}

func NewLoginHandler(repo repository.UserReposerty, refreshRepo repository.RefreshTokenRepository, jwtSecret string, emailLimiter *ratelimit.Limiter) *LoginHandler {
	return &LoginHandler{userRepo: repo, refreshRepo: refreshRepo, JwtSecret: jwtSecret, emailLimiter: emailLimiter}
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type loginResponce struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
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

	if !h.emailLimiter.Allow(req.Email) {
		writeError(w, http.StatusTooManyRequests, "too many login attempts, please try again later")
		return
	}

	if req.Password == "" {
		writeError(w, http.StatusBadRequest, "password is required")
		return
	}

	user, err := h.userRepo.GetByEmail(req.Email)
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

	if !user.EmailVerified {
		writeError(w, http.StatusForbidden, "please verifiy email before login")
		return
	}

	token, err := auth.GenerateAccessToken(user.Id, h.JwtSecret)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not generate token")
		return
	}

	refreshPlain, err := auth.GenerateRefreshToken()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not generate refresh token")
		return
	}

	refreshToken := &models.RefreshToken{
		Id:         uuid.NewString(),
		UserId:     user.Id,
		TokenHash:  auth.HashRefreashToken(refreshPlain),
		ExpairesAt: time.Now().Add(auth.RefreshTokenTTL),
	}

	if err := h.refreshRepo.Create(refreshToken); err != nil {
		writeError(w, http.StatusInternalServerError, "could not store refresh token")
		return
	}

	writeJson(w, http.StatusOK, loginResponce{
		AccessToken:  token,
		RefreshToken: refreshPlain,
		ExpiresIn:    int(auth.AccessTokenTTL.Seconds()),
	})
}
