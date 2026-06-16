package base

import (
	"context"

	"github.com/google/uuid"
)

type GameEngine interface {
	GetGameCode() string
	GetMinBet() int64
	GetMaxBet() int64
	ValidateBet(amount int64) error
	Execute(ctx context.Context, userID uuid.UUID, betAmount int64, options map[string]interface{}) (*GameResult, error)
}

type GameResult struct {
	SessionID  uuid.UUID   `json:"session_id,omitempty"`
	WinAmount  int64       `json:"win_amount"`
	IsWin      bool        `json:"is_win"`
	Details    interface{} `json:"details"`
	Multiplier float64     `json:"multiplier"`
}

type GameSessionResult struct {
	SessionID uuid.UUID   `json:"session_id"`
	GameCode  string      `json:"game_code"`
	BetAmount int64       `json:"bet_amount"`
	WinAmount int64       `json:"win_amount"`
	IsWin     bool        `json:"is_win"`
	Details   interface{} `json:"details"`
	Balance   int64       `json:"balance"`
}
