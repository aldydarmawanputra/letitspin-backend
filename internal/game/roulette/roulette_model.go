package roulette

type RouletteResult struct {
	Number     int    `json:"number"`
	Color      string `json:"color"`
	Parity     string `json:"parity"`
	BetType    string `json:"bet_type"`
	BetValue   string `json:"bet_value"`
	WinAmount  int64  `json:"win_amount"`
	Multiplier int    `json:"multiplier"`
}

type BetType string

const (
	BetTypeRed    BetType = "RED"
	BetTypeBlack  BetType = "BLACK"
	BetTypeGreen  BetType = "GREEN"
	BetTypeOdd    BetType = "ODD"
	BetTypeEven   BetType = "EVEN"
	BetTypeNumber BetType = "NUMBER"
)

type RouletteNumber struct {
	Number int
	Color  string
	Parity string
}

var RouletteWheel = []RouletteNumber{
	{0, "GREEN", "ZERO"},
	{32, "RED", "EVEN"}, {15, "BLACK", "ODD"}, {19, "RED", "ODD"}, {4, "BLACK", "EVEN"},
	{21, "RED", "ODD"}, {2, "BLACK", "EVEN"}, {25, "RED", "ODD"}, {17, "BLACK", "ODD"},
	{34, "RED", "EVEN"}, {6, "BLACK", "EVEN"}, {27, "RED", "ODD"}, {13, "BLACK", "ODD"},
	{36, "RED", "EVEN"}, {11, "BLACK", "ODD"}, {30, "RED", "EVEN"}, {8, "BLACK", "EVEN"},
	{23, "RED", "ODD"}, {10, "BLACK", "EVEN"}, {5, "RED", "ODD"}, {24, "BLACK", "EVEN"},
	{16, "RED", "EVEN"}, {33, "BLACK", "ODD"}, {1, "RED", "ODD"}, {20, "BLACK", "EVEN"},
	{14, "RED", "EVEN"}, {31, "BLACK", "ODD"}, {9, "RED", "ODD"}, {22, "BLACK", "EVEN"},
	{18, "RED", "EVEN"}, {29, "BLACK", "ODD"}, {7, "RED", "ODD"}, {28, "BLACK", "EVEN"},
	{12, "RED", "EVEN"}, {35, "BLACK", "ODD"}, {3, "RED", "ODD"}, {26, "BLACK", "EVEN"},
}

var Multipliers = map[BetType]int{
	BetTypeRed:    2,
	BetTypeBlack:  2,
	BetTypeGreen:  35,
	BetTypeOdd:    2,
	BetTypeEven:   2,
	BetTypeNumber: 35,
}
