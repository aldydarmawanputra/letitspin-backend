package repository

import (
	"context"
	"database/sql"
	"fmt"

	"let-it-spin/internal/model"

	"github.com/google/uuid"
)

type WalletRepository struct {
	db *sql.DB
}

func NewWalletRepository(db *sql.DB) *WalletRepository {
	return &WalletRepository{db: db}
}

func (r *WalletRepository) GetOrCreateWallet(ctx context.Context, userID uuid.UUID) (*model.Wallet, error) {
	query := `
		INSERT INTO wallets (user_id, balance, currency)
		VALUES ($1, 0, 'IDR')
		ON CONFLICT (user_id) DO UPDATE SET user_id = EXCLUDED.user_id
		RETURNING id, user_id, balance, currency, created_at, updated_at
	`

	var wallet model.Wallet
	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&wallet.ID,
		&wallet.UserID,
		&wallet.Balance,
		&wallet.Currency,
		&wallet.CreatedAt,
		&wallet.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get or create wallet: %w", err)
	}

	return &wallet, nil
}

func (r *WalletRepository) GetWalletByUserID(ctx context.Context, userID uuid.UUID) (*model.Wallet, error) {
	query := `
		SELECT id, user_id, balance, currency, created_at, updated_at
		FROM wallets
		WHERE user_id = $1
	`

	var wallet model.Wallet
	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&wallet.ID,
		&wallet.UserID,
		&wallet.Balance,
		&wallet.Currency,
		&wallet.CreatedAt,
		&wallet.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("wallet not found")
		}
		return nil, fmt.Errorf("failed to get wallet: %w", err)
	}

	return &wallet, nil
}

func (r *WalletRepository) UpdateBalance(ctx context.Context, tx *sql.Tx, walletID uuid.UUID, newBalance int64) error {
	query := `
		UPDATE wallets
		SET balance = $2, updated_at = NOW()
		WHERE id = $1
	`

	_, err := tx.ExecContext(ctx, query, walletID, newBalance)
	if err != nil {
		return fmt.Errorf("failed to update balance: %w", err)
	}

	return nil
}

func (r *WalletRepository) UpdateBalanceDirect(ctx context.Context, tx *sql.Tx, walletID uuid.UUID, newBalance int64) error {
	query := `
        UPDATE wallets
        SET balance = $2, updated_at = NOW()
        WHERE id = $1
    `

	_, err := tx.ExecContext(ctx, query, walletID, newBalance)
	if err != nil {
		return fmt.Errorf("failed to update balance: %w", err)
	}

	return nil
}

func (r *WalletRepository) CreateTransaction(ctx context.Context, tx *sql.Tx, transaction *model.Transaction) error {
	query := `
		INSERT INTO transactions (
			id, user_id, wallet_id, amount, type, reference_id,
			description, balance_before, balance_after
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	_, err := tx.ExecContext(ctx, query,
		transaction.ID,
		transaction.UserID,
		transaction.WalletID,
		transaction.Amount,
		transaction.Type,
		transaction.ReferenceID,
		transaction.Description,
		transaction.BalanceBefore,
		transaction.BalanceAfter,
	)
	if err != nil {
		return fmt.Errorf("failed to create transaction: %w", err)
	}

	return nil
}

func (r *WalletRepository) GetTransactions(ctx context.Context, userID uuid.UUID, limit, offset int) ([]model.Transaction, error) {
	query := `
		SELECT id, user_id, wallet_id, amount, type, reference_id,
			description, balance_before, balance_after, created_at
		FROM transactions
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get transactions: %w", err)
	}
	defer rows.Close()

	var transactions []model.Transaction
	for rows.Next() {
		var t model.Transaction
		err := rows.Scan(
			&t.ID,
			&t.UserID,
			&t.WalletID,
			&t.Amount,
			&t.Type,
			&t.ReferenceID,
			&t.Description,
			&t.BalanceBefore,
			&t.BalanceAfter,
			&t.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan transaction: %w", err)
		}
		transactions = append(transactions, t)
	}

	return transactions, nil
}

func (r *WalletRepository) BeginTx(ctx context.Context) (*sql.Tx, error) {
	return r.db.BeginTx(ctx, nil)
}

func (r *WalletRepository) CommitTx(tx *sql.Tx) error {
	return tx.Commit()
}

func (r *WalletRepository) RollbackTx(tx *sql.Tx) error {
	return tx.Rollback()
}
