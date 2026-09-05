package handlers

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/Ashoke15/AuthX/internal/auth"
	"github.com/Ashoke15/AuthX/internal/middleware"
	"github.com/Ashoke15/AuthX/internal/models"
	"github.com/Ashoke15/AuthX/internal/repository"
	"github.com/Ashoke15/AuthX/internal/sms"
	"github.com/Ashoke15/AuthX/internal/validation"
	"github.com/google/uuid"
)

type SendPhoneOTPHandler struct {
	userRepo  repository.UserReposerty
	phoneRepo repository.PhoneVerificationRepo
	Sender    sms.Sender
}

func NewSentPhoneOTPHandler(userRepo repository.UserReposerty, phoneRepo repository.PhoneVerificationRepo, Sender sms.Sender) *SendPhoneOTPHandler {
	return &SendPhoneOTPHandler{userRepo: userRepo, phoneRepo: phoneRepo, Sender: Sender}
}

type sendPhoneOTPRequest struct {
	Phone string `json:"phone"`
}

func (h *SendPhoneOTPHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	userId, ok := middleware.UserIdFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "not authenticated")
		return
	}

	var req sendPhoneOTPRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := validation.ValidatePhone(req.Phone); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.userRepo.SetPhone(userId, req.Phone); err != nil {
		if errors.Is(err, repository.ErrPhoneTaken) {
			writeError(w, http.StatusConflict, "phone number alredy register to another account")
			return
		}
		writeError(w, http.StatusInternalServerError, "could not set phone number")
		return
	}

	code, err := auth.GenerateOTP()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not generate the otp")
		return
	}

	verification := &models.PhoneVerification{
		ID:        uuid.NewString(),
		UserId:    userId,
		CodeHash:  auth.HashOTP(code),
		ExpiresAt: time.Now().Add(auth.OTPTTL),
	}

	if err := h.phoneRepo.Create(verification); err != nil {
		writeError(w, http.StatusInternalServerError, "could not store code")
		return
	}

	if err := h.Sender.SendOTPSMS(req.Phone, code); err != nil {
		log.Printf("failed to send otp sms to %s:%v", req.Phone, err)
	}

	writeJson(w, http.StatusOK, map[string]string{"message": "verification code sent"})
}
