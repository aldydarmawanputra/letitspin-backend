package service

import (
	"context"
	"encoding/json"
	"fmt"

	"let-it-spin/internal/game/base"
	"let-it-spin/internal/game/repository"
	"let-it-spin/internal/game/slot"
	"let-it-spin/internal/model"
	"let-it-spin/internal/wallet/service"

	"github.com/google/uuid"
)

type GameService struct {
	gameRepo  *repository.GameRepository
	walletSvc *service.WalletService
	engines   map[string]base.GameEngine
}

func NewGameService(gameRepo *repository.GameRepository, walletSvc *service.WalletService) *GameService {
	return &GameService{
		gameRepo:  gameRepo,
		walletSvc: walletSvc,
		engines:   make(map[string]base.GameEngine),
	}
}

func (s *GameService) RegisterEngine(engine base.GameEngine) {
	s.engines[engine.GetGameCode()] = engine
}

func (s *GameService) GetGameConfig(ctx context.Context, gameCode string) (map[string]interface{}, error) {
	gameType, err := s.gameRepo.GetGameTypeByCode(ctx, gameCode)
	if err != nil {
		return nil, err
	}

	engine, exists := s.engines[gameCode]
	if !exists {
		return nil, fmt.Errorf("game engine not found")
	}

	result := map[string]interface{}{
		"id":        gameType.ID,
		"code":      gameType.Code,
		"name":      gameType.Name,
		"min_bet":   engine.GetMinBet(),
		"max_bet":   engine.GetMaxBet(),
		"is_active": gameType.IsActive,
	}

	if slotEngine, ok := engine.(*slot.SlotEngine); ok {
		if runtimeConfig := slotEngine.GetRuntimeConfig(); runtimeConfig != nil {
			result["rtp"] = runtimeConfig.RTP
			result["house_edge"] = runtimeConfig.HouseEdge
			result["max_win_multiplier"] = runtimeConfig.MaxWinMultiplier
		}
	}

	return result, nil
}

func (s *GameService) Play(ctx context.Context, userID uuid.UUID, gameCode string, betAmount int64, options map[string]interface{}) (*base.GameSessionResult, error) {
	engine, exists := s.engines[gameCode]
	if !exists {
		return nil, fmt.Errorf("game engine not found for code: %s", gameCode)
	}

	gameType, err := s.gameRepo.GetGameTypeByCode(ctx, gameCode)
	if err != nil {
		return nil, err
	}

	if err := engine.ValidateBet(betAmount); err != nil {
		return nil, err
	}

	if betAmount < gameType.MinBet || betAmount > gameType.MaxBet {
		return nil, fmt.Errorf("bet amount must be between %d and %d", gameType.MinBet, gameType.MaxBet)
	}

	wallet, err := s.walletSvc.GetBalance(ctx, userID)
	if err != nil {
		return nil, err
	}

	if wallet.Balance < betAmount {
		return nil, fmt.Errorf("insufficient balance")
	}

	_, _, err = s.walletSvc.Withdraw(ctx, service.WithdrawRequest{
		UserID:      userID,
		Amount:      betAmount,
		ReferenceID: nil,
		Description: &[]string{fmt.Sprintf("bet on %s", gameCode)}[0],
	})
	if err != nil {
		return nil, err
	}

	gameResult, err := engine.Execute(ctx, userID, betAmount, options)
	if err != nil {
		return nil, err
	}

	resultData, err := json.Marshal(gameResult.Details)
	if err != nil {
		return nil, err
	}

	tx, err := s.gameRepo.BeginTx(ctx)
	if err != nil {
		return nil, err
	}
	defer s.gameRepo.RollbackTx(tx)

	session := &model.GameSession{
		ID:         uuid.New(),
		UserID:     userID,
		GameTypeID: gameType.ID,
		BetAmount:  betAmount,
		WinAmount:  gameResult.WinAmount,
		ResultData: resultData,
		Status:     model.GameSessionStatusCompleted,
	}

	err = s.gameRepo.CreateGameSession(ctx, tx, session)
	if err != nil {
		return nil, err
	}

	finalBalance := wallet.Balance - betAmount

	if gameResult.WinAmount > 0 {
		_, _, err = s.walletSvc.Deposit(ctx, service.DepositRequest{
			UserID:      userID,
			Amount:      gameResult.WinAmount,
			ReferenceID: &[]string{session.ID.String()}[0],
			Description: &[]string{fmt.Sprintf("win on %s", gameCode)}[0],
		})
		if err != nil {
			return nil, err
		}
		finalBalance = finalBalance + gameResult.WinAmount
	}

	if err := s.gameRepo.CommitTx(tx); err != nil {
		return nil, err
	}

	return &base.GameSessionResult{
		SessionID: session.ID,
		GameCode:  gameCode,
		BetAmount: betAmount,
		WinAmount: gameResult.WinAmount,
		IsWin:     gameResult.WinAmount > 0,
		Details:   gameResult.Details,
		Balance:   finalBalance,
	}, nil
}

func (s *GameService) GetHistory(ctx context.Context, userID uuid.UUID, gameCode *string, page, limit int) ([]model.GameSession, int, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	offset := (page - 1) * limit

	var gameTypeID *uuid.UUID
	if gameCode != nil {
		gameType, err := s.gameRepo.GetGameTypeByCode(ctx, *gameCode)
		if err != nil {
			return nil, 0, err
		}
		gameTypeID = &gameType.ID
	}

	sessions, err := s.gameRepo.GetGameSessions(ctx, userID, gameTypeID, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	return sessions, len(sessions), nil
}
