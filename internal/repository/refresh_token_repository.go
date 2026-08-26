package repository

import (
	"database/sql"
	"errors"

	"github.com/Ashoke15/AuthX/internal/models"
)

var (
	ErrRefreshTokenNotFound = errors.New("refresh token not found")
)

type RefreshTokenRepository interface {
	Create(rt *models.RefreshToken) error
	GetByHash(tokenHash string) (*models.RefreshToken, error)
	Revoked(id string) error
}

type PRTRepository struct {
	db *sql.DB
}

func NPRTRepositry(db *sql.DB) *PRTRepository {
	return &PRTRepository{db: db}
}

func (r *PRTRepository) Create(rt *models.RefreshToken) error {
	return r.db.QueryRow(
		`INSERT INTO refresh_tokens(id, user_id, token_hash, expires_at)
		VALUES ($1, $2, $3, $4)
		RETURNING created_at`,
		rt.Id, rt.UserId, rt.TokenHash, rt.ExpairesAt,
	).Scan(&rt.CreatedAt)
}

func (r *PRTRepository) GetByHash(tokenHash string) (*models.RefreshToken, error) {
	row := r.db.QueryRow(
		`SELECT id, user_id, token_hash, expires_at, created_at, revoked_at
		FROM refresh_tokens WHERE token_hash = $1`,
		tokenHash,
	)

	var rt models.RefreshToken

	err := row.Scan(&rt.Id, &rt.UserId, &rt.TokenHash, &rt.ExpairesAt, &rt.CreatedAt, &rt.RevokedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrRefreshTokenNotFound
		}
		return nil, err
	}

	return &rt, nil
}

func (r *PRTRepository) Revoked(id string) error {
	_, err := r.db.Exec(`UPDATE refresh_tokens SET revoked_at = NOW() WHERE id = $1`, id)

	return err
}
