package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	authv1 "github.com/raibid-labs/mop/examples/02-grpc-service/proto/auth/v1"
	"github.com/raibid-labs/mop/examples/02-grpc-service/internal/service"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	// Initialize logger
	logger, err := zap.NewProduction()
	if err != nil {
		panic(fmt.Sprintf("failed to initialize logger: %v", err))
	}
	defer logger.Sync()

	// Get port from environment or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "9090"
	}

	// Create listener
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", port))
	if err != nil {
		logger.Fatal("failed to listen", zap.Error(err))
	}

	// Create gRPC server
	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			unaryLoggingInterceptor(logger),
		),
		grpc.ChainStreamInterceptor(
			streamLoggingInterceptor(logger),
		),
	)

	// Register auth service
	authService := service.NewAuthService(logger)
	authv1.RegisterAuthServiceServer(grpcServer, authService)

	// Enable reflection for tools like grpcurl
	reflection.Register(grpcServer)

	// Setup graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Start server in goroutine
	go func() {
		logger.Info("gRPC server starting", zap.String("port", port))
		if err := grpcServer.Serve(lis); err != nil {
			logger.Fatal("failed to serve", zap.Error(err))
		}
	}()

	// Wait for shutdown signal
	<-sigChan
	logger.Info("shutting down gRPC server")

	// Graceful shutdown
	stopped := make(chan struct{})
	go func() {
		grpcServer.GracefulStop()
		close(stopped)
	}()

	// Wait for graceful stop or timeout
	select {
	case <-stopped:
		logger.Info("server stopped gracefully")
	case <-time.After(10 * time.Second):
		logger.Warn("server stop timeout, forcing shutdown")
		grpcServer.Stop()
	}
}

// unaryLoggingInterceptor logs unary RPC calls
func unaryLoggingInterceptor(logger *zap.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		start := time.Now()

		// Call the handler
		resp, err := handler(ctx, req)

		// Log the request
		duration := time.Since(start)
		if err != nil {
			logger.Error("gRPC unary call failed",
				zap.String("method", info.FullMethod),
				zap.Duration("duration", duration),
				zap.Error(err),
			)
		} else {
			logger.Info("gRPC unary call succeeded",
				zap.String("method", info.FullMethod),
				zap.Duration("duration", duration),
			)
		}

		return resp, err
	}
}

// streamLoggingInterceptor logs streaming RPC calls
func streamLoggingInterceptor(logger *zap.Logger) grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		start := time.Now()

		// Call the handler
		err := handler(srv, ss)

		// Log the request
		duration := time.Since(start)
		if err != nil {
			logger.Error("gRPC stream call failed",
				zap.String("method", info.FullMethod),
				zap.Duration("duration", duration),
				zap.Error(err),
			)
		} else {
			logger.Info("gRPC stream call succeeded",
				zap.String("method", info.FullMethod),
				zap.Duration("duration", duration),
			)
		}

		return err
	}
}
