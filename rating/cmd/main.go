package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/meirongdev/movie-microservice/gen"
	"github.com/meirongdev/movie-microservice/pkg/discovery"
	"github.com/meirongdev/movie-microservice/pkg/discovery/consul"
	"github.com/meirongdev/movie-microservice/rating/internal/controller/rating"
	grpchandler "github.com/meirongdev/movie-microservice/rating/internal/handler/grpc"
	"github.com/meirongdev/movie-microservice/rating/internal/ingester/kafka"
	"github.com/meirongdev/movie-microservice/rating/internal/repository/mysql"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

const serviceName = "rating"

func main() {
	var configPath string
	flag.StringVar(&configPath, "config", "config.yml", "path to the config file")
	flag.Parse()
	config, err := locaConfig(configPath)
	if err != nil {
		panic(err)
	}
	port := config.API.Port
	log.Printf("Starting the rating service on port %d", port)
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
				log.Println("Failed to report healthy state: " + err.Error())
			}
			time.Sleep(1 * time.Second)
		}
	}()
	defer registry.Deregister(ctx, instanceID, serviceName)

	mysqlConfig := config.API.MysqlConfig
	dsn := mysqlConfig.FormatDSN()
	repo, err := mysql.New(dsn)
	if err != nil {
		panic(err)
	}
	kafkaConfig := config.API.KafkaConfig
	ing, err := kafka.NewIngester(kafkaConfig.Address, kafkaConfig.GroupID, kafkaConfig.Topic)
	if err != nil {
		panic(err)
	}
	ctrl := rating.New(repo, rating.WithIngester(ing))
	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Println("recovered from panic: ", r)
			}
		}()
		ingestErr := ctrl.StartIngestion(ctx)
		if ingestErr != nil {
			panic(ingestErr)
		}
	}()
	h := grpchandler.New(ctrl)
	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%v", port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	srv := grpc.NewServer()
	reflection.Register(srv)
	gen.RegisterRatingServiceServer(srv, h)

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

		// TODO DB cleanup
		// TODO Kafka cleanup
	}()
	if err := srv.Serve(lis); err != nil {
		panic(err)
	}
	wg.Wait()
}
