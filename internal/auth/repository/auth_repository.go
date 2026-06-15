package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
)

type AuthRepository struct {
	db *sql.DB
}

func NewAuthRepository(db *sql.DB) *AuthRepository {
	return &AuthRepository{db: db}
}

func (r *AuthRepository) GetUserCredentialByUserID(ctx context.Context, userID uuid.UUID) (string, error) {
	var hash string

	query := `
		SELECT password
		FROM user_credentials
		WHERE user_id = $1
		LIMIT 1
	`

	err := r.db.QueryRowContext(ctx, query, userID).Scan(&hash)
	if err != nil {
		return "", fmt.Errorf("GetUserCredentialByUserID failed: %w", err)
	}

	return hash, nil
}

func (r *AuthRepository) StoreRefreshToken(ctx context.Context, userID uuid.UUID, token string, expiresAt any) error {
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

func (r *AuthRepository) ValidateRefreshToken(ctx context.Context, token string) (uuid.UUID, error) {
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

func (r *AuthRepository) RevokeRefreshToken(ctx context.Context, token string) error {
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

func (r *AuthRepository) RevokeAllUserTokens(ctx context.Context, userID uuid.UUID) error {
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

func (r *AuthRepository) GetUserByEmail(ctx context.Context, email string) (*UserResult, error) {
	var user UserResult

	query := `
		SELECT id, email, is_active, email_verified
		FROM users
		WHERE email = $1 AND deleted_at IS NULL
		LIMIT 1
	`

	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.Email,
		&user.IsActive,
		&user.EmailVerified,
	)

	if err != nil {
		return nil, fmt.Errorf("GetUserByEmail failed: %w", err)
	}

	return &user, nil
}

func (r *AuthRepository) GetUserByID(ctx context.Context, id uuid.UUID) (*UserResult, error) {
	var user UserResult

	query := `
		SELECT id, email, is_active, email_verified
		FROM users
		WHERE id = $1 AND deleted_at IS NULL
		LIMIT 1
	`

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID,
		&user.Email,
		&user.IsActive,
		&user.EmailVerified,
	)

	if err != nil {
		return nil, fmt.Errorf("GetUserByID failed: %w", err)
	}

	return &user, nil
}

type UserResult struct {
	ID            uuid.UUID
	Email         string
	IsActive      bool
	EmailVerified bool
}
