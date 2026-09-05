package repository

import (
	"database/sql"
	"errors"

	"github.com/Ashoke15/AuthX/internal/models"
)

var ErrPhoneVerificationNotFound = errors.New("Phone Verification Code Not Found")

type PhoneVerificationRepo interface {
	Create(pv *models.PhoneVerification) error
	GetLatestbyUserId(userId string) (*models.PhoneVerification, error)
	IncrementsAttempts(id string) error
	MarkUsed(id string) error
}

type PGPVRepo struct {
	db *sql.DB
}

func NewPGPVRepo(db *sql.DB) *PGPVRepo {
	return &PGPVRepo{db: db}
}

func (r *PGPVRepo) Create(pv *models.PhoneVerification) error {
	return r.db.QueryRow(
		`INSERT INTO phone_verifications (id, user_id, code_hash, expires_at)
		VALUES ($1, $2, $3, $4)
		RETURNING created_at`,
		pv.ID, pv.UserId, pv.CodeHash, pv.ExpiresAt,
	).Scan(&pv.CreatedAt)
}

func (r *PGPVRepo) GetLatestbyUserId(userId string) (*models.PhoneVerification, error) {
	row := r.db.QueryRow(
		`SELECT id, user_id, code_hash, expires_at, attempts, used_at, created_at
		FROM phone_verifications WHERE user_id = $1
		ORDER BY created_at DESC LIMIT 1`,
		userId,
	)

	var pv models.PhoneVerification

	err := row.Scan(&pv.ID, &pv.UserId, &pv.CodeHash, &pv.ExpiresAt, &pv.Attempts, &pv.UsedAt, &pv.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return  nil, ErrPhoneVerificationNotFound
		}
		return nil, err
	}

	return &pv, nil
}

func (r *PGPVRepo) IncrementsAttempts(id string) error {
	_, err := r.db.Exec(`UPDATE phone_verifications SET attempts = attempts + 1 WHERE id = $1`, id)
	return err
}

func (r *PGPVRepo) MarkUsed(id string) error {
	_, err := r.db.Exec(`UPDATE phone_verifications SET used_at = NOW() WHERE id = $1`, id)
	return err
}
