package core

import (
	"context"
	"github.com/dgraph-io/badger/v3"
	"github.com/ethereum/go-ethereum/common"
	"github.com/rovergulf/rbn/core/types"
)

func (bc *BlockChain) SaveReceipt(ctx context.Context, receipt *types.Receipt) error {
	data, err := receipt.Serialize()
	if err != nil {
		return err
	}

	key := receiptDbPrefix(receipt.TxHash)
	return bc.db.Update(func(txn *badger.Txn) error {
		return txn.Set(key, data)
	})
}

func (bc *BlockChain) GetReceipt(ctx context.Context, txHash common.Hash) (*types.Receipt, error) {
	var receipt types.Receipt

	key := receiptDbPrefix(txHash)
	if err := bc.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(key)
		if err != nil {
			if err == badger.ErrKeyNotFound {
				return ErrReceiptNotExists
			}
			return err
		}

		return item.Value(func(val []byte) error {
			return receipt.Deserialize(val)
		})
	}); err != nil {
		return nil, err
	}

	return &receipt, nil
}

func (bc *BlockChain) ListReceipts(ctx context.Context) ([]*types.Receipt, error) {
	var res []*types.Receipt

	if err := bc.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.Prefix = receiptsPrefix
		it := txn.NewIterator(opts)
		defer it.Close()
		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			var r types.Receipt

			if err := item.Value(func(val []byte) error {
				return r.Deserialize(val)
			}); err != nil {
				return err
			}

			res = append(res, &r)
		}
		return nil
	}); err != nil {
		bc.logger.Errorw("Unable to iterate db view", "err", err)
		return nil, err
	}

	return res, nil
}
