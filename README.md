# Movie microservice

## Description

This is a microservice that provides information about movies. It is part of the movie microservices project.


## Run with Consul as service discovery

```bash
docker run -d -p 8500:8500 -p 8600:8600/udp --name=dev-consul hashicorp/consul agent -server -ui -node=server-1 -bootstrap-expect=1 -client=0.0.0.0
```

can check the consul UI at http://localhost:8500/ui

```bash
go run rating/cmd/main.go 
go run metadata/cmd/main.go 
go run movie/cmd/main.go 
curl -v "localhost:8083/movie?id=1"
```

Movie will call metadata and rating services to get information about the movie.


## Grpcurl to test the service

Installation
```bash
brew install grpcurl
```

```bash
# Server supports reflection
grpcurl -plaintext  localhost:8083 list
# MovieService
# grpc.reflection.v1alpha.ServerReflection
```

test the movie grpc service
```bash
grpcurl -d '{"movie_id": "1"}' -plaintext localhost:8083 MovieService/GetMovieDetails
```

## Run with config file locally
    
```bash
make compose/up
go run metadata/cmd/main.go -config metadata/config/config.yaml
go run rating/cmd/main.go -config rating/config/config.yaml
go run movie/cmd/main.go -config movie/config/config.yaml
grpcurl -d '{"movie_id": "1"}' -plaintext localhost:8083 MovieService/GetMovieDetails
```