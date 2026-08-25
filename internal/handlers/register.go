package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/mail"
	"strings"
	"time"

	"github.com/Ashoke15/AuthX/internal/auth"
	"github.com/Ashoke15/AuthX/internal/models"
	"github.com/Ashoke15/AuthX/internal/repository"
	"github.com/google/uuid"
)

type RegisterHandeler struct {
	repo repository.UserReposerty
}

func NewRegisterHandeler(repo repository.UserReposerty) *RegisterHandeler {
	return &RegisterHandeler{repo: repo}
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

	writeJson(w, http.StatusCreated, registerResponce{
		ID:            user.Id,
		Email:         user.Email,
		CreatedAt:     user.CreatedAt,
		EmailVerified: user.EmailVerified,
	})
}

func validateRegisterRequest(req registerRequest) error {
	if req.Email == "" || req.Password == "" {
		return errors.New("email and password are required")
	}

	if _, err := mail.ParseAddress(req.Email); err != nil {
		return errors.New("invalid email format")
	}

	if len(req.Password) < 8 {
		return errors.New("passwored must be 8 characters")
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
