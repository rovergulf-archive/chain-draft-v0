package core

import (
	"bytes"
	"encoding/gob"
	"github.com/dgraph-io/badger/v3"
	"github.com/ethereum/go-ethereum/common"
	"math/big"
)

var (
	balancesPrefix = []byte("balances/")
)

type Balance struct {
	Address    common.Address `json:"address" yaml:"address"`
	Balance    *big.Int       `json:"balance" yaml:"balance"`
	Nonce      uint64         `json:"nonce" yaml:"nonce"`
	PrivateKey []byte         `json:"private_key" yaml:"private_key"`
}

func (b *Balance) Serialize() ([]byte, error) {
	var result bytes.Buffer
	encoder := gob.NewEncoder(&result)
	if err := encoder.Encode(*b); err != nil {
		return nil, err
	}

	return result.Bytes(), nil
}

func (b *Balance) Deserialize(data []byte) error {
	decoder := gob.NewDecoder(bytes.NewReader(data))
	if err := decoder.Decode(b); err != nil {
		return err
	}

	return nil
}

type Balances struct {
	*Blockchain
}

func (b *Balances) GetBalance(addr common.Address) (*Balance, error) {
	var balance Balance

	key := append(balancesPrefix, addr.Bytes()...)
	if err := b.Db.View(func(txn *badger.Txn) error {
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

func (b *Balances) ListBalances() ([]*Balance, error) {
	var balances []*Balance

	if err := b.Db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.Prefix = balancesPrefix
		it := txn.NewIterator(opts)
		defer it.Close()
		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			var balance Balance

			if err := item.Value(func(val []byte) error {
				return balance.Deserialize(val)
			}); err != nil {
				return err
			}

			balances = append(balances, &balance)
		}
		return nil
	}); err != nil {
		b.logger.Errorw("Unable to iterate db view", "err", err)
		return nil, err
	}

	return balances, nil
}
