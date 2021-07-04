package core

import (
	"github.com/dgraph-io/badger/v3"
	"github.com/ethereum/go-ethereum/common"
	"github.com/opentracing/opentracing-go"
	"go.uber.org/zap"
)

// BlockchainIterator is used to iterate over blockchain blocks
type BlockchainIterator struct {
	CurrentHash common.Hash
	db          *badger.DB
	logger      *zap.SugaredLogger
	tracer      opentracing.Tracer
}

// Iterator returns a BlockchainIterator
func (bc *Blockchain) Iterator() *BlockchainIterator {
	bci := &BlockchainIterator{
		CurrentHash: bc.LastHash,
		db:          bc.db,
		logger:      bc.logger,
		tracer:      bc.tracer,
	}

	return bci
}

// Next returns next block starting from the tip
func (i *BlockchainIterator) Next() (*Block, error) {
	var block *Block

	if err := i.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(i.CurrentHash.Bytes())
		if err != nil {
			return err
		}

		return item.Value(func(val []byte) error {
			var nextBlock Block
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
