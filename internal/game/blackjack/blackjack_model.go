package blackjack

type Card struct {
	Suit  string `json:"suit"`
	Rank  string `json:"rank"`
	Value int    `json:"value"`
}

type GameState struct {
	PlayerCards  []Card  `json:"player_cards"`
	DealerCards  []Card  `json:"dealer_cards"`
	PlayerValue  int     `json:"player_value"`
	DealerValue  int     `json:"dealer_value"`
	IsPlayerBust bool    `json:"is_player_bust"`
	IsDealerBust bool    `json:"is_dealer_bust"`
	IsBlackjack  bool    `json:"is_blackjack"`
	GameStatus   string  `json:"game_status"`
	WinAmount    int64   `json:"win_amount"`
	Multiplier   float64 `json:"multiplier"`
}

type Action string

const (
	ActionHit    Action = "HIT"
	ActionStand  Action = "STAND"
	ActionDouble Action = "DOUBLE"
)

const (
	StatusPlaying   = "PLAYING"
	StatusPlayerWin = "PLAYER_WIN"
	StatusDealerWin = "DEALER_WIN"
	StatusPush      = "PUSH"
)

var Suits = []string{"♠", "♥", "♣", "♦"}
var Ranks = []string{"2", "3", "4", "5", "6", "7", "8", "9", "10", "J", "Q", "K", "A"}
