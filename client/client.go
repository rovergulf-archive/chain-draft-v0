package client

import (
	"context"
	"github.com/rovergulf/rbn/node/pb"
	"github.com/rovergulf/rbn/pkg/traceutil"
	"go.uber.org/zap"
)

// NetherClient represents Rovergulf BlockChain Network gRPC client interface
type NetherClient struct {
	logger *zap.SugaredLogger
	tracer *traceutil.Tracer
}

func NewClient(ctx context.Context, lg *zap.SugaredLogger, addr string) (*NetherClient, error) {

	return &NetherClient{
		logger: lg,
	}, nil
}

func (c *NetherClient) HealthCheck(ctx context.Context) error {

	return nil
}

func (c *NetherClient) Stop() {
	// TBD
}

func (c *NetherClient) MakeCall(ctx context.Context, cmd pb.Command, ent pb.Entity, req []byte) (*pb.CallResponse, error) {

	return nil, nil
}
