package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"let-it-spin/internal/model"

	"github.com/google/uuid"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) CreateUser(ctx context.Context, tx *sql.Tx, user *model.User) error {
	if tx == nil {
		return errors.New("transaction is nil")
	}

	query := `
		INSERT INTO users (
			id,
			email,
			is_active,
			email_verified,
			created_at,
			updated_at,
			deleted_at
		)
		VALUES ($1,$2,$3,$4,$5,$6,$7)
	`

	_, err := tx.ExecContext(
		ctx,
		query,
		user.ID,
		user.Email,
		user.IsActive,
		user.EmailVerified,
		user.CreatedAt,
		user.UpdatedAt,
		user.DeletedAt,
	)

	if err != nil {
		return fmt.Errorf("CreateUser failed: %w", err)
	}

	return nil
}

func (r *UserRepository) CreateCredential(ctx context.Context, tx *sql.Tx, userID uuid.UUID, passwordHash string) error {
	if tx == nil {
		return errors.New("transaction is nil")
	}

	query := `
		INSERT INTO user_credentials (
			user_id,
			password,
			created_at,
			updated_at
		)
		VALUES ($1,$2,NOW(),NOW())
	`

	_, err := tx.ExecContext(ctx, query, userID, passwordHash)
	if err != nil {
		return fmt.Errorf("CreateCredential failed: %w", err)
	}

	return nil
}

func (r *UserRepository) BeginTx(ctx context.Context) (*sql.Tx, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("BeginTx failed: %w", err)
	}
	return tx, nil
}

func (r *UserRepository) CommitTx(tx *sql.Tx) error {
	if tx == nil {
		return errors.New("transaction is nil")
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("CommitTx failed: %w", err)
	}
	return nil
}

func (r *UserRepository) RollbackTx(tx *sql.Tx) error {
	if tx == nil {
		return errors.New("transaction is nil")
	}
	if err := tx.Rollback(); err != nil {
		return fmt.Errorf("RollbackTx failed: %w", err)
	}
	return nil
}

func (r *UserRepository) GetUserByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	query := `
		SELECT id, email, is_active, email_verified, created_at, updated_at, deleted_at
		FROM users
		WHERE id = $1
		LIMIT 1
	`

	var user model.User

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID,
		&user.Email,
		&user.IsActive,
		&user.EmailVerified,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.DeletedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("GetUserByID: user not found")
		}
		return nil, fmt.Errorf("GetUserByID failed: %w", err)
	}

	return &user, nil
}

func (r *UserRepository) GetUserByEmail(ctx context.Context, email string) (*model.User, error) {
	query := `
		SELECT id, email, is_active, email_verified, created_at, updated_at, deleted_at
		FROM users
		WHERE email = $1
		LIMIT 1
	`

	var user model.User

	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.Email,
		&user.IsActive,
		&user.EmailVerified,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.DeletedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("GetUserByEmail: user not found")
		}
		return nil, fmt.Errorf("GetUserByEmail failed: %w", err)
	}

	return &user, nil
}

func (r *UserRepository) EmailExists(ctx context.Context, email string) (bool, error) {
	var exists bool

	query := `
		SELECT EXISTS (
			SELECT 1 FROM users WHERE email = $1
		)
	`

	err := r.db.QueryRowContext(ctx, query, email).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("EmailExists failed: %w", err)
	}

	return exists, nil
}

func (r *UserRepository) UpdateUser(ctx context.Context, user *model.User) error {
	query := `
		UPDATE users
		SET email = $2,
			is_active = $3,
			email_verified = $4,
			updated_at = $5
		WHERE id = $1
	`

	res, err := r.db.ExecContext(
		ctx,
		query,
		user.ID,
		user.Email,
		user.IsActive,
		user.EmailVerified,
		user.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("UpdateUser failed: %w", err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("UpdateUser RowsAffected failed: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("UpdateUser: no rows updated")
	}

	return nil
}

func (r *UserRepository) PatchUser(ctx context.Context, id uuid.UUID, email *string, isActive *bool, emailVerified *bool) error {
	query := `
		UPDATE users
		SET
			email = COALESCE($2, email),
			is_active = COALESCE($3, is_active),
			email_verified = COALESCE($4, email_verified),
			updated_at = NOW()
		WHERE id = $1
	`

	res, err := r.db.ExecContext(ctx, query, id, email, isActive, emailVerified)
	if err != nil {
		return fmt.Errorf("PatchUser failed: %w", err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("PatchUser RowsAffected failed: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("PatchUser: no rows updated")
	}

	return nil
}

func (r *UserRepository) DeleteUser(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE users
		SET deleted_at = NOW()
		WHERE id = $1
	`

	res, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("DeleteUser failed: %w", err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("DeleteUser RowsAffected failed: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("DeleteUser: no rows updated")
	}

	return nil
}
