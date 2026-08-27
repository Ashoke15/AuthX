package db

import (
	"database/sql"
	"fmt"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func New(dsn string) (*sql.DB, error) {
	cnn, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("Open db: %w", err)
	}

	if err := cnn.Ping(); err != nil {
		return nil, fmt.Errorf("ping db: %w", err)
	}

	if err := migrate(cnn); err != nil {
		return nil, fmt.Errorf("migrate: %w", err)
	}

	return cnn, nil
}

func migrate(conn *sql.DB) error {
	schema := `
	CREATE TABLE IF NOT EXISTS users(
		id UUID PRIMARY KEY,
		email TEXT UNIQUE NOT NULL,
		password_hash TEXT NOT NULL,
		email_verified BOOLEAN NOT NULL DEFAULT FALSE,
		created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
	);
	
	CREATE TABLE IF NOT EXISTS refresh_tokens (
		id UUID PRIMARY KEY,
		user_id UUID REFERENCES users(id) ON DELETE CASCADE,
		token_hash TEXT UNIQUE NOT NULL,
		expires_at TIMESTAMPTZ NOT NULL,
		created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		revoked_at TIMESTAMPTZ
	);

	CREATE TABLE IF NOT EXISTS email_verifications (
		id UUID PRIMARY KEY,
		user_id UUID REFERENCES users(id) ON DELETE CASCADE,
		code_hash TEXT NOT NULL,
		expires_at TIMESTAMPTZ NOT NULL,
		attempts INT NOT NULL DEFAULT 0,
		used_at TIMESTAMPTZ,
		created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
	);
	`
	_, err := conn.Exec(schema)
	return err
}
