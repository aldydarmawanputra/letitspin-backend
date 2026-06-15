package dice

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"

	"let-it-spin/internal/game/base"
	"let-it-spin/internal/game/repository"
	"let-it-spin/internal/model"

	"github.com/google/uuid"
)

type DiceEngine struct {
	gameCode      string
	configRepo    *repository.ConfigRepository
	runtimeConfig *model.GameRuntimeConfig
}

func NewDiceEngine(configRepo *repository.ConfigRepository) *DiceEngine {
	return &DiceEngine{
		gameCode:   "DICE",
		configRepo: configRepo,
	}
}

func (e *DiceEngine) loadConfig(ctx context.Context) error {
	config, err := e.configRepo.GetGameConfig(ctx, e.gameCode)
	if err != nil {
		return err
	}
	e.runtimeConfig = config
	return nil
}

func (e *DiceEngine) GetGameCode() string {
	return e.gameCode
}

func (e *DiceEngine) GetMinBet() int64 {
	if e.runtimeConfig != nil {
		return e.runtimeConfig.MinBet
	}
	return 100
}

func (e *DiceEngine) GetMaxBet() int64 {
	if e.runtimeConfig != nil {
		return e.runtimeConfig.MaxBet
	}
	return 100000
}

func (e *DiceEngine) ValidateBet(amount int64) error {
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

func (e *DiceEngine) Execute(ctx context.Context, userID uuid.UUID, betAmount int64, options map[string]interface{}) (*base.GameResult, error) {
	if err := e.loadConfig(ctx); err != nil {
		return nil, err
	}

	prediction, ok := options["prediction"].(string)
	if !ok {
		return nil, fmt.Errorf("prediction is required: UNDER, OVER, or EXACT")
	}

	rollNumber := e.rollDice()
	betValue := e.getBetValue(options)

	var isWin bool
	var multiplier int
	var winAmount int64

	switch prediction {
	case "UNDER":
		if rollNumber >= 1 && rollNumber <= 49 {
			isWin = true
			multiplier = Multipliers["UNDER"]
		}
	case "OVER":
		if rollNumber >= 51 && rollNumber <= 99 {
			isWin = true
			multiplier = Multipliers["OVER"]
		}
	case "EXACT":
		if betValue != nil && rollNumber == *betValue {
			if rollNumber == 1 || rollNumber == 100 {
				multiplier = Multipliers["EDGE"]
			} else {
				multiplier = Multipliers["EXACT"]
			}
			isWin = true
		}
	default:
		return nil, fmt.Errorf("invalid prediction: %s", prediction)
	}

	if isWin {
		winAmount = betAmount * int64(multiplier)
	}

	result := &DiceResult{
		RollNumber: rollNumber,
		Prediction: prediction,
		BetValue:   betValue,
		WinAmount:  winAmount,
		Multiplier: multiplier,
	}

	return &base.GameResult{
		WinAmount:  winAmount,
		IsWin:      isWin,
		Details:    result,
		Multiplier: float64(winAmount) / float64(betAmount),
	}, nil
}

func (e *DiceEngine) rollDice() int {
	n, _ := rand.Int(rand.Reader, big.NewInt(100))
	return int(n.Int64()) + 1
}

func (e *DiceEngine) getBetValue(options map[string]interface{}) *int {
	if val, ok := options["bet_value"]; ok {
		if floatVal, ok := val.(float64); ok {
			intVal := int(floatVal)
			return &intVal
		}
	}
	return nil
}
