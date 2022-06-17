package portfolio

import "errors"

var (
	ErrNotFound           = errors.New("not found")
	ErrInternalError      = errors.New("internal error")
	ErrFailedPrecondition = errors.New("failed precondition")
)
