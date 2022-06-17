package portfolio

import (
	"context"
	"cryptowatch/internal/app/token"
)

type Service interface {
	Buy(ctx context.Context, userID uint64, portfolioID uint64, ticker string, quantity float64, price float64, fee float64) error
	Sell(ctx context.Context, userID uint64, portfolioID uint64, ticker string, quantity float64, price float64, fee float64) error
	CreatePortfolio(ctx context.Context, userID uint64, name string) (uint64, error)
	Info(ctx context.Context, usreID uint64, portfolioID uint64) (*SvcInfoRes, error)
}

type SvcInfoRes struct {
	Profit float64 `json:"profit"`
}

type service struct {
	repo     Repository
	tokenSvc token.Service

	updates chan map[string]float64
}

func NewService(repo Repository, tokenSvc token.Service) *service {
	return &service{
		repo:     repo,
		tokenSvc: tokenSvc,

		updates: make(chan map[string]float64),
	}
}

func (s *service) Buy(ctx context.Context, userID uint64, portfolioID uint64, ticker string, quantity float64, price float64, fee float64) error {
	_, err := s.tokenSvc.Add(ctx, ticker)
	if err != nil {
		return ErrInternalError
	}
	return s.repo.Buy(ctx, userID, portfolioID, ticker, quantity, price, fee)
}

func (s *service) Sell(ctx context.Context, userID uint64, portfolioID uint64, ticker string, quantity float64, price float64, fee float64) error {
	_, err := s.tokenSvc.Add(ctx, ticker)
	if err != nil {
		return ErrInternalError
	}
	return s.repo.Sell(ctx, userID, portfolioID, ticker, quantity, price, fee)
}

func (s *service) CreatePortfolio(ctx context.Context, userID uint64, name string) (uint64, error) {
	return s.repo.CreatePortfolio(ctx, userID, name)
}

func (s *service) Info(ctx context.Context, userId uint64, portfolioID uint64) (*SvcInfoRes, error) {
	res, err := s.repo.Info(ctx, userId, portfolioID)
	if err != nil {
		return nil, err
	}

	return &SvcInfoRes{Profit: res.Profit}, nil
}
