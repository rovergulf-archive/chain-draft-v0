package client

import (
	"context"
	"github.com/rovergulf/rbn/pkg/traceutil"
	"github.com/rovergulf/rbn/proto"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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

func (c *NetherClient) HealthCheck(ctx context.Context) error {
	healthCheck, err := c.rpcClient.Check(ctx, &proto.HealthCheckRequest{Service: viper.GetString("app.name")})
	if err != nil {
		return err
	} else {
		c.logger.Debugw("Node service health check", "status", healthCheck.Status)
	}
	return nil
}

func (c *NetherClient) Stop() {
	if err := c.conn.Close(); err != nil {
		c.logger.Errorf("Unable to close grpc conn: %s", err)
	}
}

func (c *NetherClient) MakeCall(ctx context.Context, cmd proto.Command, ent proto.Entity, req []byte) (*proto.CallResponse, error) {
	res, err := c.rpcClient.RpcCall(ctx, &proto.CallRequest{
		Cmd:    cmd,
		Entity: ent,
		Data:   req,
	})
	if err != nil {
		return nil, err
	}

	c.logger.Debugf("Res status: %d (0 = Success!), Data len: %d", res.Status, len(res.Data))

	if res.Status != 0 {
		return nil, status.Errorf(codes.Code(res.Status), string(res.Data))
	}

	return res, nil
}
