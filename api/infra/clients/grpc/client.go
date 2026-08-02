package grpc

import (
	"github.com/liraraphael/go-framework-bench/api/core/domain/enums"
	"github.com/liraraphael/go-framework-bench/api/core/domain/errors"
	"github.com/liraraphael/go-framework-bench/api/core/domain/requests"
	"github.com/liraraphael/go-framework-bench/api/core/domain/responses"
	"github.com/liraraphael/go-framework-bench/api/core/ports"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

type grpcClient struct {
	conn *grpc.ClientConn
}

func NewGrpcClient(target string) (ports.Client, error) {
	conn, err := grpc.Dial(target, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	return &grpcClient{conn: conn}, nil
}

func (c *grpcClient) Do(ctx ports.Context, method enums.HttpMethod, url string, req requests.Request[any, any]) (responses.Response[any, any], error) {
	if req.Headers() != nil {
		md := metadata.New(nil)
		for k, v := range req.Headers() {
			for _, val := range v {
				md.Append(k, val)
			}
		}
		ctx = metadata.NewOutgoingContext(ctx, md)
	}

	return nil, &errors.ClientError{
		Message:    "generic gRPC Do not fully implemented - requires service-specific invocation",
		StatusCode: 501,
	}
}

func (c *grpcClient) Close() error {
	return c.conn.Close()
}
