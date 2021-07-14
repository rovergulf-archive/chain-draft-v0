package core

import (
	"context"
	"github.com/dgraph-io/badger/v3"
	"github.com/ethereum/go-ethereum/common"
	"github.com/opentracing/opentracing-go"
	"github.com/rovergulf/rbn/core/types"
	"github.com/rovergulf/rbn/database/badgerdb"
	"github.com/rovergulf/rbn/params"
	"go.uber.org/zap"
	"io"
)

type BlockChain struct {
	LastHash    common.Hash `json:"last_hash" yaml:"last_hash"`
	ChainLength uint64      `json:"chain_length" yaml:"chain_length"`

	genesis *Genesis
	//currentBlock *types.Block

	//mu *sync.RWMutex

	db     *badger.DB
	logger *zap.SugaredLogger
	tracer opentracing.Tracer
	closer io.Closer
}

func (bc *BlockChain) Shutdown() {
	if bc.db != nil {
		if err := bc.db.Close(); err != nil {
			bc.logger.Errorf("Unable to close db: %s", err)
		}
	}
}

func NewBlockChain(opts params.Options) (*BlockChain, error) {
	opts.Badger = badger.DefaultOptions(opts.DbFilePath)
	db, err := badgerdb.OpenDB(opts.DbFilePath, opts.Badger)
	if err != nil {
		opts.Logger.Errorf("Unable to open db file: %s", err)
		return nil, err
	}

	return &BlockChain{
		LastHash:    common.HexToHash(""),
		ChainLength: 0,
		db:          db,
		logger:      opts.Logger,
		tracer:      opts.Tracer,
	}, nil
}

func (bc *BlockChain) Run(ctx context.Context) error {
	if err := bc.loadGenesis(ctx); err != nil {
		return err
	}

	return bc.LoadChainState(ctx)
}

// LoadChainState loads BlockChain state from database
func (bc *BlockChain) LoadChainState(ctx context.Context) error {
	return bc.db.View(func(txn *badger.Txn) error {
		lh, err := txn.Get(lastHashKey)
		if err != nil {
			// is it ok??
			if err == badger.ErrKeyNotFound {
				return nil
			}
			bc.logger.Errorf("Unable to get lastHash: %s", err)
			return err
		}

		return lh.Value(func(val []byte) error {
			bc.LastHash = common.BytesToHash(val)

			key := blockDbPrefix(bc.LastHash)
			lastBlockValue, err := txn.Get(key)
			if err != nil {
				bc.logger.Errorf("Unable to get last block value: %s", err)
				return err
			}

			return lastBlockValue.Value(func(val []byte) error {
				var b types.Block

				if err := b.Deserialize(val); err != nil {
					bc.logger.Errorf("Unable to decode last block value: %s", err)
					return err
				}

				bc.LastHash = b.BlockHeader.BlockHash
				bc.ChainLength = b.Number + 1
				return nil
			})
		})
	})
}

func (bc *BlockChain) DbSize() (int64, int64) {
	return bc.db.Size()
}
