//go:generate mockgen -destination=mock/user.go -package=mock . Repository
package user

import "context"

type Repository interface {
	Create(ctx context.Context, req RepoCreateReq) (*User, error)
	GetByUsername(ctx context.Context, username string) (*User, error)
}

type RepoCreateReq struct {
	Username     string `json:"username" validate:"required"`
	PasswordHash string `json:"password_hash" validate:"required"`
	FirstName    string `json:"first_name" validate:"required"`
	LastName     string `json:"last_name" validate:"required"`
}
