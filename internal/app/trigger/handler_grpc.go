package trigger

import (
	"context"
	pb "cryptowatch/pkg/api/cryptowatchv1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

type GRPCHandler struct {
	svc Service
	pb.UnimplementedTriggersServer
}

func NewGRPCHandler(svc Service) *GRPCHandler {
	return &GRPCHandler{
		svc: svc,
	}
}

func (h *GRPCHandler) Add(ctx context.Context, req *pb.Req) (*emptypb.Empty, error) {
	err := h.svc.Add(ctx, req.GetUserId(), req.GetTicker())
	if err != nil {
		return nil, status.New(codes.Internal, err.Error()).Err()
	}

	return &emptypb.Empty{}, status.New(codes.OK, "OK").Err()
}

func (h *GRPCHandler) Remove(ctx context.Context, req *pb.Req) (*emptypb.Empty, error) {
	err := h.svc.Remove(ctx, req.GetUserId(), req.GetTicker())
	if err != nil {
		return nil, status.New(codes.Internal, err.Error()).Err()
	}

	return &emptypb.Empty{}, status.New(codes.OK, "OK").Err()
}

func (h *GRPCHandler) Subscribe(value *wrapperspb.UInt64Value, server pb.Triggers_SubscribeServer) error {
	ch := h.svc.Subcribe(server.Context(), value.GetValue())

	for tkn := range ch {
		err := server.Send(&pb.Token{
			Ticker: tkn.Ticker,
			Price:  tkn.Price,
		})
		if err != nil {
			return status.New(codes.Internal, err.Error()).Err()
		}
	}

	return status.New(codes.OK, "OK").Err()
}
