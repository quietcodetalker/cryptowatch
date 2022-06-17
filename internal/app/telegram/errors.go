package telegram

import "errors"

var (
	ErrInternalError     = errors.New("internal error")
	ErrUnexpectedMessage = errors.New("unexpected message")
	ErrUnauthenticated   = errors.New("unauthenticated")
	ErrNotFound          = errors.New("not found")
)
