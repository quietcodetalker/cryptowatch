package token

import "errors"

var (
	ErrInternalError = errors.New("internal error")
	ErrNotFound      = errors.New("not found")
)
