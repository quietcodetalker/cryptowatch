package portfolio

import "context"

type Repository interface {
	Buy(ctx context.Context, userID uint64, portfolioID uint64, ticker string, quantity float64, price float64, fee float64) error
	Sell(ctx context.Context, userID uint64, portfolioID uint64, ticker string, quantity float64, price float64, fee float64) error
	CreatePortfolio(ctx context.Context, userID uint64, name string) (uint64, error)
	Info(ctx context.Context, userID uint64, portfolioID uint64) (*RepoInfoRes, error)
}

type RepoInfoRes struct {
	Profit float64 `json:"profit"`
}
