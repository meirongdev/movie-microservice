package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/meirongdev/movie-microservice/gen"
	"github.com/meirongdev/movie-microservice/metadata/internal/controller/metadata"
	grpchandler "github.com/meirongdev/movie-microservice/metadata/internal/handler/grpc"
	"github.com/meirongdev/movie-microservice/metadata/internal/repository/mysql"
	"go.uber.org/zap"
	"go.uber.org/zap/exp/zapslog"

	"github.com/meirongdev/movie-microservice/pkg/discovery"
	"github.com/meirongdev/movie-microservice/pkg/discovery/consul"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

const serviceName = "metadata"

func main() {
	zapL := zap.Must(zap.NewProduction())
	defer zapL.Sync()

	logger := slog.New(zapslog.NewHandler(zapL.Core(), nil))

	var configPath string
	flag.StringVar(&configPath, "config", "config.yml", "path to the config file")
	flag.Parse()
	config, err := loadConfig(configPath)
	if err != nil {
		panic(err)
	}
	port := config.API.Port

	logger.Info("Starting the metadata service on port", slog.Int("port", port))
	// Register the service in Consul start
	registry, err := consul.NewRegistry("localhost:8500")
	if err != nil {
		panic(err)
	}
	// app level context
	ctx, cancel := context.WithCancel(context.Background())

	instanceID := discovery.GenerateInstanceID(serviceName)
	if err := registry.Register(ctx, instanceID, serviceName, fmt.Sprintf("localhost:%d", port)); err != nil {
		panic(err)
	}
	go func() {
		for {
			if err := registry.ReportHealthyState(instanceID, serviceName); err != nil {
				logger.Error("Failed to report healthy state", slog.Any("error", err))
			}
			time.Sleep(1 * time.Second)
		}
	}()
	defer registry.Deregister(ctx, instanceID, serviceName)
	// Register the service in Consul end
	mysqlConfig := config.API.MysqlConfig
	dsn := mysqlConfig.FormatDSN()
	repo, err := mysql.New(dsn)
	if err != nil {
		panic(err)
	}
	ctrl := metadata.New(repo)
	h := grpchandler.New(ctrl)
	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%v", port))
	if err != nil {
		logger.Error("Failed to listen", slog.Any("error", err))
		os.Exit(1)
	}
	srv := grpc.NewServer()
	reflection.Register(srv)
	gen.RegisterMetadataServiceServer(srv, h)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		s := <-sigChan
		cancel()
		logger.Info("Received signal, attempting graceful shutdown", slog.Any("signal", s))
		srv.GracefulStop()
		logger.Info("Gracefully stopped the gRPC server")

		// TODO DB cleanup
	}()
	if err := srv.Serve(lis); err != nil {
		panic(err)
	}
	wg.Wait()
}
