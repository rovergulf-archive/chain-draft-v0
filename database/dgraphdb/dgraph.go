package dgraphdb

import (
	"context"
	"github.com/dgraph-io/dgo/v210"
	"github.com/dgraph-io/dgo/v210/protos/api"
	"github.com/opentracing/opentracing-go"
	"github.com/rovergulf/rbn/params"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type dgraphDb struct {
	*dgo.Dgraph
	conn *grpc.ClientConn

	logger *zap.SugaredLogger
	tracer opentracing.Tracer
}

func (d *dgraphDb) Shutdown() {
	if d.conn != nil {
		if err := d.conn.Close(); err != nil {
			d.logger.Errorf("Unable to close dgraph db conn: %s", err)
		}
	}
}

// NewClient returns dgo.Dgraph gRPC client
func NewClient() (*dgo.Dgraph, error) {
	// Dial a gRPC connection. The address to dial to can be configured when
	// setting up the dgraph cluster.
	d, err := grpc.Dial(viper.GetString("dgraph.host"), grpc.WithInsecure())
	if err != nil {
		// log error
		return nil, err
	}

	// TODO setup tls (if specified? or it must be required?)
	// tlsConf, err := setupTLS()
	// if err != nil {}

	return dgo.NewDgraphClient(
		api.NewDgraphClient(d),
	), nil
}

// setupTLS runs TLS connection with dgo.Client
func setupTLS() {
	// TBD
}

type chainDb struct {
	dgraphDb
}

type keystoreDb struct {
	dgraphDb
}

type nodeDb struct {
	dgraphDb
}

func newDgraphClient(ctx context.Context, opts params.Options) (*dgraphDb, error) {
	d, err := NewClient()
	if err != nil {
		return nil, err
	}

	db := &dgraphDb{
		Dgraph: d,
		logger: opts.Logger,
		tracer: opts.Tracer,
	}

	return db, nil
}

func NewChainDatabase(ctx context.Context, opts params.Options) (*chainDb, error) {
	return nil, nil
}

func NewKeystoreDatabase(ctx context.Context, opts params.Options) (*keystoreDb, error) {
	return nil, nil
}

func NewNodeDatabase(ctx context.Context, opts params.Options) (*nodeDb, error) {
	return nil, nil
}
