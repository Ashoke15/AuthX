package repository

import (
	"database/sql"
	"errors"

	"github.com/Ashoke15/AuthX/internal/models"
)

var ErrVerificationNotFound = errors.New("verification code not found")

type EmailVerificationRepo interface {
	Create(ev *models.EmailVerification) error
	GetLatestbyUserId(userId string) (*models.EmailVerification, error)
	IncrementsAttempts(id string) error
	MarkUsed(id string) error
	MarkEmailVerified(userId string) error
}

type PGEVRepo struct {
	db *sql.DB
}

func NewPGEVRepo(db *sql.DB) *PGEVRepo {
	return &PGEVRepo{db: db}
}

func (r *PGEVRepo) Create(ev *models.EmailVerification) error {
	return r.db.QueryRow(
		`INSERT INTO email_verifications (id, user_id, code_hash, expires_at)
		VALUES ($1, $2, $3, $4)
		RETURNING created_at`,
		ev.ID, ev.UserId, ev.CodeHash, ev.ExpiresAt,
	).Scan(&ev.CreatedAt)
}

func (r *PGEVRepo) GetLatestbyUserId(userId string) (*models.EmailVerification, error) {
	row := r.db.QueryRow(
		`SELECT id, user_id, code_hash, expires_at, attempts, used_at, created_at
		FROM email_verifications WHERE user_id = $1
		ORDER BY created_at DESC LIMIT 1`,
		userId,
	)

	var ev models.EmailVerification

	err := row.Scan(&ev.ID, &ev.UserId, &ev.CodeHash, &ev.ExpiresAt, &ev.Attempts, &ev.UsedAt, &ev.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrVerificationNotFound
		}
		return nil, err
	}

	return &ev, nil
}

func (r *PGEVRepo) IncrementsAttempts(id string) error {
	_, err := r.db.Exec(`UPDATE email_verifications SET attempts = attempts +1 WHERE id = $1`, id)
	return err
}

func (r *PGEVRepo) MarkUsed(id string) error {
	_, err := r.db.Exec(`UPDATE email_verifications SET used_at = NOW() WHERE id = $1`, id)
	return err
}

func (r *PGEVRepo) MarkEmailVerified(userId string) error {
	_, err := r.db.Exec(`UPDATE users SET email_verified = TRUE, updated_at = NOW() WHERE id = $1`, userId)
	return err
}
