package portfolio

import (
	"context"
	pb "cryptowatch/pkg/api/cryptowatchv1"
	"errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

type GRPCHandler struct {
	svc Service

	pb.UnimplementedPortfoliosServer
}

func NewGRPCHandler(svc Service) *GRPCHandler {
	return &GRPCHandler{
		svc: svc,
	}
}

func (h *GRPCHandler) CreatePortfolio(ctx context.Context, req *pb.CreatePortfolioReq) (*wrapperspb.UInt64Value, error) {
	id, err := h.svc.CreatePortfolio(ctx, req.GetUserId(), req.GetName())
	if err != nil {
		if errors.Is(err, ErrFailedPrecondition) {
			return nil, status.New(codes.FailedPrecondition, err.Error()).Err()
		}
		return nil, status.New(codes.Internal, err.Error()).Err()
	}

	return &wrapperspb.UInt64Value{Value: id}, status.New(codes.OK, "OK").Err()
}

func (h *GRPCHandler) Buy(ctx context.Context, req *pb.BuySellReq) (*emptypb.Empty, error) {
	err := h.svc.Buy(ctx, req.GetUserId(), req.GetPortfolioId(), req.GetTicker(), req.GetQuantity(), req.GetPrice(), req.GetFee())
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, status.New(codes.NotFound, err.Error()).Err()
		}
		if errors.Is(err, ErrFailedPrecondition) {
			return nil, status.New(codes.FailedPrecondition, err.Error()).Err()
		}
		return nil, status.New(codes.Internal, err.Error()).Err()
	}

	return &emptypb.Empty{}, status.New(codes.OK, "OK").Err()
}

func (h *GRPCHandler) Sell(ctx context.Context, req *pb.BuySellReq) (*emptypb.Empty, error) {
	err := h.svc.Sell(ctx, req.GetUserId(), req.GetPortfolioId(), req.GetTicker(), req.GetQuantity(), req.GetPrice(), req.GetFee())
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, status.New(codes.NotFound, err.Error()).Err()
		}
		if errors.Is(err, ErrFailedPrecondition) {
			return nil, status.New(codes.FailedPrecondition, err.Error()).Err()
		}
		return nil, status.New(codes.Internal, err.Error()).Err()
	}

	return &emptypb.Empty{}, status.New(codes.OK, "OK").Err()
}

func (h *GRPCHandler) Info(ctx context.Context, req *pb.InfoReq) (*pb.InfoRes, error) {
	res, err := h.svc.Info(ctx, req.GetUserId(), req.GetPortfolioId())
	if err != nil {
		return nil, status.New(codes.Internal, err.Error()).Err()
	}

	return &pb.InfoRes{Profit: res.Profit}, status.New(codes.OK, "OK").Err()
}
