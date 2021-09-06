package client

import (
	"context"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/rovergulf/rbn/node/pb"
	"github.com/rovergulf/rbn/pkg/traceutil"
	"go.uber.org/zap"
)

// NetherClient represents Rovergulf BlockChain Network gRPC client interface
type NetherClient struct {
	host   host.Host
	logger *zap.SugaredLogger
	tracer *traceutil.Tracer
}

func NewClient(ctx context.Context, lg *zap.SugaredLogger, addr string) (*NetherClient, error) {

	return &NetherClient{
		logger: lg,
		host:   nil,
	}, nil
}

func (c *NetherClient) HealthCheck(ctx context.Context) error {

	return nil
}

func (c *NetherClient) Stop() {
	if err := c.host.Close(); err != nil {
		c.logger.Errorf("Unable to close p2p conn: %s", err)
	}
}

func (c *NetherClient) MakeCall(ctx context.Context, cmd pb.Command, ent pb.Entity, req []byte) (*pb.CallResponse, error) {

	return nil, nil
}
