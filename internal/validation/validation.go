package validation

import (
	"errors"
	"net/mail"
	"regexp"
)

const MinPasswordLength = 8

var (
	ErrEmailRequired = errors.New("email is required")
	ErrEmailInvalid = errors.New("invalid email format")
	ErrPasswordRequired = errors.New("password is required")
	ErrPasswordTooShort = errors.New("passwored must be at least 8 characters")
	ErrPhoneRequire = errors.New("phone number is require")
	ErrPhoneInvalid = errors.New("phone number must be in E.164 format, e.g. +919876897468")
)

var phoneRegex = regexp.MustCompile(`^\+[1-9]\d{7,14}$`)

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

func ValidatePhone(phone string) error {
	if phone == "" {
		return ErrPhoneRequire
	}

	if !phoneRegex.MatchString(phone) {
		return ErrPhoneInvalid
	}

	return nil
}