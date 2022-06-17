package user

import (
	"context"
	"cryptowatch/pkg/util/authtoken"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type WithUserID interface {
	GetUserId() uint64
}

func AuthUnaryInterceptor(authtokenMaker authtoken.Maker) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		method := info.FullMethod
		if method == "/cryptowatch.Portfolios/CreatePortfolio" ||
			method == "/cryptowatch.Portfolios/Buy" ||
			method == "/cryptowatch.Portfolios/Sell" ||
			method == "/cryptowatch.Portfolios/Info" ||
			method == "/cryptowatch.Triggers/Add" ||
			method == "/cryptowatch.Triggers/Remove" {

			md, ok := metadata.FromIncomingContext(ctx)
			if !ok {
				return nil, status.New(codes.PermissionDenied, "permission denied").Err()
			}
			values, ok := md["authorization"]
			if !ok || len(values) == 0 {
				return nil, status.New(codes.PermissionDenied, "permission denied").Err()
			}
			token := values[0]
			claims, err := authtokenMaker.VerifyToken(token)
			if err != nil {
				return nil, status.New(codes.PermissionDenied, "permission denied").Err()
			}

			wuid, ok := req.(WithUserID)
			if !ok {
				return nil, status.New(codes.PermissionDenied, "permission denied").Err()
			}
			if wuid.GetUserId() != claims.UserID {
				return nil, status.New(codes.PermissionDenied, "permission denied").Err()
			}
		}

		return handler(ctx, req)
	}
}
