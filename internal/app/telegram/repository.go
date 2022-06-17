package telegram

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

const (
	telegramAccountsTable = "telegram_accounts"
)

type Repository interface {
	AddAccount(ctx context.Context, id int64) error
	GetAccount(ctx context.Context, id int64) (*Account, error)
	SetAuthToken(ctx context.Context, id int64, token string, userID uint64) error
}

type postgresRepo struct {
	db *pgxpool.Pool
}

func NewPostgresRepo(db *pgxpool.Pool) *postgresRepo {
	return &postgresRepo{db: db}
}

var addAccountQuery = fmt.Sprintf(`
INSERT INTO %s
(id)
VALUES ($1)
ON CONFLICT (id)
DO NOTHING
`, telegramAccountsTable)

func (r *postgresRepo) AddAccount(ctx context.Context, id int64) error {
	_, err := r.db.Exec(ctx, addAccountQuery, id)
	if err != nil {
		return ErrInternalError
	}

	return nil
}

var getAccountQuery = fmt.Sprintf(`
SELECT auth_token, user_id FROM %s
WHERE id = $1
`, telegramAccountsTable)

func (r *postgresRepo) GetAccount(ctx context.Context, id int64) (*Account, error) {
	acc := Account{ID: id}
	err := r.db.QueryRow(ctx, getAccountQuery, id).Scan(&acc.AuthToken, &acc.UserID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, ErrInternalError
	}

	return &acc, nil
}

var setAuthTokenQuery = fmt.Sprintf(`
UPDATE %s
SET
	auth_token = $2, 
	user_id = $3
WHERE id = $1
`, telegramAccountsTable)

func (r *postgresRepo) SetAuthToken(ctx context.Context, id int64, token string, userID uint64) error {
	cmd, err := r.db.Exec(ctx, setAuthTokenQuery, id, token, userID)
	if err != nil {
		return ErrInternalError
	}
	if cmd.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}
