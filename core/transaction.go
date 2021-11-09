package core

import (
	"github.com/dgraph-io/badger/v3"
	"github.com/ethereum/go-ethereum/common"
	"github.com/rovergulf/chain/core/types"
)

func (bc *BlockChain) ListTransactions() ([]types.SignedTx, error) {
	var txs []types.SignedTx

	if err := bc.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.Prefix = txsPrefix
		it := txn.NewIterator(opts)
		defer it.Close()
		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			var tx types.SignedTx

			if err := item.Value(func(val []byte) error {
				return tx.Deserialize(val)
			}); err != nil {
				return err
			}

			txs = append(txs, tx)
		}
		return nil
	}); err != nil {
		bc.logger.Errorw("Unable to iterate db view", "err", err)
		return nil, err
	}

	return txs, nil
}

func (bc *BlockChain) FindTransaction(txHash common.Hash) (*types.SignedTx, error) {
	var tx types.SignedTx

	if err := bc.db.View(func(txn *badger.Txn) error {
		key := txDbPrefix(txHash)
		item, err := txn.Get(key)
		if err != nil {
			if err == badger.ErrKeyNotFound {
				return ErrTxNotExists
			}
			return err
		}

		return item.Value(func(val []byte) error {
			return tx.Deserialize(val)
		})
	}); err != nil {
		return nil, err
	}

	return &tx, nil
}

func (bc *BlockChain) SaveTx(txHash common.Hash, tx types.SignedTx) error {
	encodedTx, err := tx.Serialize()
	if err != nil {
		return err
	}

	key := txDbPrefix(txHash)
	return bc.db.Update(func(txn *badger.Txn) error {
		return txn.Set(key, encodedTx)
	})
}
