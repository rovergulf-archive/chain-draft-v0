package badgerdb

import (
	"context"
	"fmt"
	"github.com/dgraph-io/badger/v3"
	"github.com/ethereum/go-ethereum/common"
	"github.com/rovergulf/rbn/core/types"
)

var (
	balancesPrefix = []byte("balances/")
)

func getBalanceKey(addr common.Address) []byte {
	return append(balancesPrefix, addr.Bytes()...)
}

func (bc *chainDb) NewBalance(ctx context.Context, balance *types.Balance) error {
	data, err := balance.Serialize()
	if err != nil {
		return err
	}

	key := getBalanceKey(balance.Address)
	if err := bc.db.Update(func(txn *badger.Txn) error {
		if _, err := txn.Get(key); err == nil {
			return fmt.Errorf("'%s' already exists", balance.Address)
		}

		return txn.Set(key, data)
	}); err != nil {
		return err
	}

	return nil
}

func (bc *chainDb) UpdateBalance(ctx context.Context, balance *types.Balance) error {
	data, err := balance.Serialize()
	if err != nil {
		return err
	}

	key := getBalanceKey(balance.Address)
	return bc.db.Update(func(txn *badger.Txn) error {
		return txn.Set(key, data)
	})
}

func (bc *chainDb) GetBalance(ctx context.Context, addr common.Address) (*types.Balance, error) {
	var balance types.Balance

	key := getBalanceKey(addr)
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

func (bc *chainDb) SearchBalances(ctx context.Context) ([]*types.Balance, error) {
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
