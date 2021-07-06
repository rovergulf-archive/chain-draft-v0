package dgraphdb

import (
	"github.com/dgraph-io/dgo/v210"
	"github.com/dgraph-io/dgo/v210/protos/api"
	"github.com/opentracing/opentracing-go"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type dgraphDb struct {
	*dgo.Dgraph

	logger *zap.SugaredLogger
	tracer opentracing.Tracer
}

// newClient returns dgo.Dgraph gRPC client
func newClient() (*dgo.Dgraph, error) {
	// Dial a gRPC connection. The address to dial to can be configured when
	// setting up the dgraph cluster.
	d, err := grpc.Dial(viper.GetString("dgraph.host"), grpc.WithInsecure())
	if err != nil {
		return nil, err
	}

	return dgo.NewDgraphClient(
		api.NewDgraphClient(d),
	), nil
}

type chainDb struct {
	dgraphDb
}
