package dice

type DiceResult struct {
	RollNumber int    `json:"roll_number"`
	Prediction string `json:"prediction"`
	BetValue   *int   `json:"bet_value,omitempty"`
	WinAmount  int64  `json:"win_amount"`
	Multiplier int    `json:"multiplier"`
}

type DiceBetType string

const (
	BetTypeUnder DiceBetType = "UNDER"
	BetTypeOver  DiceBetType = "OVER"
	BetTypeExact DiceBetType = "EXACT"
)

var Multipliers = map[string]int{
	"UNDER": 2,
	"OVER":  2,
	"EXACT": 50,
	"EDGE":  100,
}
