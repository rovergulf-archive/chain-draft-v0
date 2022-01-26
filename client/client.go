package client

import (
	"context"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/rovergulf/chain/pkg/traceutil"
	"github.com/rovergulf/chain/rpc"
	"go.uber.org/zap"
)

// NetherClient represents Rovergulf BlockChain Network RPC client interface
type NetherClient struct {
	logger *zap.SugaredLogger
	tracer *traceutil.Tracer
	*ethclient.Client
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
	if c.Client != nil {
		c.Client.Close()
	}
}

func (c *NetherClient) MakeCall(ctx context.Context, req *rpc.ApiRequest) (*rpc.ApiResponse, error) {

	return nil, nil
}
