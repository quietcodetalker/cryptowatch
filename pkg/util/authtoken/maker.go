package authtoken

import "time"

// Maker is an interface for managing tokens.
type Maker interface {
	// CreateToken creates a new token for a specific user id and duration.
	CreateToken(userID uint64, duration time.Duration) (string, error)

	// VerifyToken checks if the provided token is valid or not.
	VerifyToken(token string) (*Payload, error)
}
