package main

import (
	"context"
	"cryptowatch/internal/app/portfolio"
	"cryptowatch/internal/app/telegram"
	"cryptowatch/internal/app/token"
	"cryptowatch/internal/app/trigger"
	"cryptowatch/internal/app/user"
	pb "cryptowatch/pkg/api/cryptowatchv1"
	"cryptowatch/pkg/config"
	"cryptowatch/pkg/util"
	"cryptowatch/pkg/util/authtoken"
	"flag"
	"fmt"
	runtime2 "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	"net"
	"net/http"
	"path"
	"runtime"
	"time"
)

var (
	grpcServerEndpoint = flag.String("grpc-server-endpoint", "localhost:9090", "gRPC server endpoint")
)

func run() error {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Register gRPC server endpoint
	// Note: Make sure the gRPC server is running properly and accessible
	mux := runtime2.NewServeMux()
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	err := pb.RegisterUsersHandlerFromEndpoint(ctx, mux, *grpcServerEndpoint, opts)
	if err != nil {
		return err
	}
	err = pb.RegisterPortfoliosHandlerFromEndpoint(ctx, mux, *grpcServerEndpoint, opts)
	if err != nil {
		return err
	}
	err = pb.RegisterTriggersHandlerFromEndpoint(ctx, mux, *grpcServerEndpoint, opts)
	if err != nil {
		return err
	}

	// Serve HTTP server (and proxy calls to gRPC server endpoint)
	return http.ListenAndServe(":8081", mux)
}

func main() {
	_, filename, _, _ := runtime.Caller(0)
	rootDir := path.Join(path.Dir(filename), "../..")

	cfg, err := config.LoadConfig(
		path.Join(rootDir, "configs"),
		"local",
	)
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	dbSource := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBSSLMode,
	)

	db, err := util.OpenDB(dbSource)
	if err != nil {
		log.Fatalf("failed to open db: %v", err)
	}

	userRepo := user.NewPostgresRepo(db)
	paseto, err := authtoken.NewPasetoMaker(cfg.SymmetricKey)
	if err != nil {
		log.Fatalf("failed to craete paseto token maker: %v", err)
	}
	otpManager := user.NewInMemOTPManager()
	userSvc := user.NewService(userRepo, paseto, otpManager)
	userSrv := user.NewGRPCHandler(userSvc)

	exchange := token.NewCryptoCompareProvider(
		cfg.CryptoCompareToken,
		&http.Client{Timeout: 10 * time.Second},
	)
	tokenRepo := token.NewPostgresRepo(db)
	tokenSvc := token.NewService(tokenRepo, exchange)

	portfolioRepo := portfolio.NewPostgresRepo(db)
	portfolioSvc := portfolio.NewService(portfolioRepo, tokenSvc)
	portfolioSrv := portfolio.NewGRPCHandler(portfolioSvc)

	triggerRepo := trigger.NewPostgresRepo(db)
	triggerSvc := trigger.NewService(triggerRepo, tokenSvc)
	triggerSrv := trigger.NewGRPCHandler(triggerSvc)

	lis, err := net.Listen("tcp", cfg.BindAddr)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	var opts []grpc.ServerOption

	opts = append(opts, grpc.ChainUnaryInterceptor(user.AuthUnaryInterceptor(paseto)))

	grpcServer := grpc.NewServer(opts...)

	pb.RegisterUsersServer(grpcServer, userSrv)
	pb.RegisterPortfoliosServer(grpcServer, portfolioSrv)
	pb.RegisterTriggersServer(grpcServer, triggerSrv)

	userClient := telegram.NewUserClient(cfg.BindAddr)
	tgRepo := telegram.NewPostgresRepo(db)
	tgSvc := telegram.New(cfg.TelegramToken, userClient, tgRepo)
	go func() {
		err := tgSvc.Serve(context.Background())
		if err != nil {
			log.Fatal(err)
		}
	}()

	err = tokenSvc.Start(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		err := run()
		if err != nil {
			log.Fatal(err)
		}
	}()

	log.Println("Listening on " + cfg.BindAddr)
	err = grpcServer.Serve(lis)
	if err != nil {
		log.Fatal(err)
	}
}
