package model

import (
	"time"

	"github.com/google/uuid"
)

type GameType struct {
	ID        uuid.UUID `db:"id" json:"id"`
	Code      string    `db:"code" json:"code"`
	Name      string    `db:"name" json:"name"`
	MinBet    int64     `db:"min_bet" json:"min_bet"`
	MaxBet    int64     `db:"max_bet" json:"max_bet"`
	Config    []byte    `db:"config" json:"config,omitempty"`
	IsActive  bool      `db:"is_active" json:"is_active"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

type GameSession struct {
	ID         uuid.UUID `db:"id" json:"id"`
	UserID     uuid.UUID `db:"user_id" json:"user_id"`
	GameTypeID uuid.UUID `db:"game_type_id" json:"game_type_id"`
	BetAmount  int64     `db:"bet_amount" json:"bet_amount"`
	WinAmount  int64     `db:"win_amount" json:"win_amount"`
	ResultData []byte    `db:"result_data" json:"result_data"`
	Status     string    `db:"status" json:"status"`
	CreatedAt  time.Time `db:"created_at" json:"created_at"`
}

const (
	GameSessionStatusPending   = "PENDING"
	GameSessionStatusCompleted = "COMPLETED"
	GameSessionStatusFailed    = "FAILED"
)
