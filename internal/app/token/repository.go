package token

import "context"

type Repository interface {
	Add(ctx context.Context, ticker string) (bool, error)
	Update(ctx context.Context, ticker string, price float64) error
	ListTickers(ctx context.Context) ([]string, error)
}
