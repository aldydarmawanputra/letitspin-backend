package slot

var Symbols = map[string]Symbol{
	"CHERRY": {Code: "CHERRY", Name: "Cherry", Value: 10, Image: "🍒"},
	"LEMON":  {Code: "LEMON", Name: "Lemon", Value: 20, Image: "🍋"},
	"ORANGE": {Code: "ORANGE", Name: "Orange", Value: 30, Image: "🍊"},
	"PLUM":   {Code: "PLUM", Name: "Plum", Value: 40, Image: " plum"},
	"BELL":   {Code: "BELL", Name: "Bell", Value: 50, Image: "🔔"},
	"BAR":    {Code: "BAR", Name: "Bar", Value: 100, Image: "BAR"},
	"SEVEN":  {Code: "SEVEN", Name: "Seven", Value: 500, Image: "7️⃣"},
}

var ReelsConfig = [3][]string{
	// Reel 1
	{"CHERRY", "LEMON", "ORANGE", "PLUM", "BELL", "BAR", "SEVEN", "CHERRY", "LEMON", "ORANGE"},
	// Reel 2
	{"CHERRY", "ORANGE", "PLUM", "BELL", "BAR", "SEVEN", "LEMON", "CHERRY", "ORANGE", "PLUM"},
	// Reel 3
	{"CHERRY", "PLUM", "BELL", "BAR", "SEVEN", "ORANGE", "LEMON", "CHERRY", "PLUM", "BELL"},
}

var Paylines = []Payline{
	{ID: 1, Name: "Top Row", Points: [][2]int{{0, 0}, {0, 1}, {0, 2}}},
	{ID: 2, Name: "Middle Row", Points: [][2]int{{1, 0}, {1, 1}, {1, 2}}},
	{ID: 3, Name: "Bottom Row", Points: [][2]int{{2, 0}, {2, 1}, {2, 2}}},
	{ID: 4, Name: "Diagonal Down", Points: [][2]int{{0, 0}, {1, 1}, {2, 2}}},
	{ID: 5, Name: "Diagonal Up", Points: [][2]int{{2, 0}, {1, 1}, {0, 2}}},
}

var Payouts = map[string]map[int]int64{
	"CHERRY": {1: 5, 2: 25, 3: 100},
	"LEMON":  {1: 10, 2: 50, 3: 200},
	"ORANGE": {1: 15, 2: 75, 3: 300},
	"PLUM":   {1: 20, 2: 100, 3: 400},
	"BELL":   {1: 25, 2: 125, 3: 500},
	"BAR":    {1: 50, 2: 250, 3: 1000},
	"SEVEN":  {1: 100, 2: 500, 3: 5000},
}
