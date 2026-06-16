package model

import (
	"time"

	"github.com/google/uuid"
)

type BlackjackSession struct {
	ID           uuid.UUID `db:"id" json:"id"`
	SessionID    string    `db:"session_id" json:"session_id"`
	UserID       uuid.UUID `db:"user_id" json:"user_id"`
	BetAmount    int64     `db:"bet_amount" json:"bet_amount"`
	PlayerCards  []byte    `db:"player_cards" json:"player_cards"`
	DealerCards  []byte    `db:"dealer_cards" json:"dealer_cards"`
	PlayerValue  int       `db:"player_value" json:"player_value"`
	DealerValue  int       `db:"dealer_value" json:"dealer_value"`
	IsPlayerBust bool      `db:"is_player_bust" json:"is_player_bust"`
	IsBlackjack  bool      `db:"is_blackjack" json:"is_blackjack"`
	GameStatus   string    `db:"game_status" json:"game_status"`
	CreatedAt    time.Time `db:"created_at" json:"created_at"`
	UpdatedAt    time.Time `db:"updated_at" json:"updated_at"`
}

const (
	BlackjackStatusPlaying   = "PLAYING"
	BlackjackStatusPlayerWin = "PLAYER_WIN"
	BlackjackStatusDealerWin = "DEALER_WIN"
	BlackjackStatusPush      = "PUSH"
)
