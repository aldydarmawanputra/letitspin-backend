package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type TokenRepository struct {
	db *sql.DB
}

func NewTokenRepository(db *sql.DB) *TokenRepository {
	return &TokenRepository{db: db}
}

func (r *TokenRepository) StoreRefreshToken(ctx context.Context, userID uuid.UUID, token string, expiresAt time.Time) error {
	query := `
		INSERT INTO refresh_tokens (
			id,
			user_id,
			token,
			expires_at,
			revoked,
			created_at
		)
		VALUES ($1,$2,$3,$4,false,NOW())
	`

	_, err := r.db.ExecContext(
		ctx,
		query,
		uuid.New(),
		userID,
		token,
		expiresAt,
	)

	if err != nil {
		return fmt.Errorf("StoreRefreshToken failed: %w", err)
	}

	return nil
}

func (r *TokenRepository) ValidateRefreshToken(ctx context.Context, token string) (uuid.UUID, error) {
	var userID uuid.UUID

	query := `
		SELECT user_id
		FROM refresh_tokens
		WHERE token = $1
			AND revoked = false
			AND expires_at > NOW()
		LIMIT 1
	`

	err := r.db.QueryRowContext(ctx, query, token).Scan(&userID)
	if err != nil {
		return uuid.Nil, fmt.Errorf("ValidateRefreshToken failed: %w", err)
	}

	return userID, nil
}

func (r *TokenRepository) RevokeRefreshToken(ctx context.Context, token string) error {
	query := `
		UPDATE refresh_tokens
		SET revoked = true
		WHERE token = $1
	`

	_, err := r.db.ExecContext(ctx, query, token)
	if err != nil {
		return fmt.Errorf("RevokeRefreshToken failed: %w", err)
	}

	return nil
}

func (r *TokenRepository) RevokeAllUserTokens(ctx context.Context, userID uuid.UUID) error {
	query := `
		UPDATE refresh_tokens
		SET revoked = true
		WHERE user_id = $1
	`

	_, err := r.db.ExecContext(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("RevokeAllUserTokens failed: %w", err)
	}

	return nil
}
