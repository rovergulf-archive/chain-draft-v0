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
	if bytes.Compare(next.BlockHeader.Hash.Bytes(), emptyHash.Bytes()) == 0 {
		return fmt.Errorf("invalid block hash: %s", next.BlockHeader.Hash)
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

func (bc *BlockChain) GetBlock(hash common.Hash) (types.Block, error) {
	var block types.Block

	err := bc.db.View(func(txn *badger.Txn) error {
		if item, err := txn.Get(hash.Bytes()); err != nil {
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

func (bc *BlockChain) GetBlockHashes() ([]common.Hash, error) {
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
