package node

import (
	"context"
	"github.com/rovergulf/rbn/proto"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type Client struct {
	conn *grpc.ClientConn
	lg   *zap.SugaredLogger
	proto.NodeServiceClient
}

func NewClient(ctx context.Context, lg *zap.SugaredLogger, addr string) (*Client, error) {
	conn, err := grpc.Dial(addr, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	c := proto.NewNodeServiceClient(conn)

	// Contact the server and print out its response.

	healthCheck, err := c.Check(ctx, &proto.HealthCheckRequest{Service: viper.GetString("app.name")})
	if err != nil {
		return nil, err
	} else {
		lg.Debugw("Node service health check", "status", healthCheck.Status)
	}

	return &Client{
		conn:              conn,
		lg:                lg,
		NodeServiceClient: c,
	}, nil
}

func (c *Client) Stop() {
	if err := c.conn.Close(); err != nil {
		c.lg.Errorf("Unable to close grpc conn: %s", err)
	}
}
