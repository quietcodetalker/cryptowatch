package trigger

import (
	"context"
	"cryptowatch/internal/app/token"
	"log"
)

type Service interface {
	Add(ctx context.Context, userID uint64, ticker string) error
	Remove(ctx context.Context, userID uint64, ticker string) error
	Subcribe(ctx context.Context, userID uint64) chan *token.Token
}

type service struct {
	tokenSvc token.Service
	repo     Repository
}

func NewService(repo Repository, tokenSvc token.Service) *service {
	return &service{
		repo:     repo,
		tokenSvc: tokenSvc,
	}
}

func (s *service) Add(ctx context.Context, userID uint64, ticker string) error {
	_, err := s.tokenSvc.Add(ctx, ticker)
	if err != nil {
		return ErrInternalError
	}
	return s.repo.Add(ctx, userID, ticker)
}

func (s *service) Remove(ctx context.Context, userID uint64, ticker string) error {
	_, err := s.tokenSvc.Add(ctx, ticker)
	if err != nil {
		return ErrInternalError
	}
	return s.repo.Remove(ctx, userID, ticker)
}

func (s *service) Subcribe(ctx context.Context, userID uint64) chan *token.Token {
	out := make(chan *token.Token, 1)
	in := s.tokenSvc.Subscribe(ctx)

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case tkn := <-in:
				ok, err := s.repo.Check(ctx, userID, tkn.Ticker)
				if err != nil {
					log.Printf("err: %v", err)
					continue
				}

				if ok {
					out <- &token.Token{
						Ticker: tkn.Ticker,
						Price:  tkn.Price,
					}
				}
			}
		}
	}()

	return out
}

func (s *service) SubcribeAll(ctx context.Context, userID uint64) chan *token.Token {
	out := make(chan *token.Token, 1)
	in := s.tokenSvc.Subscribe(ctx)

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case tkn := <-in:
				ok, err := s.repo.Check(ctx, userID, tkn.Ticker)
				if err != nil {
					log.Printf("err: %v", err)
					continue
				}

				if ok {
					out <- &token.Token{
						Ticker: tkn.Ticker,
						Price:  tkn.Price,
					}
				}
			}
		}
	}()

	return out
}
