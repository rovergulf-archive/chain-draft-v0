package core

import (
	"github.com/dgraph-io/badger/v3"
	"github.com/ethereum/go-ethereum/common"
	"github.com/rovergulf/rbn/core/types"
)

var (
	balancesPrefix = []byte("balances/")
)

func (bc *Blockchain) GetBalance(addr common.Address) (*types.Balance, error) {
	var balance types.Balance

	key := append(balancesPrefix, addr.Bytes()...)
	if err := bc.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(key)
		if err != nil {
			return err
		}

		return item.Value(func(val []byte) error {
			return balance.Deserialize(val)
		})
	}); err != nil {
		return nil, err
	}

	return &balance, nil
}

func (bc *Blockchain) ListBalances() ([]*types.Balance, error) {
	var balances []*types.Balance

	if err := bc.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.Prefix = balancesPrefix
		it := txn.NewIterator(opts)
		defer it.Close()
		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			var balance types.Balance

			if err := item.Value(func(val []byte) error {
				return balance.Deserialize(val)
			}); err != nil {
				return err
			}

			balances = append(balances, &balance)
		}
		return nil
	}); err != nil {
		bc.logger.Errorw("Unable to iterate db view", "err", err)
		return nil, err
	}

	return balances, nil
}

func (bc *Blockchain) GetNextAccountNonce(addr common.Address) uint64 {
	b, err := bc.GetBalance(addr)
	if err != nil {
		if err != badger.ErrKeyNotFound {
			bc.logger.Errorw("Unable to get balance: %s", err)
		}
		return 0
	}

	return b.Nonce + 1
}
