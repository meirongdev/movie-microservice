package grpc

import (
	"context"

	"github.com/meirongdev/movie-microservice/gen"
	"github.com/meirongdev/movie-microservice/internal/grpcutil"
	"github.com/meirongdev/movie-microservice/metadata/pkg/model"
	"github.com/meirongdev/movie-microservice/pkg/discovery"
	"github.com/meirongdev/movie-microservice/pkg/retry"
)

// Gateway defines a movie metadata gRPC gateway.
type Gateway struct {
	registry discovery.Registry
}

// New creates a new gRPC gateway for a movie metadata service.
func New(registry discovery.Registry) *Gateway {
	return &Gateway{registry}
}

// Get returns movie metadata by a movie id.
func (g *Gateway) Get(ctx context.Context, id string) (*model.Metadata, error) {
	conn, err := grpcutil.ServiceConnection(ctx, "metadata", g.registry)
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	client := gen.NewMetadataServiceClient(conn)

	var resp *gen.GetMetadataResponse
	var grpcErr error
	err = retry.GrpcCall(func() error {
		resp, grpcErr = client.GetMetadata(ctx,
			&gen.GetMetadataRequest{MovieId: id})
		return grpcErr
	})
	if err != nil {
		return nil, err
	}
	return model.MetadataFromProto(resp.Metadata), err
}
