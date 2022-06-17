package portfolio

type Portfolio struct {
	ID     uint64 `json:"id"`
	UserID uint64 `json:"user_id"`
	Name   string `json:"name"`
}

type Transaction struct {
	ID          uint64  `json:"id"`
	PortfolioID uint64  `json:"portfolio_id"`
	TokenTicker string  `json:"token_ticker"`
	Quantity    float64 `json:"quantity"`
	Price       float64 `json:"price"`
	Fee         float64 `json:"fee"`
}
