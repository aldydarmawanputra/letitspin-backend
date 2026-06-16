package blackjack

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"math/big"

	"let-it-spin/internal/game/base"
	"let-it-spin/internal/game/repository"
	"let-it-spin/internal/model"

	"github.com/google/uuid"
)

type BlackjackEngine struct {
	gameCode      string
	configRepo    *repository.ConfigRepository
	bjRepo        *BlackjackRepository
	runtimeConfig *model.GameRuntimeConfig
	deck          []Card
}

func NewBlackjackEngine(configRepo *repository.ConfigRepository, bjRepo *BlackjackRepository) *BlackjackEngine {
	return &BlackjackEngine{
		gameCode:   "BLACKJACK",
		configRepo: configRepo,
		bjRepo:     bjRepo,
	}
}

func (e *BlackjackEngine) loadConfig(ctx context.Context) error {
	config, err := e.configRepo.GetGameConfig(ctx, e.gameCode)
	if err != nil {
		return err
	}
	e.runtimeConfig = config
	return nil
}

func (e *BlackjackEngine) GetGameCode() string {
	return e.gameCode
}

func (e *BlackjackEngine) GetMinBet() int64 {
	if e.runtimeConfig != nil {
		return e.runtimeConfig.MinBet
	}
	return 100
}

func (e *BlackjackEngine) GetMaxBet() int64 {
	if e.runtimeConfig != nil {
		return e.runtimeConfig.MaxBet
	}
	return 100000
}

