package dgraphdb

import (
	"github.com/dgraph-io/dgo/v210"
	"github.com/dgraph-io/dgo/v210/protos/api"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type dgraph struct {
	*dgo.Dgraph

	logger *zap.SugaredLogger
}

// newClient returns dgo.Dgraph gRPC client
func newClient() (*dgo.Dgraph, error) {
	// Dial a gRPC connection. The address to dial to can be configured when
	// setting up the dgraph cluster.
	d, err := grpc.Dial("localhost:9080", grpc.WithInsecure())
	if err != nil {
		return nil, err
	}

	return dgo.NewDgraphClient(
		api.NewDgraphClient(d),
	), nil
}
