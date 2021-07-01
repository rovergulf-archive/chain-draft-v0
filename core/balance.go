package core

import (
	"bytes"
	"encoding/gob"
	"github.com/dgraph-io/badger/v3"
	"github.com/ethereum/go-ethereum/common"
)

var (
	balancesPrefix = []byte("balances/")
)

type Balance struct {
	Address common.Address `json:"address" yaml:"address"`
	Balance uint64         `json:"balance" yaml:"balance"`
	Nonce   uint64         `json:"nonce" yaml:"nonce"`
	Symbol  string         `json:"symbol" yaml:"symbol"`
	Units   string         `json:"units" yaml:"units"`
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

func (bc *Blockchain) GetBalance(addr common.Address) (*Balance, error) {
	b, ok := bc.Balances[addr]
	if ok {
		return &b, nil
	}

	var balance Balance

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

func (bc *Blockchain) ListBalances() ([]*Balance, error) {
	var balances []*Balance

	if err := bc.db.View(func(txn *badger.Txn) error {
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
		bc.logger.Errorw("Unable to iterate db view", "err", err)
		return nil, err
	}

	return balances, nil
}

func (bc *Blockchain) GetNextAccountNonce(addr common.Address) uint64 {
	balance, ok := bc.Balances[addr]
	if ok {
		return balance.Balance
	}

	b, err := bc.GetBalance(addr)
	if err != nil {
		bc.logger.Errorw("Unable to get balance")
		return 0
	}

	bc.Balances[addr] = *b

	return b.Nonce
}
