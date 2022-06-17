package token

import (
	"context"
	"log"
	"sync"
)

type Service interface {
	Add(ctx context.Context, ticker string) (bool, error)
	Subscribe(ctx context.Context) <-chan *Token
	Start(ctx context.Context) error
}

type service struct {
	repo          Repository
	exchange      Exchange
	updates       chan map[string]float64
	mu            sync.RWMutex
	subscriptions []chan *Token
}

func NewService(repo Repository, exch Exchange) *service {
	return &service{
		repo:     repo,
		exchange: exch,
		updates:  make(chan map[string]float64),
	}
}

func (s *service) Add(ctx context.Context, ticker string) (bool, error) {
	err := s.exchange.Subscribe(ctx, []string{ticker})
	if err != nil {
		return false, err
	}

	ok, err := s.repo.Add(ctx, ticker)
	if err != nil {
		return false, err
	}

	return ok, nil
}

func (s *service) Subscribe(ctx context.Context) <-chan *Token {
	ch := make(chan *Token, 1)
	s.mu.Lock()
	s.subscriptions = append(s.subscriptions, ch)
	s.mu.Unlock()

	go func() {
		<-ctx.Done()
		log.Print("[TOKENS SVC][UNSUBSCRIBE]")
		s.mu.Lock()
		for i, c := range s.subscriptions {
			if c == ch {
				s.subscriptions = append(s.subscriptions[:i], s.subscriptions[i+1:]...)
				break
			}
		}
		s.mu.Unlock()
		close(ch)
	}()

	return ch
}

func (s *service) Start(ctx context.Context) error {
	s.exchange.Start(ctx, s.updates)

	tickers, err := s.repo.ListTickers(ctx)
	if err != nil {
		return err
	}

	err = s.exchange.Subscribe(ctx, tickers)
	if err != nil {
		return err
	}

	go func() {
		for {
			select {
			case update := <-s.updates:
				//log.Printf("update: %v", update)

				for ticker, price := range update {
					err := s.repo.Update(ctx, ticker, price)
					if err != nil {
						log.Printf("err: %v", err)
					}
				}

				s.mu.RLock()
				for _, sub := range s.subscriptions {
					for ticker, price := range update {
						sub <- &Token{
							Ticker: ticker,
							Price:  price,
						}
					}
				}
				s.mu.RUnlock()
			}
		}
	}()

	return nil
}
