package repository

import (
	"database/sql"
	"errors"

	"github.com/Ashoke15/AuthX/internal/models"
)

var ErrPasswordResetNotFound = errors.New("passwored reset code not found")

type PRRepo interface {
	Create(pr *models.PasswordReset) error
	GetLatestByUserId(userId string) (*models.PasswordReset, error)
	IncrementsAttempts(id string) error
	MarkUsed(id string) error
}

type PgPRRepo struct {
	db *sql.DB
}

func NPgPRRepo(db *sql.DB) *PgPRRepo {
	return &PgPRRepo{db: db}
}

func (r *PgPRRepo) Create(pr *models.PasswordReset) error {
	return r.db.QueryRow(
		`INSERT INTO password_resets (id, user_id, code_hash, expires_at)
		VALUES($1, $2, $3, $4)
		RETURNING created_at`,
		pr.ID, pr.UserId, pr.CodeHash, pr.ExpiresAt,
	).Scan(&pr.CreatedAt)
}

func (r *PgPRRepo) GetLatestByUserId(userId string) (*models.PasswordReset, error) {
	row := r.db.QueryRow(
		`SELECT id, user_id, code_hash, expires_at, attempts, used_at, created_at
		FROM password_resets WHERE user_id = $1
		ORDER BY created_at DESC LIMIT 1`,
		userId,
	)

	var pr models.PasswordReset

	err := row.Scan(&pr.ID, &pr.UserId, &pr.CodeHash, &pr.ExpiresAt, &pr.Attempts, &pr.UsedAt, &pr.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrPasswordResetNotFound
		}
		return nil, err
	}

	return &pr, nil
}

func (r *PgPRRepo) IncrementsAttempts(id string) error {
	_, err := r.db.Exec(`UPDATE password_resets SET attempts = attempts + 1 WHERE id = $1`, id)

	return err
}

func (r *PgPRRepo) MarkUsed(id string) error {
	_, err := r.db.Exec(`UPDATE password_resets SET used_at = NOW() WHERE id = $1`, id)
	return err
}
