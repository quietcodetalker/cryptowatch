package token

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

var (
	tokensTable = "tokens"
)

type postgresRepo struct {
	db *pgxpool.Pool
}

func NewPostgresRepo(db *pgxpool.Pool) *postgresRepo {
	return &postgresRepo{
		db: db,
	}
}

var addQuery = fmt.Sprintf(`
INSERT INTO %s
(ticker)
VALUES ($1)
ON CONFLICT (ticker)
DO NOTHING
`, tokensTable)

// Add inserts a new row into the db table.
// It returns bool values which is true when new rows is inserted and
// false otherwise (i.e. such token already exists)
// and error.
func (r *postgresRepo) Add(ctx context.Context, ticker string) (bool, error) {
	cmd, err := r.db.Exec(ctx, addQuery, ticker)
	if err != nil {
		return false, fmt.Errorf("exec query error: %w", ErrInternalError)
	}

	if cmd.RowsAffected() == 0 {
		return false, nil
	}

	return true, nil
}

var updateQuery = fmt.Sprintf(`
UPDATE %s
SET price = $2
WHERE ticker = $1
`, tokensTable)

// Update updates db table rows setting price where ticker is equal to given one.
// If now rows affected ErrNotFound returned.
// If any other error occurred wrapped ErrInternalError returned.
func (r *postgresRepo) Update(ctx context.Context, ticker string, price float64) error {
	cmd, err := r.db.Exec(ctx, updateQuery, ticker, price)
	if err != nil {
		return fmt.Errorf("exec query error: %w", ErrInternalError)
	}

	if cmd.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}

var listTickersQuery = fmt.Sprintf(`
SELECT ticker FROM %s
`, tokensTable)

func (r *postgresRepo) ListTickers(ctx context.Context) ([]string, error) {
	var tickers []string
	rows, err := r.db.Query(ctx, listTickersQuery)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return tickers, nil
		}
		return nil, ErrInternalError
	}

	var ticker string
	for rows.Next() {
		err = rows.Scan(&ticker)
		if err != nil {
			return nil, ErrInternalError
		}
		tickers = append(tickers, ticker)
	}
	if rows.Err() != nil {
		return nil, ErrInternalError
	}

	return tickers, nil
}
