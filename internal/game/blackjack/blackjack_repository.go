package blackjack

import (
	"context"
	"database/sql"
	"fmt"

	"let-it-spin/internal/model"
)

type BlackjackRepository struct {
	db *sql.DB
}

func NewBlackjackRepository(db *sql.DB) *BlackjackRepository {
	return &BlackjackRepository{db: db}
}

func (r *BlackjackRepository) CreateSession(ctx context.Context, session *model.BlackjackSession) error {
	query := `
        INSERT INTO blackjack_sessions (
            id, session_id, user_id, bet_amount, 
            player_cards, dealer_cards, player_value, dealer_value,
            is_player_bust, is_blackjack, game_status
        )
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
    `

	_, err := r.db.ExecContext(ctx, query,
		session.ID,
		session.SessionID,
		session.UserID,
		session.BetAmount,
		session.PlayerCards,
		session.DealerCards,
		session.PlayerValue,
		session.DealerValue,
		session.IsPlayerBust,
		session.IsBlackjack,
		session.GameStatus,
	)
	if err != nil {
		return fmt.Errorf("failed to create blackjack session: %w", err)
	}

	return nil
}

func (r *BlackjackRepository) GetSession(ctx context.Context, sessionID string) (*model.BlackjackSession, error) {
	query := `
        SELECT id, session_id, user_id, bet_amount, 
               player_cards, dealer_cards, player_value, dealer_value,
               is_player_bust, is_blackjack, game_status, created_at, updated_at
        FROM blackjack_sessions
        WHERE session_id = $1 AND game_status = 'PLAYING'
    `

	var session model.BlackjackSession
	err := r.db.QueryRowContext(ctx, query, sessionID).Scan(
		&session.ID,
		&session.SessionID,
		&session.UserID,
		&session.BetAmount,
		&session.PlayerCards,
		&session.DealerCards,
		&session.PlayerValue,
		&session.DealerValue,
		&session.IsPlayerBust,
		&session.IsBlackjack,
		&session.GameStatus,
		&session.CreatedAt,
		&session.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("blackjack session not found")
		}
		return nil, fmt.Errorf("failed to get blackjack session: %w", err)
	}

	return &session, nil
}

func (r *BlackjackRepository) UpdateSession(ctx context.Context, session *model.BlackjackSession) error {
	query := `
        UPDATE blackjack_sessions
        SET player_cards = $2,
            dealer_cards = $3,
            player_value = $4,
            dealer_value = $5,
            is_player_bust = $6,
            is_blackjack = $7,
            game_status = $8,
            updated_at = NOW()
        WHERE session_id = $1
    `

	_, err := r.db.ExecContext(ctx, query,
		session.SessionID,
		session.PlayerCards,
		session.DealerCards,
		session.PlayerValue,
		session.DealerValue,
		session.IsPlayerBust,
		session.IsBlackjack,
		session.GameStatus,
	)
	if err != nil {
		return fmt.Errorf("failed to update blackjack session: %w", err)
	}

	return nil
}

func (r *BlackjackRepository) DeleteSession(ctx context.Context, sessionID string) error {
	query := `DELETE FROM blackjack_sessions WHERE session_id = $1`

	_, err := r.db.ExecContext(ctx, query, sessionID)
	if err != nil {
		return fmt.Errorf("failed to delete blackjack session: %w", err)
	}

	return nil
}
