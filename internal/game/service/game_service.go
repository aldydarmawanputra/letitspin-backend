package service

import (
	"context"
	"encoding/json"
	"fmt"

	"let-it-spin/internal/game/base"
	"let-it-spin/internal/game/blackjack"
	gameRepository "let-it-spin/internal/game/repository"
	"let-it-spin/internal/game/slot"
	"let-it-spin/internal/model"
	walletRepository "let-it-spin/internal/wallet/repository"
	walletService "let-it-spin/internal/wallet/service"

	"github.com/google/uuid"
)

type GameService struct {
	gameRepo   *gameRepository.GameRepository
	walletRepo *walletRepository.WalletRepository
	walletSvc  *walletService.WalletService
	engines    map[string]base.GameEngine
}

func NewGameService(gameRepo *gameRepository.GameRepository, walletRepo *walletRepository.WalletRepository, walletSvc *walletService.WalletService) *GameService {
	return &GameService{
		gameRepo:   gameRepo,
		walletRepo: walletRepo,
		walletSvc:  walletSvc,
		engines:    make(map[string]base.GameEngine),
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

	gameResult, err := engine.Execute(ctx, userID, betAmount, options)
	if err != nil {
		return nil, err
	}

	resultData, err := json.Marshal(gameResult.Details)
	if err != nil {
		return nil, err
	}

	balanceBefore := wallet.Balance
	balanceAfter := wallet.Balance - betAmount + gameResult.WinAmount

	sessionID := gameResult.SessionID
	if sessionID == uuid.Nil {
		sessionID = uuid.New()
	}

	// Cek apakah game sudah selesai (bukan PLAYING)
	isFinal := true
	if gameCode == "BLACKJACK" {
		if details, ok := gameResult.Details.(*blackjack.GameState); ok {
			if details.GameStatus == blackjack.StatusPlaying {
				isFinal = false
			}
		}
	}

	if isFinal {
		tx, err := s.gameRepo.BeginTx(ctx)
		if err != nil {
			return nil, err
		}
		defer s.gameRepo.RollbackTx(tx)

		session := &model.GameSession{
			ID:            uuid.New(),
			UserID:        userID,
			GameTypeID:    gameType.ID,
			BetAmount:     betAmount,
			WinAmount:     gameResult.WinAmount,
			ResultData:    resultData,
			BalanceBefore: balanceBefore,
			BalanceAfter:  balanceAfter,
			Status:        model.GameSessionStatusCompleted,
		}

		if gameResult.WinAmount > 0 {
			session.Result = model.GameResultWin
		} else {
			session.Result = model.GameResultLose
		}

		err = s.gameRepo.CreateGameSession(ctx, tx, session)
		if err != nil {
			return nil, err
		}

		err = s.walletRepo.UpdateBalance(ctx, tx, wallet.ID, balanceAfter)
		if err != nil {
			return nil, err
		}

		if err := s.gameRepo.CommitTx(tx); err != nil {
			return nil, err
		}
	}

	return &base.GameSessionResult{
		SessionID: sessionID,
		GameCode:  gameCode,
		BetAmount: betAmount,
		WinAmount: gameResult.WinAmount,
		IsWin:     gameResult.IsWin,
		Details:   gameResult.Details,
		Balance:   balanceAfter,
	}, nil
}

func (s *GameService) GetHistory(ctx context.Context, userID uuid.UUID, gameCode *string, result *string, page, limit int) ([]model.GameSession, int, error) {
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

	sessions, err := s.gameRepo.GetGameSessions(ctx, userID, gameTypeID, result, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	return sessions, len(sessions), nil
}
