package token

type Token struct {
	Ticker string  `json:"ticker"`
	Price  float64 `json:"price"`
}
