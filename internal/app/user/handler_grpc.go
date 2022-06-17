package user

import (
	"context"
	pb "cryptowatch/pkg/api/cryptowatchv1"
	"errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

type GRPCHandler struct {
	svc Service

	pb.UnimplementedUsersServer
}

func NewGRPCHandler(svc Service) *GRPCHandler {
	return &GRPCHandler{
		svc: svc,
	}
}

func (h *GRPCHandler) CreateUser(ctx context.Context, req *pb.CreateUserReq) (*wrapperspb.UInt64Value, error) {
	u, err := h.svc.Create(ctx, SvcCreateReq{
		Username:  req.GetUsername(),
		Password:  req.GetPassword(),
		FirstName: req.GetFirstName(),
		LastName:  req.GetLastName(),
	})
	if err != nil {
		return nil, ErrToGRPCErr(err)
	}

	return &wrapperspb.UInt64Value{Value: u.ID}, status.New(codes.OK, "OK").Err()
}

func (h *GRPCHandler) Login(ctx context.Context, req *pb.LoginReq) (*wrapperspb.StringValue, error) {
	accessToken, err := h.svc.Login(ctx, SvcLoginReq{
		Username: req.GetUsername(),
		Password: req.GetPassword(),
	})
	if err != nil {
		return nil, ErrToGRPCErr(err)
	}

	return &wrapperspb.StringValue{Value: accessToken}, status.New(codes.OK, "OK").Err()
}

func (h *GRPCHandler) GetUser(ctx context.Context, req *wrapperspb.StringValue) (*pb.User, error) {
	u, err := h.svc.GetByUsername(ctx, req.GetValue())
	if err != nil {
		return nil, ErrToGRPCErr(err)
	}

	return &pb.User{
		Id:         u.ID,
		Username:   u.Username,
		FirstName:  u.FirstName,
		LastName:   u.LastName,
		CreateTime: timestamppb.New(u.CreateTime),
	}, status.New(codes.OK, "OK").Err()
}

func (h *GRPCHandler) GenerateOTP(ctx context.Context, value *wrapperspb.StringValue) (*emptypb.Empty, error) {
	err := h.svc.GenerateOTP(ctx, value.GetValue())
	if err != nil {
		return nil, status.New(codes.Internal, err.Error()).Err()
	}

	return &emptypb.Empty{}, status.New(codes.OK, "OK").Err()
}

func (h *GRPCHandler) GetOTP(ctx context.Context, value *wrapperspb.UInt64Value) (*wrapperspb.StringValue, error) {
	code, err := h.svc.GetOTP(ctx, value.GetValue())
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, status.New(codes.NotFound, err.Error()).Err()
		}
		return nil, status.New(codes.Internal, err.Error()).Err()
	}

	return &wrapperspb.StringValue{Value: code}, status.New(codes.OK, "OK").Err()
}
func (h *GRPCHandler) VerifyOTP(ctx context.Context, req *pb.VerifyOTPReq) (*pb.VerifyOTPRes, error) {
	res, err := h.svc.VerifyOTP(ctx, req.GetUsername(), req.GetCode())
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, status.New(codes.Unauthenticated, err.Error()).Err()
		}
		return nil, status.New(codes.Internal, err.Error()).Err()
	}

	return &pb.VerifyOTPRes{
		UserId: res.UserID,
		Token:  res.Token,
	}, status.New(codes.OK, "OK").Err()
}
