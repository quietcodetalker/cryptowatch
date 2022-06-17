package trigger

import "context"

type Repository interface {
	Add(ctx context.Context, userID uint64, ticker string) error
	Remove(ctx context.Context, userID uint64, ticker string) error
	Check(ctx context.Context, userID uint64, ticker string) (bool, error)
}
