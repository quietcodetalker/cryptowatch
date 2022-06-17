package user

import (
	"context"
	"log"
	"math/rand"
	"strings"
	"sync"
	"time"
)

const (
	codeLen = 6
)

func init() {
	rand.Seed(time.Now().Unix())
}

type OTPManager interface {
	Add(ctx context.Context, userID uint64) error
	Get(ctx context.Context, userID uint64) (string, error)
	Verify(ctx context.Context, userID uint64, code string) error
}

type inMemOTPManager struct {
	codes map[uint64]string
	mu    sync.RWMutex
}

func NewInMemOTPManager() *inMemOTPManager {
	return &inMemOTPManager{
		codes: make(map[uint64]string),
	}
}

func (m *inMemOTPManager) Add(ctx context.Context, userID uint64) error {
	m.mu.Lock()

	code := m.generateCode()

	m.codes[userID] = code

	log.Printf("OTP: %v", m.codes)

	m.mu.Unlock()

	return nil
}

func (m *inMemOTPManager) Get(ctx context.Context, userID uint64) (string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	code, ok := m.codes[userID]
	if !ok {
		return "", ErrNotFound
	}

	return code, nil
}

func (m *inMemOTPManager) Verify(ctx context.Context, userID uint64, code string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	c, ok := m.codes[userID]
	if !ok {
		return ErrNotFound
	}

	if c != code {
		return ErrNotFound
	}

	delete(m.codes, userID)
	return nil
}

var digits = "0123456789"

func (m *inMemOTPManager) generateCode() string {
	var result strings.Builder
	for i := 0; i < codeLen; i++ {
		j := rand.Intn(len(digits))
		result.WriteByte(digits[j])
	}

	return result.String()
}
