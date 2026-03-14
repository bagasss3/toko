package console

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/bagasss3/toko/packages/cache"
	pkgconfig "github.com/bagasss3/toko/packages/config"
	"github.com/bagasss3/toko/packages/database"
	"github.com/bagasss3/toko/packages/logger"
	pb "github.com/bagasss3/toko/services/auth-service/pb/auth"

	"github.com/bagasss3/toko/services/auth-service/internal/config"
	grpchandler "github.com/bagasss3/toko/services/auth-service/internal/delivery/handler"
	"github.com/bagasss3/toko/services/auth-service/internal/helper/token"
	"github.com/bagasss3/toko/services/auth-service/internal/repository"
	"github.com/bagasss3/toko/services/auth-service/internal/usecase"
)

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Start gRPC server",
	Run:   runServer,
}

func runServer(cmd *cobra.Command, args []string) {
	log := logger.NewEntry("auth-service")

	pkgconfig.Init()
	cfg := config.Load()

	db, err := database.Init(database.Config{
		DSN:             fmt.Sprintf("postgresql://%s:%s@%s/%s", cfg.DBUser, cfg.DBPassword, cfg.DBHost, cfg.DBName),
		MaxIdleConns:    3,
		MaxOpenConns:    15,
		ConnMaxLifetime: 30 * time.Minute,
		ConnMaxIdleTime: 10 * time.Minute,
		PingInterval:    10 * time.Second,
		RetryAttempts:   5,
	})
	if err != nil {
		log.WithError(err).Fatal("failed to init database")
	}
	defer db.Close()

	keeper := cache.NewKeeper(cache.Config{
		Host:       cfg.RedisHost,
		Password:   cfg.RedisPassword,
		DB:         cfg.RedisDB,
		DefaultTTL: cfg.RedisExpired,
		MaxIdle:    100,
		MaxActive:  10000,
	})

	tokenMaker, err := token.NewMaker(cfg.PasetoKey)
	if err != nil {
		log.WithError(err).Fatal("failed to create token maker")
	}

	// Initialize transactioner
	transactioner := repository.NewGormTransactioner(db)

	// Initialize repositories
	userRepo := repository.NewUserRepository(db, keeper, cfg.RedisExpired)
	addressRepo := repository.NewAddressRepository(db)

	// Initialize usecases
	authUsecase := usecase.NewAuthUsecase(userRepo, tokenMaker, cfg)
	userUsecase := usecase.NewUserUsecase(userRepo, addressRepo, transactioner)
	addressUsecase := usecase.NewAddressUsecase(addressRepo, transactioner)

	// Create first admin if none exists
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	if err := userUsecase.CreateFirstAdmin(ctx); err != nil {
		log.WithError(err).Warn("failed to create first admin")
	}
	cancel()

	// Initialize handler with all usecases
	authGRPCHandler := grpchandler.NewAuthGRPCHandler(authUsecase, userUsecase, addressUsecase)

	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", cfg.Port))
	if err != nil {
		log.WithError(err).Fatal("failed to listen")
	}

	grpcServer := grpc.NewServer()
	pb.RegisterAuthServiceServer(grpcServer, authGRPCHandler)
	reflection.Register(grpcServer)

	go func() {
		log.Infof("auth-service gRPC listening on :%s", cfg.Port)
		if err := grpcServer.Serve(lis); err != nil {
			log.WithError(err).Fatal("gRPC server failed")
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("shutting down gracefully...")

	stopped := make(chan struct{})
	go func() {
		grpcServer.GracefulStop()
		close(stopped)
	}()

	select {
	case <-stopped:
		log.Info("gRPC server stopped gracefully")
	case <-time.After(10 * time.Second):
		grpcServer.Stop()
		log.Warn("gRPC server force stopped after timeout")
	}

	log.Info("auth-service exited")
}
