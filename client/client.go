package client

import (
	"context"
	"github.com/rovergulf/rbn/pkg/traceutil"
	"github.com/rovergulf/rbn/proto"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

// NetherClient represents Rovergulf BlockChain Network gRPC client interface
type NetherClient struct {
	conn      *grpc.ClientConn
	rpcClient proto.NodeServiceClient
	logger    *zap.SugaredLogger
	tracer    *traceutil.Tracer
}

func NewClient(ctx context.Context, lg *zap.SugaredLogger, addr string) (*NetherClient, error) {
	conn, err := grpc.Dial(addr, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		return nil, err
	}

	c := proto.NewNodeServiceClient(conn)

	// Contact the server and print out its response.

	healthCheck, err := c.Check(ctx, &proto.HealthCheckRequest{Service: viper.GetString("app.name")})
	if err != nil {
		return nil, err
	} else {
		lg.Debugw("Node service health check", "status", healthCheck.Status)
	}

	return &NetherClient{
		conn:      conn,
		logger:    lg,
		rpcClient: c,
	}, nil
}

func (c *NetherClient) Stop() {
	if err := c.conn.Close(); err != nil {
		c.logger.Errorf("Unable to close grpc conn: %s", err)
	}
}
