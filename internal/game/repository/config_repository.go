package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"

	"let-it-spin/internal/model"

	"github.com/google/uuid"
)

type ConfigRepository struct {
	db *sql.DB
}

func NewConfigRepository(db *sql.DB) *ConfigRepository {
	return &ConfigRepository{db: db}
}

func (r *ConfigRepository) GetGameConfig(ctx context.Context, gameCode string) (*model.GameRuntimeConfig, error) {
	query := `
        SELECT config_key, config_value, value_type
        FROM game_configs
        WHERE game_code = $1 AND is_active = true
    `

	rows, err := r.db.QueryContext(ctx, query, gameCode)
	if err != nil {
		return nil, fmt.Errorf("failed to get game config: %w", err)
	}
	defer rows.Close()

	config := &model.GameRuntimeConfig{
		RTP:              96,
		MinBet:           100,
		MaxBet:           100000,
		MaxWinMultiplier: 5000,
		HouseEdge:        4,
		JackpotChance:    0.1,
	}

	for rows.Next() {
		var key, value, valueType string
		if err := rows.Scan(&key, &value, &valueType); err != nil {
			continue
		}

		switch key {
		case "rtp":
			if val, err := strconv.Atoi(value); err == nil {
				config.RTP = val
			}
		case "min_bet":
			if val, err := strconv.ParseInt(value, 10, 64); err == nil {
				config.MinBet = val
			}
		case "max_bet":
			if val, err := strconv.ParseInt(value, 10, 64); err == nil {
				config.MaxBet = val
			}
		case "max_win_multiplier":
			if val, err := strconv.Atoi(value); err == nil {
				config.MaxWinMultiplier = val
			}
		case "house_edge":
			if val, err := strconv.Atoi(value); err == nil {
				config.HouseEdge = val
			}
		case "jackpot_chance":
			if val, err := strconv.ParseFloat(value, 64); err == nil {
				config.JackpotChance = val
			}
		}
	}

	return config, nil
}

func (r *ConfigRepository) UpdateGameConfig(ctx context.Context, gameCode string, configKey string, configValue string) error {
	query := `
        INSERT INTO game_configs (id, game_code, config_key, config_value, value_type)
        VALUES ($1, $2, $3, $4, 'string')
        ON CONFLICT (game_code, config_key) 
        DO UPDATE SET config_value = $4, updated_at = NOW()
    `

	_, err := r.db.ExecContext(ctx, query, uuid.New(), gameCode, configKey, configValue)
	if err != nil {
		return fmt.Errorf("failed to update game config: %w", err)
	}

	return nil
}
