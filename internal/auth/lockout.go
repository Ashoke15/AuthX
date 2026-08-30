package auth

import "time"

const (
	MaxFailedLoginAttempts = 5
	LockOutDuration        = 15 * time.Minute
)