func (e *BlackjackEngine) ValidateBet(amount int64) error {
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

func (e *BlackjackEngine) Execute(ctx context.Context, userID uuid.UUID, betAmount int64, options map[string]interface{}) (*base.GameResult, error) {
	if err := e.loadConfig(ctx); err != nil {
		return nil, err
	}

	action, ok := options["action"].(string)
	if !ok {
		action = "START"
	}

	e.initDeck()

	switch action {
	case "START":
		return e.startGame(ctx, userID, betAmount)
	case "HIT":
		return e.hit(ctx, betAmount, options)
	case "STAND":
		return e.stand(ctx, betAmount, options)
	default:
		return nil, fmt.Errorf("invalid action: %s", action)
	}
}

func (e *BlackjackEngine) startGame(ctx context.Context, userID uuid.UUID, betAmount int64) (*base.GameResult, error) {
	e.initDeck()

	playerCards := []Card{e.drawCard(), e.drawCard()}
	dealerCards := []Card{e.drawCard(), e.drawCard()}

	playerValue := e.calculateHandValue(playerCards)
	dealerValue := e.calculateHandValue(dealerCards)
	isBlackjack := playerValue == 21 && len(playerCards) == 2

	playerCardsJSON, _ := json.Marshal(playerCards)
	dealerCardsJSON, _ := json.Marshal(dealerCards)

	session := &model.BlackjackSession{
		ID:           uuid.New(),
		SessionID:    uuid.New().String(),
		UserID:       userID,
		BetAmount:    betAmount,
		PlayerCards:  playerCardsJSON,
		DealerCards:  dealerCardsJSON,
		PlayerValue:  playerValue,
		DealerValue:  dealerValue,
		IsPlayerBust: playerValue > 21,
		IsBlackjack:  isBlackjack,
		GameStatus:   model.BlackjackStatusPlaying,
	}

	if isBlackjack {
		session.GameStatus = model.BlackjackStatusPlayerWin
	}

	err := e.bjRepo.CreateSession(ctx, session)
	if err != nil {
		return nil, err
	}

	return e.sessionToResult(session), nil
}

func (e *BlackjackEngine) hit(ctx context.Context, betAmount int64, options map[string]interface{}) (*base.GameResult, error) {
	sessionID, ok := options["session_id"].(string)
	if !ok {
		return nil, fmt.Errorf("session_id is required for HIT action")
	}

	session, err := e.bjRepo.GetSession(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	var playerCards []Card
	json.Unmarshal(session.PlayerCards, &playerCards)

	newCard := e.drawCard()
	playerCards = append(playerCards, newCard)
	playerValue := e.calculateHandValue(playerCards)

	playerCardsJSON, _ := json.Marshal(playerCards)

	session.PlayerCards = playerCardsJSON
	session.PlayerValue = playerValue
	session.IsPlayerBust = playerValue > 21

	if playerValue > 21 {
		session.GameStatus = model.BlackjackStatusDealerWin
	} else if playerValue == 21 {
		return e.standFromSession(ctx, session)
	}

	err = e.bjRepo.UpdateSession(ctx, session)
	if err != nil {
		return nil, err
	}

	return e.sessionToResult(session), nil
}

func (e *BlackjackEngine) stand(ctx context.Context, betAmount int64, options map[string]interface{}) (*base.GameResult, error) {
	sessionID, ok := options["session_id"].(string)
	if !ok {
		return nil, fmt.Errorf("session_id is required for STAND action")
	}

	session, err := e.bjRepo.GetSession(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	return e.standFromSession(ctx, session)
}

func (e *BlackjackEngine) standFromSession(ctx context.Context, session *model.BlackjackSession) (*base.GameResult, error) {
	var dealerCards []Card
	json.Unmarshal(session.DealerCards, &dealerCards)

	dealerValue := session.DealerValue

	for dealerValue < 17 {
		newCard := e.drawCard()
		dealerCards = append(dealerCards, newCard)
		dealerValue = e.calculateHandValue(dealerCards)
	}

	dealerCardsJSON, _ := json.Marshal(dealerCards)
	session.DealerCards = dealerCardsJSON
	session.DealerValue = dealerValue

	if dealerValue > 21 {
		session.GameStatus = model.BlackjackStatusPlayerWin
	} else if session.PlayerValue > dealerValue {
		session.GameStatus = model.BlackjackStatusPlayerWin
	} else if session.PlayerValue < dealerValue {
		session.GameStatus = model.BlackjackStatusDealerWin
	} else {
		session.GameStatus = model.BlackjackStatusPush
	}

	err := e.bjRepo.UpdateSession(ctx, session)
	if err != nil {
		return nil, err
	}

	result := e.sessionToResult(session)

	e.bjRepo.DeleteSession(ctx, session.SessionID)

	return result, nil
}

func (e *BlackjackEngine) sessionToResult(session *model.BlackjackSession) *base.GameResult {
	var winAmount int64
	var multiplier float64
	isWin := false

	if session.GameStatus == model.BlackjackStatusPlayerWin {
		isWin = true
		multiplier = 2.0
		if session.IsBlackjack {
			multiplier = 2.5
		}
		winAmount = int64(float64(session.BetAmount) * multiplier)
	} else if session.GameStatus == model.BlackjackStatusPush {
		winAmount = session.BetAmount
		multiplier = 1.0
	}

	var playerCards []Card
	var dealerCards []Card
	json.Unmarshal(session.PlayerCards, &playerCards)
	json.Unmarshal(session.DealerCards, &dealerCards)

	gameState := &GameState{
		PlayerCards:  playerCards,
		DealerCards:  dealerCards,
		PlayerValue:  session.PlayerValue,
		DealerValue:  session.DealerValue,
		IsPlayerBust: session.IsPlayerBust,
		IsBlackjack:  session.IsBlackjack,
		GameStatus:   session.GameStatus,
		WinAmount:    winAmount,
		Multiplier:   multiplier,
	}

	sessionUUID, _ := uuid.Parse(session.SessionID)

	return &base.GameResult{
		SessionID:  sessionUUID,
		WinAmount:  winAmount,
		IsWin:      isWin,
		Details:    gameState,
		Multiplier: multiplier,
	}
}

func (e *BlackjackEngine) initDeck() {
	suits := []string{"♠", "♥", "♣", "♦"}
	ranks := []string{"2", "3", "4", "5", "6", "7", "8", "9", "10", "J", "Q", "K", "A"}

	e.deck = []Card{}
	for _, suit := range suits {
		for _, rank := range ranks {
			value := e.getCardValue(rank)
			e.deck = append(e.deck, Card{Suit: suit, Rank: rank, Value: value})
		}
	}

	for i := range e.deck {
		j := e.randomInt(0, len(e.deck)-1)
		e.deck[i], e.deck[j] = e.deck[j], e.deck[i]
	}
}

func (e *BlackjackEngine) getCardValue(rank string) int {
	switch rank {
	case "J", "Q", "K":
		return 10
	case "A":
		return 11
	default:
		var val int
		switch rank {
		case "2":
			val = 2
		case "3":
			val = 3
		case "4":
			val = 4
		case "5":
			val = 5
		case "6":
			val = 6
		case "7":
			val = 7
		case "8":
			val = 8
		case "9":
			val = 9
		case "10":
			val = 10
		}
		return val
	}
}

func (e *BlackjackEngine) calculateHandValue(cards []Card) int {
	value := 0
	aceCount := 0

	for _, card := range cards {
		if card.Rank == "A" {
			aceCount++
		}
		value += card.Value
	}

	for aceCount > 0 && value > 21 {
		value -= 10
		aceCount--
	}

	return value
}

func (e *BlackjackEngine) drawCard() Card {
	if len(e.deck) == 0 {
		e.initDeck()
	}
	card := e.deck[0]
	e.deck = e.deck[1:]
	return card
}

func (e *BlackjackEngine) randomInt(min, max int) int {
	n, _ := rand.Int(rand.Reader, big.NewInt(int64(max-min+1)))
	return min + int(n.Int64())
}
