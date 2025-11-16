package trace

import (
	"context"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const DefaultOtlpGrpcTimeout = 30

func NewOtlpGrpcClientConnection(ctx context.Context, host string, dialOptions ...grpc.DialOption) (*grpc.ClientConn, error) {
	dialCtx, cancel := context.WithTimeout(ctx, DefaultOtlpGrpcTimeout*time.Second)
	defer cancel()

	dialContextOptions := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	}

	dialContextOptions = append(dialContextOptions, dialOptions...)

	return grpc.DialContext(dialCtx, host, dialContextOptions...)
}
