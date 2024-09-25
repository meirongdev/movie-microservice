package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"time"

	"github.com/grpc-ecosystem/go-grpc-middleware/ratelimit"
	"github.com/meirongdev/movie-microservice/gen"
	"github.com/meirongdev/movie-microservice/movie/internal/controller/movie"
	metadatagateway "github.com/meirongdev/movie-microservice/movie/internal/gateway/metadata/grpc"
	ratinggateway "github.com/meirongdev/movie-microservice/movie/internal/gateway/rating/grpc"
	grpchandler "github.com/meirongdev/movie-microservice/movie/internal/handler/grpc"
	"github.com/meirongdev/movie-microservice/pkg/discovery"
	"github.com/meirongdev/movie-microservice/pkg/discovery/consul"
	"github.com/meirongdev/movie-microservice/pkg/limiter"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

const serviceName = "movie"

func main() {
	var configPath string
	flag.StringVar(&configPath, "config", "config.yml", "path to the config file")
	flag.Parse()
	config, err := loadConfig(configPath)
	if err != nil {
		panic(err)
	}
	port := config.API.Port

	// Register with Consul start
	registry, err := consul.NewRegistry("localhost:8500")
	if err != nil {
		panic(err)
	}

	// app level context
	ctx, cancel := context.WithCancel(context.Background())

	instanceID := discovery.GenerateInstanceID((serviceName))
	if err := registry.Register(ctx, instanceID, serviceName, "localhost:"+strconv.Itoa(port)); err != nil {
		panic(err)
	}

	go func() {
		for {
			if err := registry.ReportHealthyState(instanceID, serviceName); err != nil {
				log.Println("failed to report healthy state:", err)
			}
			time.Sleep(2 * time.Second)
		}
	}()
	defer registry.Deregister(ctx, instanceID, serviceName)
	// Register with Consul end

	metadataGateway := metadatagateway.New(registry)
	ratingGateway := ratinggateway.New(registry)
	ctrl := movie.New(ratingGateway, metadataGateway)
	h := grpchandler.New(ctrl)
	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	const limit = 100
	const burst = 100
	l := limiter.New(limit, burst)
	srv := grpc.NewServer(
		grpc.UnaryInterceptor(ratelimit.UnaryServerInterceptor(l)),
	)
	reflection.Register(srv)
	gen.RegisterMovieServiceServer(srv, h)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		s := <-sigChan
		cancel()
		log.Printf("Received signal %v, attempting graceful shutdown", s)
		srv.GracefulStop()
		log.Println("Gracefully stopped the gRPC server")
	}()
	if err := srv.Serve(lis); err != nil {
		panic(err)
	}
	wg.Wait()
}
