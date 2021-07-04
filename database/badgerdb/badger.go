package badgerdb

import (
	"fmt"
	"github.com/dgraph-io/badger/v3"
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
