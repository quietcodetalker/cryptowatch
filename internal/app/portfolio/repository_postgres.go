package portfolio

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

const (
	portfoliosTable   = "portfolios"
	tokensTable       = "tokens"
	transactionsTable = "transactions"
)

type DBTX interface {
	QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row
	Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
	Exec(ctx context.Context, sql string, arguments ...interface{}) (pgconn.CommandTag, error)
}

type postgresRepo struct {
	*postgresQueries
	db *pgxpool.Pool
}

func NewPostgresRepo(db *pgxpool.Pool) *postgresRepo {
	return &postgresRepo{
		db: db,
		postgresQueries: &postgresQueries{
			db: db,
		},
	}
}

func (p *postgresRepo) Buy(ctx context.Context, userID uint64, portfolioID uint64, ticker string, quantity float64, price float64, fee float64) error {
	err := p.execTx(ctx, func(q *postgresQueries) error {
		_, err := q.GetPortfolio(ctx, userID, portfolioID) // Check whether portfolio belongs to user.
		if err != nil {
			return nil
		}

		err = q.createToken(ctx, ticker)
		if err != nil {
			return err
		}

		_, err = q.createTransaction(ctx, portfolioID, ticker, quantity, price, fee)
		if err != nil {
			return err
		}

		return nil
	})

	return err
}

func (p *postgresRepo) Sell(ctx context.Context, userID uint64, portfolioID uint64, ticker string, quantity float64, price float64, fee float64) error {
	err := p.execTx(ctx, func(q *postgresQueries) error {
		_, err := q.GetPortfolio(ctx, userID, portfolioID) // Check whether portfolio belongs to user.
		if err != nil {
			return nil
		}

		err = q.createToken(ctx, ticker)
		if err != nil {
			return err
		}

		_, err = q.createTransaction(ctx, portfolioID, ticker, -quantity, price, fee)
		if err != nil {
			return err
		}

		return nil
	})

	return err
}

func (p *postgresRepo) execTx(ctx context.Context, fn func(queries *postgresQueries) error) error {
	tx, err := p.db.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.ReadCommitted})
	if err != nil {
		return err
	}

	q := &postgresQueries{db: tx}
	err = fn(q)

	if err != nil {
		if rbErr := tx.Rollback(ctx); rbErr != nil {
			return fmt.Errorf("tx err: %v, rb err: %v", err, rbErr)
		}
		return err
	}

	return tx.Commit(ctx)
}

type postgresQueries struct {
	db DBTX
}

var createTokenQuery = fmt.Sprintf(`
INSERT INTO %s
(ticker)
VALUES ($1)
ON CONFLICT (ticker)
DO NOTHING
`, tokensTable)

func (q *postgresQueries) createToken(ctx context.Context, ticker string) error {
	_, err := q.db.Exec(ctx, createTokenQuery, ticker)
	if err != nil {
		return ErrInternalError
	}

	return nil
}

var createTransactionQuery = fmt.Sprintf(`
INSERT INTO %s
(portfolio_id, token_ticker, quantity, price, fee)
VALUES ($1, $2, $3, $4, $5)
RETURNING id
`, transactionsTable)

func (q *postgresQueries) createTransaction(
	ctx context.Context,
	portfolioID uint64,
	tokenTicker string,
	quantity float64,
	price float64,
	fee float64,
) (uint64, error) {
	var id uint64
	err := q.db.QueryRow(ctx, createTransactionQuery, portfolioID, tokenTicker, quantity, price, fee).Scan(&id)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.ConstraintName == "transactions_portfolio_id_fkey" ||
				pgErr.ConstraintName == "transactions_token_ticker_fkey" {
				return 0, ErrFailedPrecondition
			}
		}
		return 0, ErrInternalError
	}

	return id, nil
}

var createPortfolioQuery = fmt.Sprintf(`
INSERT INTO %s
(user_id, name)
VALUES ($1, $2)
RETURNING id
`, portfoliosTable)

func (q *postgresQueries) CreatePortfolio(ctx context.Context, userID uint64, name string) (uint64, error) {
	var id uint64
	err := q.db.QueryRow(ctx, createPortfolioQuery, userID, name).Scan(&id)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.ConstraintName == "portfolios_user_id_fkey" ||
				pgErr.ConstraintName == "portfolios_user_id_name_key" {
				return 0, ErrFailedPrecondition
			}
		}
		return 0, ErrInternalError
	}

	return id, nil
}

var getPortfolioQuery = fmt.Sprintf(`
SELECT id, user_id, name
FROM %s
WHERE user_id = $1 AND id = $2
`, portfoliosTable)

func (q *postgresQueries) GetPortfolio(ctx context.Context, userID uint64, portfolioID uint64) (*Portfolio, error) {
	var p Portfolio
	err := q.db.QueryRow(ctx, getPortfolioQuery, userID, portfolioID).Scan(&p.ID, &p.UserID, &p.Name)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, ErrInternalError
	}

	return &p, nil
}

var infoQuery = fmt.Sprintf(`
SELECT SUM ((tk.price - tr.price) * tr.quantity - tr.fee)
FROM %s tr
INNER JOIN %s tk ON tk.ticker = tr.token_ticker
WHERE tr.portfolio_id = $1;
`, transactionsTable, tokensTable)

func (r *postgresRepo) Info(ctx context.Context, userID uint64, portfolioID uint64) (*RepoInfoRes, error) {
	var res RepoInfoRes

	err := r.execTx(ctx, func(q *postgresQueries) error {
		_, err := q.GetPortfolio(ctx, userID, portfolioID) // Check whether portfolio belongs to user.
		if err != nil {
			return nil
		}

		err = q.db.QueryRow(ctx, infoQuery, portfolioID).Scan(&res.Profit)
		if err != nil {
			return ErrInternalError
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return &res, nil
}
