package badgerdb

import (
	"bytes"
	"context"
	"fmt"
	"github.com/dgraph-io/badger/v3"
	"github.com/ethereum/go-ethereum/common"
	"github.com/rovergulf/rbn/core/types"
	"github.com/rovergulf/rbn/database"
)

func (bc *chainDb) NewBlockHeader(ctx context.Context, tx *types.BlockHeader) error {
	return nil
}

func (bc *chainDb) GetBlockHeader(ctx context.Context, hash common.Hash) (*types.BlockHeader, error) {
	var block types.BlockHeader
	return &block, nil
}

func (bc *chainDb) SearchBlockHeaders(ctx context.Context) ([]*types.BlockHeader, error) {
	var txs []*types.BlockHeader
	return txs, nil
}

// NewBlock saves a block and last hash in one transaction
func (bc *chainDb) NewBlock(ctx context.Context, block *types.Block) error {
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

		return nil
	})
}

func (bc *chainDb) GetBlock(ctx context.Context, hash common.Hash) (*types.Block, error) {
	var block types.Block

	err := bc.db.View(func(txn *badger.Txn) error {
		if item, err := txn.Get(hash.Bytes()); err != nil {
			if err == badger.ErrKeyNotFound {
				return database.ErrBlockNotExists
			}
			return err
		} else {
			return item.Value(func(val []byte) error {
				return block.Deserialize(val)
			})
		}
	})
	if err != nil {
		return nil, err
	}

	return &block, nil
}

func (bc *chainDb) GetBlockHashes(ctx context.Context, lastHash common.Hash) ([]common.Hash, error) {
	var blocks []common.Hash

	iter := bc.Iterator(lastHash)

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

func (bc *chainDb) SearchBlocks(ctx context.Context) ([]*types.Block, error) {
	var blocks []*types.Block
	return blocks, fmt.Errorf("not implemented")
}
