package user

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

const (
	usersTable = "users"
)

type postgresRepo struct {
	db *pgxpool.Pool
}

func NewPostgresRepo(db *pgxpool.Pool) *postgresRepo {
	return &postgresRepo{
		db: db,
	}
}

var createQuery = fmt.Sprintf(`
INSERT INTO %s
(username, password_hash, first_name, last_name)
VALUES ($1, $2, $3, $4)
RETURNING id, username, password_hash, first_name, last_name, create_time
`, usersTable)

func (r *postgresRepo) Create(ctx context.Context, req RepoCreateReq) (*User, error) {
	var newUser User
	err := r.db.QueryRow(ctx, createQuery, req.Username, req.PasswordHash, req.FirstName, req.LastName).
		Scan(
			&newUser.ID,
			&newUser.Username,
			&newUser.PasswordHash,
			&newUser.FirstName,
			&newUser.LastName,
			&newUser.CreateTime,
		)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.ConstraintName {
			case "users_username_key":
				return nil, ErrFailedPrecondition
			case "users_username_valid":
				return nil, ErrInvalidArgument
			}
		}
		return nil, ErrInternalError
	}

	return &newUser, nil
}

var getByUsernameQuery = fmt.Sprintf(`
SELECT id, username, password_hash, first_name, last_name, create_time
FROM %s
WHERE username = $1
`, usersTable)

func (r *postgresRepo) GetByUsername(ctx context.Context, username string) (*User, error) {
	var u User
	err := r.db.QueryRow(ctx, getByUsernameQuery, username).
		Scan(
			&u.ID,
			&u.Username,
			&u.PasswordHash,
			&u.FirstName,
			&u.LastName,
			&u.CreateTime,
		)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, ErrInternalError
	}

	return &u, nil
}
