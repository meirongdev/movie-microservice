package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/meirongdev/movie-microservice/gen"
	"github.com/meirongdev/movie-microservice/pkg/config"
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
	var port int
	flag.IntVar(&port, "port", 8082, "API handler port")
	flag.Parse()
	log.Printf("Starting the rating service on port %d", port)
	registry, err := consul.NewRegistry("localhost:8500")
	if err != nil {
		panic(err)
	}
	ctx := context.Background()
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

	mysqlConfig := config.MySQLConfig{
		Host:     "localhost:3306",
		Username: "test",
		Password: "test",
		Database: "moviedb",
	}
	dsn := mysqlConfig.FormatDSN()
	repo, err := mysql.New(dsn)
	if err != nil {
		panic(err)
	}
	ing, err := kafka.NewIngester("localhost:9092", "rating", "ratings")
	if err != nil {
		panic(err)
	}
	ctrl := rating.New(repo, rating.WithIngester(ing))
	func() {
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
	err = ctrl.StartIngestion(ctx)
	if err != nil {
		panic(err)
	}
	h := grpchandler.New(ctrl)
	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%v", port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	srv := grpc.NewServer()
	reflection.Register(srv)
	gen.RegisterRatingServiceServer(srv, h)
	if err := srv.Serve(lis); err != nil {
		panic(err)
	}
}
