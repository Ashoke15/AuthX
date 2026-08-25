package repository

import (
	"database/sql"
	"errors"

	"github.com/Ashoke15/AuthX/internal/models"
	"github.com/jackc/pgx/v5/pgconn"
)

var (
	ErrEmailTaken   = errors.New("email alredy registerd")
	ErrUserNotFound = errors.New("user not found")
)

type UserReposerty interface {
	Create(u *models.User) error
	GetByEmail(email string) (*models.User, error)
}

type PURepository struct {
	db *sql.DB
}

func NPURepository(db *sql.DB) *PURepository {
	return &PURepository{db: db}
}

func (r *PURepository) Create(u *models.User) error {
	err := r.db.QueryRow(
		`INSERT INTO users(id, email, password_hash, email_verified)
		VALUES ($1, $2, $3, $4)
		RETURNING created_at, updated_at`,
		u.Id, u.Email, u.PasswordHash, u.EmailVerified,
	).Scan(&u.CreatedAt, &u.UpdatedAt)

	if err != nil {
		if isUniqueError(err) {
			return ErrEmailTaken
		}
		return err
	}

	return nil
}

func (r *PURepository) GetByEmail(email string) (*models.User, error) {
	row := r.db.QueryRow(
		`SELECT id, email, password_hash, email_verified, created_at, updated_at
		FROM users WHERE email = $1`,
		email,
	)

	var u models.User
	err := row.Scan(&u.Id, &u.Email, &u.PasswordHash, &u.EmailVerified, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	return &u, nil
}

func isUniqueError(err error) bool {
	var pgErr *pgconn.PgError

	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505"
	}

	return false
}