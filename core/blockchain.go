package core

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/dgraph-io/badger/v3"
	"github.com/ethereum/go-ethereum/common"
	"github.com/opentracing/opentracing-go"
	"github.com/rovergulf/rbn/core/types"
	"github.com/rovergulf/rbn/database/badgerdb"
	"github.com/rovergulf/rbn/params"
	"go.uber.org/zap"
	"io"
)

const (
	DbFileName = "chain.db"
)

var (
	emptyHash = common.HexToHash("")
)

var (
	ErrGenesisNotExists = errors.New("genesis does not exists")
)

type Blockchain struct {
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

func (bc *Blockchain) Shutdown() {
	if bc.db != nil {
		if err := bc.db.Close(); err != nil {
			bc.logger.Errorf("Unable to close db: %s", err)
		}
	}
}

func NewBlockchain(opts params.Options) (*Blockchain, error) {
	opts.Badger = badger.DefaultOptions(opts.DbFilePath)
	db, err := badgerdb.OpenDB(opts.DbFilePath, opts.Badger)
	if err != nil {
		opts.Logger.Errorf("Unable to open db file: %s", err)
		return nil, err
	}

	return &Blockchain{
		LastHash:    common.HexToHash(""),
		ChainLength: 0,
		genesis:     new(Genesis),
		db:          db,
		logger:      opts.Logger,
		tracer:      opts.Tracer,
	}, nil
}

func (bc *Blockchain) Run(ctx context.Context) error {
	if err := bc.loadGenesis(ctx); err != nil {
		return err
	}

	return bc.LoadChainState(ctx)
}

// LoadChainState loads Blockchain state from database
func (bc *Blockchain) LoadChainState(ctx context.Context) error {
	return bc.db.View(func(txn *badger.Txn) error {
		lh, err := txn.Get([]byte("lh"))
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

			lastBlockValue, err := txn.Get(bc.LastHash.Bytes())
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

				bc.LastHash = b.BlockHeader.Hash
				bc.ChainLength = b.Number + 1
				return nil
			})
		})
	})
}

func (bc *Blockchain) DbSize() (int64, int64) {
	return bc.db.Size()
}

// ValidateNextBlock simply validates base block values // TBD made more efficient validation method
func (bc *Blockchain) ValidateNextBlock(next *types.Block) error {
	if bytes.Compare(next.PrevHash.Bytes(), bc.LastHash.Bytes()) != 0 {
		return fmt.Errorf("invalid previous hash: %s", next.PrevHash)
	}
	if bytes.Compare(next.BlockHeader.Hash.Bytes(), emptyHash.Bytes()) == 0 {
		return fmt.Errorf("invalid block hash: %s", next.BlockHeader.Hash)
	}
	if next.Number != bc.ChainLength {
		return fmt.Errorf("invalid block number: %d; expected: %d", next.Number, bc.ChainLength+1)
	}
	return nil
}

// AddBlock adds a block with the provided transactions
func (bc *Blockchain) AddBlock(block *types.Block) error {
	if err := bc.ValidateNextBlock(block); err != nil {
		return err
	}

	blockData, err := block.Serialize()
	if err != nil {
		return err
	}

	return bc.db.Update(func(txn *badger.Txn) error {
		if err := txn.Set(block.BlockHeader.Hash.Bytes(), blockData); err != nil {
			return err
		}

		if err := txn.Set([]byte("lh"), block.BlockHeader.Hash.Bytes()); err != nil {
			bc.logger.Errorf("Unable to set last hash value: %s", err)
			return err
		}

		bc.LastHash = block.BlockHeader.Hash
		bc.ChainLength = block.Number + 1

		bc.logger.Infow("Saved block", "prev", block.PrevHash,
			"hash", block.Hash, "number", block.Number, "txs", len(block.Transactions))

		return nil
	})
}

func (bc *Blockchain) GetBlock(hash common.Hash) (types.Block, error) {
	var block types.Block

	err := bc.db.View(func(txn *badger.Txn) error {
		if item, err := txn.Get(hash.Bytes()); err != nil {
			return fmt.Errorf("block is not found")
		} else {
			return item.Value(func(val []byte) error {
				return block.Deserialize(val)
			})
		}
	})
	if err != nil {
		return block, err
	}

	return block, nil
}

func (bc *Blockchain) GetBlockHashes() ([]common.Hash, error) {
	var blocks []common.Hash

	iter := bc.Iterator()

	for {
		block, err := iter.Next()
		if err != nil {
			return nil, err
		}

		blocks = append(blocks, block.BlockHeader.Hash)

		if bytes.Compare(block.PrevHash.Bytes(), emptyHash.Bytes()) == 0 {
			break
		}
	}

	return blocks, nil
}
