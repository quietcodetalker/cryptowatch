package trigger

type Update struct {
	Ticker string  `json:"ticker"`
	Delta  float64 `json:"delta"`
}
