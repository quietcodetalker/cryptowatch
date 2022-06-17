package util

import (
	"context"
	"github.com/jackc/pgx/v4/pgxpool"
	"time"
)

func OpenDB(source string) (*pgxpool.Pool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	pool, err := pgxpool.Connect(ctx, source)
	if err != nil {
		return nil, err
	}

	return pool, nil
}
