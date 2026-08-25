package models

import "time"

type User struct {
	Id            string `json:"id"`
	Email         string `json:"email"`
	PasswordHash  string `json:"-"`
	EmailVerified bool `json:"email_verified"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}
