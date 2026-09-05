package repository

import (
	"database/sql"
	"errors"
	"time"

	"github.com/Ashoke15/AuthX/internal/models"
	"github.com/jackc/pgx/v5/pgconn"
)

var (
	ErrEmailTaken   = errors.New("email alredy registerd")
	ErrPhoneTaken   = errors.New("phone number alredy registerd")
	ErrUserNotFound = errors.New("user not found")
)

type UserReposerty interface {
	Create(u *models.User) error
	GetByEmail(email string) (*models.User, error)
	GetBYId(id string) (*models.User, error)
	UpdatePasswored(userId, passworedHash string) error
	IncrementFailedAttempts(userId string) (int, error)
	LockAccount(userId string, until time.Time) error
	ResetFailedAttempts(userId string) error
	SetPhone(userId, phone string) error
	VerifyPhone(userId string) error
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
		`SELECT id, email, password_hash, email_verified, phone, phone_verified, failed_login_attempts, locked_until, created_at, updated_at
		FROM users WHERE email = $1`,
		email,
	)

	return scanRow(row)
}

func (r *PURepository) GetBYId(id string) (*models.User, error) {
	row := r.db.QueryRow(
		`SELECT id, email, password_hash, email_verified, phone, phone_verified, failed_login_attempts, locked_until, created_at, updated_at
		FROM users WHERE id = $1`,
		id,
	)

	return scanRow(row)
}

func scanRow(row *sql.Row) (*models.User, error) {
	var u models.User

	err := row.Scan(&u.Id, &u.Email, &u.PasswordHash, &u.EmailVerified, &u.Phone, &u.PhoneVerified, &u.FailedLoginAttempts, &u.LockedUntil, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	return &u, nil
}

func (r *PURepository) UpdatePasswored(userId, passwordHash string) error {
	_, err := r.db.Exec(`UPDATE users SET password_hash = $1, updated_at = NOW() WHERE id = $2`, passwordHash, userId)
	return err
}

func (r *PURepository) IncrementFailedAttempts(userId string) (int, error) {
	var attempts int

	err := r.db.QueryRow(
		`UPDATE users SET failed_login_attempts = failed_login_attempts + 1, updated_at = NOW()
		WHERE id = $1 RETURNING failed_login_attempts`,
		userId,
	).Scan(&attempts)

	return attempts, err
}

func (r *PURepository) LockAccount(userId string, until time.Time) error {
	_, err := r.db.Exec(`UPDATE users SET locked_until = $1 , updated_at = NOW() WHERE id = $2`, until, userId)

	return err
}

func (r *PURepository) ResetFailedAttempts(userId string) error {
	_, err := r.db.Exec(`UPDATE users SET failed_login_attempts = 0, locked_until = NULL, updated_at = NOW() WHERE id = $1`, userId)

	return err
}

func (r *PURepository) SetPhone(userId, phone string) error {
	_, err := r.db.Exec(`UPDATE users SET phone = $1, phone_verified = FALSE, updated_at = NOW() WHERE id = $2`, phone, userId)
	if err != nil {
		if isUniqueError(err) {
			return ErrPhoneTaken
		}
		return err
	}
	return nil
}

func (r *PURepository) VerifyPhone(userId string) error {
	_, err := r.db.Exec(`UPDATE users SET phone_verified = TRUE , updated_at = NOW() WHERE id = $1`, userId)
	return err
}

func isUniqueError(err error) bool {
	var pgErr *pgconn.PgError

	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505"
	}

	return false
}
