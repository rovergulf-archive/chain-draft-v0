package client

import (
	"context"
	"github.com/rovergulf/rbn/pkg/traceutil"
	"github.com/rovergulf/rbn/rpc"
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

func (c *NetherClient) MakeCall(ctx context.Context, req *rpc.ApiRequest) (*rpc.ApiResponse, error) {

	return nil, nil
}
