package model

import (
	"time"

	"github.com/google/uuid"
)

type Wallet struct {
	ID        uuid.UUID `db:"id" json:"id"`
	UserID    uuid.UUID `db:"user_id" json:"user_id"`
	Balance   int64     `db:"balance" json:"balance"`
	Currency  string    `db:"currency" json:"currency"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

type Transaction struct {
	ID            uuid.UUID `db:"id" json:"id"`
	UserID        uuid.UUID `db:"user_id" json:"user_id"`
	WalletID      uuid.UUID `db:"wallet_id" json:"wallet_id"`
	Amount        int64     `db:"amount" json:"amount"`
	Type          string    `db:"type" json:"type"`
	ReferenceID   *string   `db:"reference_id" json:"reference_id,omitempty"`
	Description   *string   `db:"description" json:"description,omitempty"`
	BalanceBefore int64     `db:"balance_before" json:"balance_before"`
	BalanceAfter  int64     `db:"balance_after" json:"balance_after"`
	CreatedAt     time.Time `db:"created_at" json:"created_at"`
}

const (
	TransactionTypeDeposit  = "deposit"
	TransactionTypeWithdraw = "withdraw"
	TransactionTypeBet      = "bet"
	TransactionTypeWin      = "win"
)
