package token

import "context"

type Exchange interface {
	Subscribe(ctx context.Context, tickers []string) error
	GetPrices(ctx context.Context, tokens []string) (map[string]float64, error)
	Start(ctx context.Context, ch chan<- map[string]float64) error
}
