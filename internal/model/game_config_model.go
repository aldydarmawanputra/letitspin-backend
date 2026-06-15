package model

import (
	"github.com/google/uuid"
	"time"
)

type GameConfig struct {
	ID          uuid.UUID `db:"id" json:"id"`
	GameCode    string    `db:"game_code" json:"game_code"`
	ConfigKey   string    `db:"config_key" json:"config_key"`
	ConfigValue string    `db:"config_value" json:"config_value"`
	ValueType   string    `db:"value_type" json:"value_type"`
	Description string    `db:"description" json:"description"`
	IsActive    bool      `db:"is_active" json:"is_active"`
	CreatedAt   time.Time `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time `db:"updated_at" json:"updated_at"`
}

type GameRuntimeConfig struct {
	RTP              int     `json:"rtp"`
	MinBet           int64   `json:"min_bet"`
	MaxBet           int64   `json:"max_bet"`
	MaxWinMultiplier int     `json:"max_win_multiplier"`
	HouseEdge        int     `json:"house_edge"`
	JackpotChance    float64 `json:"jackpot_chance"`
}
