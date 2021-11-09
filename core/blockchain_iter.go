package core

import (
	"github.com/dgraph-io/badger/v3"
	"github.com/ethereum/go-ethereum/common"
	"github.com/opentracing/opentracing-go"
	"github.com/rovergulf/chain/core/types"
	"go.uber.org/zap"
)

// BlockChainIterator is used to iterate over blockchain blocks
type BlockChainIterator struct {
	CurrentHash common.Hash
	db          *badger.DB
	logger      *zap.SugaredLogger
	tracer      opentracing.Tracer
}

// Iterator returns a BlockChainIterator
func (bc *BlockChain) Iterator() *BlockChainIterator {
	bci := &BlockChainIterator{
		CurrentHash: bc.LastHash,
		db:          bc.db,
		logger:      bc.logger,
		tracer:      bc.tracer,
	}

	return bci
}

// Next returns next block starting from the tip
func (i *BlockChainIterator) Next() (*types.Block, error) {
	var block *types.Block

	if err := i.db.View(func(txn *badger.Txn) error {
		key := blockDbPrefix(i.CurrentHash)
		item, err := txn.Get(key)
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
