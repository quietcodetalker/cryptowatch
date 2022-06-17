package user

import (
	"context"
	"cryptowatch/pkg/util"
	"cryptowatch/pkg/util/authtoken"
	"errors"
	"time"
)

type Service interface {
	Create(ctx context.Context, req SvcCreateReq) (*User, error)
	Login(ctx context.Context, req SvcLoginReq) (string, error)
	GetByUsername(ctx context.Context, username string) (*User, error)
	GenerateOTP(ctx context.Context, username string) error
	GetOTP(ctx context.Context, userID uint64) (string, error)
	VerifyOTP(ctx context.Context, username string, code string) (*SvcVerifyOTPRes, error)
}

type SvcVerifyOTPRes struct {
	UserID uint64 `json:"user_id"`
	Token  string `json:"token"`
}

type SvcCreateReq struct {
	Username  string `json:"username" validate:"required"`
	Password  string `json:"password" validate:"required"`
	FirstName string `json:"first_name" validate:"required"`
	LastName  string `json:"last_name" validate:"required"`
}

type SvcLoginReq struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

type service struct {
	repo           Repository
	authtokenMaker authtoken.Maker
	otpManager     OTPManager
}

func NewService(repo Repository, authtokenMaker authtoken.Maker, otpManager OTPManager) *service {
	return &service{
		repo:           repo,
		authtokenMaker: authtokenMaker,
		otpManager:     otpManager,
	}
}

func (s *service) Create(ctx context.Context, req SvcCreateReq) (*User, error) {
	passwordHash, err := util.HashPassword(req.Password)
	if err != nil {
		return nil, ErrInternalError
	}

	u, err := s.repo.Create(ctx, RepoCreateReq{
		Username:     req.Username,
		PasswordHash: passwordHash,
		FirstName:    req.FirstName,
		LastName:     req.LastName,
	})
	if err != nil {
		return nil, err
	}

	return u, nil
}

func (s *service) Login(ctx context.Context, req SvcLoginReq) (string, error) {
	u, err := s.repo.GetByUsername(ctx, req.Username)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return "", ErrUnauthenticated
		}
		return "", err
	}

	err = util.CheckPassword(req.Password, u.PasswordHash)
	if err != nil {
		return "", ErrUnauthenticated
	}

	token, err := s.authtokenMaker.CreateToken(u.ID, time.Hour*24)
	if err != nil {
		return "", ErrInternalError
	}

	return token, nil
}

func (s *service) GetByUsername(ctx context.Context, username string) (*User, error) {
	return s.repo.GetByUsername(ctx, username)
}

func (s *service) GenerateOTP(ctx context.Context, username string) error {
	u, err := s.repo.GetByUsername(ctx, username)
	if err != nil {
		return err
	}

	s.otpManager.Add(ctx, u.ID)

	return nil
}

func (s *service) GetOTP(ctx context.Context, userID uint64) (string, error) {
	return s.otpManager.Get(ctx, userID)
}

func (s *service) VerifyOTP(ctx context.Context, username string, code string) (*SvcVerifyOTPRes, error) {
	u, err := s.repo.GetByUsername(ctx, username)
	if err != nil {
		return nil, err
	}

	if err = s.otpManager.Verify(ctx, u.ID, code); err != nil {
		return nil, err
	}

	token, err := s.authtokenMaker.CreateToken(u.ID, time.Hour*999) // TODO implement endless token
	if err != nil {
		return nil, err
	}

	return &SvcVerifyOTPRes{
		UserID: u.ID,
		Token:  token,
	}, nil
}
