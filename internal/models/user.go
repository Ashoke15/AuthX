package models

import "time"

type User struct {
	Id                  string     `json:"id"`
	Email               string     `json:"email"`
	PasswordHash        string     `json:"-"`
	EmailVerified       bool       `json:"email_verified"`
	Phone               *string    `json:"phone,omitempty"`
	PhoneVerified       bool       `json:"phone_verified"`
	FailedLoginAttempts int        `json:"-"`
	LockedUntil         *time.Time `json:"-"`
	CreatedAt           time.Time  `json:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at"`
}
