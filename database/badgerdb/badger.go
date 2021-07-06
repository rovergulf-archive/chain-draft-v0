package badgerdb

import (
	"context"
	"fmt"
	"github.com/dgraph-io/badger/v3"
	"github.com/ethereum/go-ethereum/common"
	"github.com/opentracing/opentracing-go"
	"github.com/rovergulf/rbn/database"
	"go.uber.org/zap"
	"os"
	"path/filepath"
	"strings"
)

func retry(dir string, originalOpts badger.Options) (*badger.DB, error) {
	lockPath := filepath.Join(dir, "LOCK")
	if err := os.Remove(lockPath); err != nil {
		return nil, fmt.Errorf("removing 'LOCK': %s", err)
	}
	retryOpts := originalOpts
	retryOpts.BypassLockGuard = true
	return badger.Open(retryOpts)
}

func OpenDB(dir string, opts badger.Options) (*badger.DB, error) {
	opts.Logger = nil
	opts = opts.WithMetricsEnabled(true)
	// TBD calculate available cache
	if db, err := badger.Open(opts); err != nil {
		if strings.Contains(err.Error(), "LOCK") {
			if db, err := retry(dir, opts); err == nil {
				opts.Logger.Debugf("database unlocked, value log truncated")
				return db, nil
			}
			opts.Logger.Errorf("could not unlock database:", err)
		}
		return nil, err
	} else {
		return db, nil
	}
}

var (
	emptyHash = common.HexToHash("")
)

type badgerDb struct {
	db     *badger.DB
	logger *zap.SugaredLogger
	tracer opentracing.Tracer
}

type chainDb struct {
	badgerDb
}

type keystoreDb struct {
	badgerDb
}

type nodeDb struct {
	badgerDb
}

func NewChainDatabase(ctx context.Context, dataDir string) (database.ChainBackend, error) {
	db, err := OpenDB(dataDir, badger.DefaultOptions(dataDir))
	if err != nil {
		return nil, err
	}

	backend := chainDb{
		badgerDb: badgerDb{db: db},
	}

	return &backend, nil
}

func NewKeystoreDatabase(ctx context.Context, dataDir string) (database.KeystoreBackend, error) {
	return nil, nil
}

func NewNodeDatabase(ctx context.Context, dataDir string) (database.NodeBackend, error) {
	return nil, nil
}
