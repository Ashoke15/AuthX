package models

import "time"

type EmailVerification struct {
	ID        string
	UserId    string
	CodeHash  string
	ExpiresAt time.Time
	Attempts  int
	UsedAt    *time.Time
	CreatedAt time.Time
}
