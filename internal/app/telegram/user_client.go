package telegram

import (
	"context"
	pb "cryptowatch/pkg/api/cryptowatchv1"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/wrapperspb"
	"io"
	"log"
	"time"
)

type UserClient interface {
	GenerateOTP(ctx context.Context, username string) error
	VerifyOTP(ctx context.Context, username string, code string) (*VerifyOTPRes, error)
	Subscribe(ctx context.Context, userID uint64, token string) (chan string, error)
}

type VerifyOTPRes struct {
	UserID uint64 `json:"user_id"`
	Token  string `json:"token"`
}

type userClient struct {
	addr string
}

func NewUserClient(addr string) *userClient {
	return &userClient{
		addr: addr,
	}
}

func (c *userClient) GenerateOTP(ctx context.Context, username string) error {
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}
	conn, err := grpc.DialContext(ctx, c.addr, opts...)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrInternalError, err)
	}
	defer conn.Close()

	client := pb.NewUsersClient(conn)

	_, err = client.GenerateOTP(ctx, &wrapperspb.StringValue{Value: username})
	if err != nil {
		return fmt.Errorf("%w: %v", ErrInternalError, err)
	}

	return nil
}

func (c *userClient) VerifyOTP(ctx context.Context, username string, code string) (*VerifyOTPRes, error) {
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}
	conn, err := grpc.DialContext(ctx, "localhost:50051", opts...)
	if err != nil {
		return nil, ErrInternalError
	}
	defer conn.Close()

	client := pb.NewUsersClient(conn)

	res, err := client.VerifyOTP(ctx, &pb.VerifyOTPReq{
		Username: username,
		Code:     code,
	})
	if err != nil {
		st, ok := status.FromError(err)
		if !ok {
			return nil, ErrInternalError
		}
		switch st.Code() {
		case codes.Unauthenticated:
			return nil, ErrUnauthenticated
		default:
			return nil, ErrInternalError
		}
	}

	return &VerifyOTPRes{
		UserID: res.GetUserId(),
		Token:  res.GetToken(),
	}, nil
}

func (c *userClient) Subscribe(ctx context.Context, userID uint64, token string) (chan string, error) {
	out := make(chan string, 1)

	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}

	md := metadata.New(map[string]string{
		"authorization": token,
	})

	ctx = metadata.NewOutgoingContext(ctx, md)

	conn, err := grpc.DialContext(ctx, "localhost:50051", opts...)
	if err != nil {
		return nil, ErrInternalError
	}

	client := pb.NewTriggersClient(conn)

	stream, err := client.Subscribe(ctx, &wrapperspb.UInt64Value{Value: userID})
	if err != nil {
		conn.Close()
		st, ok := status.FromError(err)
		if !ok {
			return nil, ErrInternalError
		}
		switch st.Code() {
		case codes.Unauthenticated:
			return nil, ErrUnauthenticated
		default:
			return nil, ErrInternalError
		}
	}

	go func() {
		defer conn.Close()
		for {
			time.Sleep(5 * time.Second)
			in, err := stream.Recv()
			log.Printf("RECV: %v", in)
			if err == io.EOF {
				close(out)
				return
			}
			if err != nil {
				log.Printf("recv err: %v", err)
				close(out)
				return
			}

			out <- fmt.Sprintf("%s: $%v", in.GetTicker(), in.GetPrice())
		}
	}()

	return out, nil
}
