package trigger

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

var (
	triggersTable = "triggers"
)

type postgresRepo struct {
	db *pgxpool.Pool
}

func NewPostgresRepo(db *pgxpool.Pool) *postgresRepo {
	return &postgresRepo{db: db}
}

var addQuery = fmt.Sprintf(`
INSERT INTO %s
(user_id, token_ticker)
VALUES ($1, $2)
`, triggersTable)

func (r *postgresRepo) Add(ctx context.Context, userID uint64, ticker string) error {
	_, err := r.db.Exec(ctx, addQuery, userID, ticker)
	if err != nil {
		return ErrInternalError
	}

	return nil
}

var removeQuery = fmt.Sprintf(`
REMOVE FROM  %s
WHERE user_id = $1 AND token_ticker = $2
`, triggersTable)

func (r *postgresRepo) Remove(ctx context.Context, userID uint64, ticker string) error {
	_, err := r.db.Exec(ctx, removeQuery, userID, ticker)
	if err != nil {
		return ErrInternalError
	}

	return nil
}

var checkQuery = fmt.Sprintf(`
SELECT * FROM %s
WHERE user_id = $1 AND token_ticker = $2
`, triggersTable)

func (r *postgresRepo) Check(ctx context.Context, userID uint64, ticker string) (bool, error) {
	rows, err := r.db.Query(ctx, checkQuery, userID, ticker)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}
		return false, ErrInternalError
	}

	defer rows.Close()
	if rows.Err() != nil {
		return false, ErrInternalError
	}

	return true, nil
}
