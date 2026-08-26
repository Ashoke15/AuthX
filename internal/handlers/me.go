package handlers

import (
	"errors"
	"net/http"

	"github.com/Ashoke15/AuthX/internal/middleware"
	"github.com/Ashoke15/AuthX/internal/repository"
)

type MeHandler struct {
	repo repository.UserReposerty
}

func NewMeHandler(repo repository.UserReposerty) *MeHandler {
	return &MeHandler{repo: repo}
}

type meResponse struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
}

func (h *MeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	userId, ok := middleware.UserIdFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "not authenticated")
		return
	}

	user, err := h.repo.GetBYId(userId)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			writeError(w, http.StatusNotFound, "user not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "could not load user")
		return
	}

	writeJson(w, http.StatusOK, meResponse{
		ID:            user.Id,
		Email:         user.Email,
		EmailVerified: user.EmailVerified,
	})
}
