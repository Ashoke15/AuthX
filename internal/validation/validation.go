package validation

import (
	"errors"
	"net/mail"
)

const MinPasswordLength = 8

var (
	ErrEmailRequired = errors.New("email is required")
	ErrEmailInvalid = errors.New("invalid email format")
	ErrPasswordRequired = errors.New("password is required")
	ErrPasswordTooShort = errors.New("passwored must be at least 8 characters")
)

func ValidateEmail(email string) error {
	if email == "" {
		return ErrEmailRequired
	}

	addr, err := mail.ParseAddress(email)
	if err != nil || addr.Address != email {
		return ErrEmailInvalid
	}

	return nil
}

func ValidatePassword(password string) error {
	if password == "" {
		return ErrPasswordRequired
	}

	if len(password) < MinPasswordLength {
		return ErrPasswordTooShort
	}

	return nil
}