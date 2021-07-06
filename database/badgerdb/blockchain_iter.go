package badgerdb

import (
	"github.com/dgraph-io/badger/v3"
	"github.com/ethereum/go-ethereum/common"
	"github.com/opentracing/opentracing-go"
	"github.com/rovergulf/rbn/core/types"
	"go.uber.org/zap"
)

// chainDbIterator is used to iterate over blockchain blocks
type chainDbIterator struct {
	CurrentHash common.Hash
	db          *badger.DB
	logger      *zap.SugaredLogger
	tracer      opentracing.Tracer
}

// Iterator returns a chainDbIterator
func (bc *chainDb) Iterator(lastHash common.Hash) *chainDbIterator {
	bci := &chainDbIterator{
		CurrentHash: lastHash,
		db:          bc.db,
		logger:      bc.logger,
		tracer:      bc.tracer,
	}

	return bci
}

// Next returns next block starting from the tip
func (i *chainDbIterator) Next() (*types.Block, error) {
	var block *types.Block

	if err := i.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(i.CurrentHash.Bytes())
		if err != nil {
			return err
		}

		return item.Value(func(val []byte) error {
			var nextBlock types.Block
			if err := nextBlock.Deserialize(val); err != nil {
				i.logger.Errorf("Unable to deserialize block: %s", err)
				return err
			} else {
				block = &nextBlock
			}
			return nil
		})
	}); err != nil {
		i.logger.Errorw("Unable to iterate db view",
			"current_hash", i.CurrentHash, "err", err,
		)
		return nil, err
	}

	i.CurrentHash = block.PrevHash

	return block, nil
}
