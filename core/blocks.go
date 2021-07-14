package core

import (
	"bytes"
	"fmt"
	"github.com/dgraph-io/badger/v3"
	"github.com/ethereum/go-ethereum/common"
	"github.com/rovergulf/rbn/core/types"
)

// ValidateNextBlock simply validates base block values // TBD made more efficient validation method
func (bc *BlockChain) ValidateNextBlock(next *types.Block) error {
	if bytes.Compare(next.PrevHash.Bytes(), bc.LastHash.Bytes()) != 0 {
		return fmt.Errorf("invalid previous hash: %s", next.PrevHash)
	}
	if IsHashEmpty(next.Root) {
		return fmt.Errorf("invalid root hash: %s", next.BlockHeader.BlockHash)
	}
	if IsHashEmpty(next.BlockHash) {
		return fmt.Errorf("invalid block hash: %s", next.BlockHeader.BlockHash)
	}
	if next.Number != bc.ChainLength {
		return fmt.Errorf("invalid block number: %d; expected: %d", next.Number, bc.ChainLength+1)
	}
	return nil
}

// AddBlock adds a block with the provided transactions
func (bc *BlockChain) AddBlock(block *types.Block) error {
	if err := bc.ValidateNextBlock(block); err != nil {
		return err
	}

	blockData, err := block.Serialize()
	if err != nil {
		return err
	}

	key := blockDbPrefix(block.BlockHeader.BlockHash)
	numKey := blockNumDbPrefix(block.Number)
	hashValue := block.BlockHeader.BlockHash.Bytes()
	return bc.db.Update(func(txn *badger.Txn) error {
		if err := txn.Set(numKey, hashValue); err != nil {
			return err
		}

		if err := txn.Set(key, blockData); err != nil {
			return err
		}

		if err := txn.Set(lastHashKey, hashValue); err != nil {
			bc.logger.Errorf("Unable to set last hash value: %s", err)
			return err
		}

		bc.LastHash = block.BlockHeader.BlockHash
		bc.ChainLength = block.Number + 1

		bc.logger.Infow("Saved block", "prev", block.PrevHash,
			"hash", hashValue, "number", block.Number, "txs", len(block.Transactions))

		return nil
	})
}

func (bc *BlockChain) GetBlock(hash common.Hash) (types.Block, error) {
	var block types.Block

	key := blockDbPrefix(hash)
	err := bc.db.View(func(txn *badger.Txn) error {
		if item, err := txn.Get(key); err != nil {
			if err == badger.ErrKeyNotFound {
				return ErrBlockNotExists
			}
			return err
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

func (bc *BlockChain) GetBlockByNumber(blockNumber uint64) (*types.Block, error) {
	var blockHash common.Hash
	numKey := blockNumDbPrefix(blockNumber)
	err := bc.db.View(func(txn *badger.Txn) error {
		if item, err := txn.Get(numKey); err != nil {
			if err == badger.ErrKeyNotFound {
				return ErrBlockNotExists
			}
			return err
		} else {
			return item.Value(func(val []byte) error {
				blockHash = common.BytesToHash(val)
				return nil
			})
		}
	})
	if err != nil {
		return nil, err
	}

	block, err := bc.GetBlock(blockHash)
	if err != nil {
		return nil, err
	}

	return &block, nil
}

func (bc *BlockChain) SearchBlocks() ([]*types.Block, error) {
	var blocks []*types.Block

	if err := bc.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.Prefix = blocksPrefix
		it := txn.NewIterator(opts)
		defer it.Close()
		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			var b types.Block

			if err := item.Value(func(val []byte) error {
				return b.Deserialize(val)
			}); err != nil {
				return err
			}

			blocks = append(blocks, &b)
		}
		return nil
	}); err != nil {
		bc.logger.Errorw("Unable to iterate db view", "err", err)
		return nil, err
	}

	return blocks, nil
}
