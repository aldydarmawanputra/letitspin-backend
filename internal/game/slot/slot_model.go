package slot

type Symbol struct {
	Code  string `json:"code"`
	Name  string `json:"name"`
	Value int64  `json:"value"`
	Image string `json:"image"`
}

type Reel struct {
	Symbols []string `json:"symbols"`
}

type Payline struct {
	ID     int      `json:"id"`
	Name   string   `json:"name"`
	Points [][2]int `json:"points"`
}

type SpinResult struct {
	Reels       [3][3]string `json:"reels"`
	PaylinesHit []PaylineHit `json:"paylines_hit"`
	TotalWin    int64        `json:"total_win"`
}

type PaylineHit struct {
	PaylineID int    `json:"payline_id"`
	Symbol    string `json:"symbol"`
	Count     int    `json:"count"`
	WinAmount int64  `json:"win_amount"`
}
