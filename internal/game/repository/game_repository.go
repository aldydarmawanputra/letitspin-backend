package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"

	"let-it-spin/internal/model"

	"github.com/google/uuid"
)

type GameRepository struct {
	db *sql.DB
}

func NewGameRepository(db *sql.DB) *GameRepository {
	return &GameRepository{db: db}
}

func (r *GameRepository) GetGameTypeByCode(ctx context.Context, code string) (*model.GameType, error) {
	query := `
        SELECT id, code, name, min_bet, max_bet, config, is_active, created_at, updated_at
        FROM game_types
        WHERE code = $1 AND is_active = true
    `

	var gameType model.GameType
	err := r.db.QueryRowContext(ctx, query, code).Scan(
		&gameType.ID,
		&gameType.Code,
		&gameType.Name,
		&gameType.MinBet,
		&gameType.MaxBet,
		&gameType.Config,
		&gameType.IsActive,
		&gameType.CreatedAt,
		&gameType.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("game type not found")
		}
		return nil, fmt.Errorf("failed to get game type: %w", err)
	}

	return &gameType, nil
}

func (r *GameRepository) CreateGameSession(ctx context.Context, tx *sql.Tx, session *model.GameSession) error {
	query := `
        INSERT INTO game_sessions (
            id, user_id, game_type_id, bet_amount, win_amount, 
            result_data, balance_before, balance_after, result, status
        )
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
    `

	_, err := tx.ExecContext(ctx, query,
		session.ID,
		session.UserID,
		session.GameTypeID,
		session.BetAmount,
		session.WinAmount,
		session.ResultData,
		session.BalanceBefore,
		session.BalanceAfter,
		session.Result,
		session.Status,
	)
	if err != nil {
		return fmt.Errorf("failed to create game session: %w", err)
	}

	return nil
}

func (r *GameRepository) GetGameSessions(ctx context.Context, userID uuid.UUID, gameTypeID *uuid.UUID, result *string, limit, offset int) ([]model.GameSession, error) {
	var query string
	var args []interface{}

	if gameTypeID != nil && result != nil {
		query = `
            SELECT id, user_id, game_type_id, bet_amount, win_amount, result_data, balance_before, balance_after, result, status, created_at
            FROM game_sessions
            WHERE user_id = $1 AND game_type_id = $2 AND result = $3
            ORDER BY created_at DESC
            LIMIT $4 OFFSET $5
        `
		args = []interface{}{userID, gameTypeID, *result, limit, offset}
	} else if gameTypeID != nil {
		query = `
            SELECT id, user_id, game_type_id, bet_amount, win_amount, result_data, balance_before, balance_after, result, status, created_at
            FROM game_sessions
            WHERE user_id = $1 AND game_type_id = $2
            ORDER BY created_at DESC
            LIMIT $3 OFFSET $4
        `
		args = []interface{}{userID, gameTypeID, limit, offset}
	} else if result != nil {
		query = `
            SELECT id, user_id, game_type_id, bet_amount, win_amount, result_data, balance_before, balance_after, result, status, created_at
            FROM game_sessions
            WHERE user_id = $1 AND result = $2
            ORDER BY created_at DESC
            LIMIT $3 OFFSET $4
        `
		args = []interface{}{userID, *result, limit, offset}
	} else {
		query = `
            SELECT id, user_id, game_type_id, bet_amount, win_amount, result_data, balance_before, balance_after, result, status, created_at
            FROM game_sessions
            WHERE user_id = $1
            ORDER BY created_at DESC
            LIMIT $2 OFFSET $3
        `
		args = []interface{}{userID, limit, offset}
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get game sessions: %w", err)
	}
	defer rows.Close()

	var sessions []model.GameSession
	for rows.Next() {
		var s model.GameSession
		err := rows.Scan(
			&s.ID,
			&s.UserID,
			&s.GameTypeID,
			&s.BetAmount,
			&s.WinAmount,
			&s.ResultData,
			&s.BalanceBefore,
			&s.BalanceAfter,
			&s.Result,
			&s.Status,
			&s.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan game session: %w", err)
		}
		sessions = append(sessions, s)
	}

	return sessions, nil
}

func (r *GameRepository) BeginTx(ctx context.Context) (*sql.Tx, error) {
	return r.db.BeginTx(ctx, nil)
}

func (r *GameRepository) CommitTx(tx *sql.Tx) error {
	return tx.Commit()
}

func (r *GameRepository) RollbackTx(tx *sql.Tx) error {
	return tx.Rollback()
}

func (r *GameRepository) GetGameConfig(ctx context.Context, gameCode string) (*model.GameRuntimeConfig, error) {
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

func (r *GameRepository) UpdateGameConfig(ctx context.Context, gameCode string, configKey string, configValue string) error {
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
