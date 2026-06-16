package roulette

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"strings"

	"let-it-spin/internal/game/base"
	"let-it-spin/internal/game/repository"
	"let-it-spin/internal/model"

	"github.com/google/uuid"
)

type RouletteEngine struct {
	gameCode      string
	configRepo    *repository.ConfigRepository
	runtimeConfig *model.GameRuntimeConfig
}

func NewRouletteEngine(configRepo *repository.ConfigRepository) *RouletteEngine {
	return &RouletteEngine{
		gameCode:   "ROULETTE",
		configRepo: configRepo,
	}
}

func (e *RouletteEngine) loadConfig(ctx context.Context) error {
	config, err := e.configRepo.GetGameConfig(ctx, e.gameCode)
	if err != nil {
		return err
	}
	e.runtimeConfig = config
	return nil
}

func (e *RouletteEngine) GetGameCode() string {
	return e.gameCode
}

func (e *RouletteEngine) GetMinBet() int64 {
	if e.runtimeConfig != nil {
		return e.runtimeConfig.MinBet
	}
	return 100
}

func (e *RouletteEngine) GetMaxBet() int64 {
	if e.runtimeConfig != nil {
		return e.runtimeConfig.MaxBet
	}
	return 100000
}

func (e *RouletteEngine) ValidateBet(amount int64) error {
	minBet := e.GetMinBet()
	maxBet := e.GetMaxBet()

	if amount < minBet {
		return fmt.Errorf("bet amount must be at least %d", minBet)
	}
	if amount > maxBet {
		return fmt.Errorf("bet amount cannot exceed %d", maxBet)
	}
	return nil
}

func (e *RouletteEngine) Execute(ctx context.Context, userID uuid.UUID, betAmount int64, options map[string]interface{}) (*base.GameResult, error) {
	if err := e.loadConfig(ctx); err != nil {
		return nil, err
	}

	betType, ok := options["bet_type"].(string)
	if !ok {
		return nil, fmt.Errorf("bet_type is required: RED, BLACK, GREEN, ODD, EVEN, or NUMBER")
	}

	result := e.spin()
	isWin, multiplier := e.checkWin(result, betType, options)

	var winAmount int64
	if isWin {
		winAmount = betAmount * int64(multiplier)
	}

	details := &RouletteResult{
		Number:     result.Number,
		Color:      result.Color,
		Parity:     result.Parity,
		BetType:    betType,
		BetValue:   e.getBetValue(options),
		WinAmount:  winAmount,
		Multiplier: multiplier,
	}

	return &base.GameResult{
		WinAmount:  winAmount,
		IsWin:      isWin,
		Details:    details,
		Multiplier: float64(winAmount) / float64(betAmount),
	}, nil
}

func (e *RouletteEngine) spin() RouletteNumber {
	n, _ := rand.Int(rand.Reader, big.NewInt(37))
	index := int(n.Int64())
	return RouletteWheel[index]
}

func (e *RouletteEngine) checkWin(result RouletteNumber, betType string, options map[string]interface{}) (bool, int) {
	betTypeUpper := strings.ToUpper(betType)

	switch BetType(betTypeUpper) {
	case BetTypeRed:
		return result.Color == "RED", Multipliers[BetTypeRed]
	case BetTypeBlack:
		return result.Color == "BLACK", Multipliers[BetTypeBlack]
	case BetTypeGreen:
		return result.Number == 0, Multipliers[BetTypeGreen]
	case BetTypeOdd:
		return result.Parity == "ODD", Multipliers[BetTypeOdd]
	case BetTypeEven:
		return result.Parity == "EVEN", Multipliers[BetTypeEven]
	case BetTypeNumber:
		if val, ok := options["bet_value"]; ok {
			if floatVal, ok := val.(float64); ok {
				return result.Number == int(floatVal), Multipliers[BetTypeNumber]
			}
		}
		return false, 0
	default:
		return false, 0
	}
}

func (e *RouletteEngine) getBetValue(options map[string]interface{}) string {
	if val, ok := options["bet_value"]; ok {
		if floatVal, ok := val.(float64); ok {
			return fmt.Sprintf("%d", int(floatVal))
		}
		if strVal, ok := val.(string); ok {
			return strVal
		}
	}
	return ""
}
