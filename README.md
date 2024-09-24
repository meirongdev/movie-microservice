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