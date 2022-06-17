package user

import (
	"errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	ErrInternalError      = errors.New("internal error")
	ErrInvalidArgument    = errors.New("invalid argument")
	ErrFailedPrecondition = errors.New("failed precondition")
	ErrNotFound           = errors.New("not found")
	ErrUnauthenticated    = errors.New("unauthenticated")
)

func ErrToGRPCErr(err error) error {
	switch {
	case errors.Is(err, ErrInternalError):
		return status.New(codes.Internal, err.Error()).Err()
	case errors.Is(err, ErrInvalidArgument):
		return status.New(codes.InvalidArgument, err.Error()).Err()
	case errors.Is(err, ErrFailedPrecondition):
		return status.New(codes.FailedPrecondition, err.Error()).Err()
	case errors.Is(err, ErrNotFound):
		return status.New(codes.NotFound, err.Error()).Err()
	case errors.Is(err, ErrUnauthenticated):
		return status.New(codes.Unauthenticated, err.Error()).Err()
	default:
		return status.New(codes.Unknown, err.Error()).Err()
	}
}
