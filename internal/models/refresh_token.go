package models

import "time"

type RefreshToken struct {
	Id         string
	UserId     string
	TokenHash  string
	ExpairesAt time.Time
	CreatedAt  time.Time
	RevokedAt  *time.Time
}
