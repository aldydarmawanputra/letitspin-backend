package slot

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

type SlotEngine struct {
	gameCode      string
	configRepo    *repository.ConfigRepository
	runtimeConfig *model.GameRuntimeConfig
}

func NewSlotEngine(configRepo *repository.ConfigRepository) *SlotEngine {
	return &SlotEngine{
		gameCode:   "SLOT_CLASSIC",
		configRepo: configRepo,
	}
}

func (e *SlotEngine) loadConfig(ctx context.Context) error {
	config, err := e.configRepo.GetGameConfig(ctx, e.gameCode)
	if err != nil {
		return err
	}
	e.runtimeConfig = config
	return nil
}

func (e *SlotEngine) GetGameCode() string {
	return e.gameCode
}

func (e *SlotEngine) GetMinBet() int64 {
	if e.runtimeConfig != nil {
		return e.runtimeConfig.MinBet
	}
	return 100
}

func (e *SlotEngine) GetMaxBet() int64 {
	if e.runtimeConfig != nil {
		return e.runtimeConfig.MaxBet
	}
	return 100000
}

func (e *SlotEngine) ValidateBet(amount int64) error {
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

func (e *SlotEngine) Execute(ctx context.Context, userID uuid.UUID, betAmount int64, options map[string]interface{}) (*base.GameResult, error) {
	if err := e.loadConfig(ctx); err != nil {
		return nil, err
	}

	rtp := e.runtimeConfig.RTP
	randRoll := e.randomInt(1, 100)

	isWin := randRoll <= rtp

	if !isWin {
		result := e.generateLosingSpin()
		return &base.GameResult{
			WinAmount:  0,
			IsWin:      false,
			Details:    result,
			Multiplier: 0,
		}, nil
	}

	result := e.generateWinningSpin(betAmount)

	maxWin := betAmount * int64(e.runtimeConfig.MaxWinMultiplier)
	if result.TotalWin > maxWin {
		result.TotalWin = maxWin
	}

	multiplier := float64(result.TotalWin) / float64(betAmount)

	return &base.GameResult{
		WinAmount:  result.TotalWin,
		IsWin:      result.TotalWin > 0,
		Details:    result,
		Multiplier: multiplier,
	}, nil
}

func (e *SlotEngine) generateLosingSpin() *SpinResult {
	var reels [3][3]string

	for col := 0; col < 3; col++ {
		for row := 0; row < 3; row++ {
			reel := ReelsConfig[col]
			randomIndex := e.randomInt(0, len(reel)-1)
			reels[row][col] = reel[randomIndex]
		}
	}

	reels[0][0] = "CHERRY"
	reels[0][1] = "LEMON"
	reels[0][2] = "ORANGE"

	return &SpinResult{
		Reels:       reels,
		PaylinesHit: []PaylineHit{},
		TotalWin:    0,
	}
}

func (e *SlotEngine) generateWinningSpin(betAmount int64) *SpinResult {
	var reels [3][3]string

	winningSymbols := []string{"CHERRY", "LEMON", "ORANGE", "PLUM", "BELL", "BAR"}
	selectedSymbol := winningSymbols[e.randomInt(0, len(winningSymbols)-1)]

	for col := 0; col < 3; col++ {
		reels[1][col] = selectedSymbol
	}

	for col := 0; col < 3; col++ {
		for row := 0; row < 3; row++ {
			if row != 1 {
				reel := ReelsConfig[col]
				randomIndex := e.randomInt(0, len(reel)-1)
				reels[row][col] = reel[randomIndex]
			}
		}
	}

	result := &SpinResult{
		Reels:       reels,
		PaylinesHit: []PaylineHit{},
		TotalWin:    0,
	}

	totalWin := e.calculateWin(reels, betAmount)
	result.TotalWin = totalWin

	return result
}

func (e *SlotEngine) calculateWin(reels [3][3]string, betAmount int64) int64 {
	var totalWin int64
	var paylinesHit []PaylineHit

	for _, payline := range Paylines {
		symbols := make([]string, 3)
		for i, point := range payline.Points {
			row, col := point[0], point[1]
			symbols[i] = reels[row][col]
		}

		firstSymbol := symbols[0]
		count := 1
		for i := 1; i < 3; i++ {
			if symbols[i] == firstSymbol {
				count++
			} else {
				break
			}
		}

		if count >= 1 {
			if payout, exists := Payouts[firstSymbol]; exists {
				if winAmount, exists := payout[count]; exists {
					win := winAmount * (betAmount / 100)
					if win > 0 {
						totalWin += win
						paylinesHit = append(paylinesHit, PaylineHit{
							PaylineID: payline.ID,
							Symbol:    firstSymbol,
							Count:     count,
							WinAmount: win,
						})
					}
				}
			}
		}
	}

	return totalWin
}

func (e *SlotEngine) randomInt(min, max int) int {
	n, _ := rand.Int(rand.Reader, big.NewInt(int64(max-min+1)))
	return min + int(n.Int64())
}

func (e *SlotEngine) GetRuntimeConfig() *model.GameRuntimeConfig {
	return e.runtimeConfig
}
