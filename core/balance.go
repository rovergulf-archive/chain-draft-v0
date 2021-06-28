package core

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"github.com/dgraph-io/badger/v3"
	"github.com/ethereum/go-ethereum/common"
	"math/big"
)

var (
	balancesPrefix = []byte("balances/")
)

type Balance struct {
	Address common.Address `json:"address" yaml:"address"`
	Balance *big.Int       `json:"balance" yaml:"balance"`
	Nonce   uint64         `json:"nonce" yaml:"nonce"`
	//Storage    map[common.Hash]common.Hash `json:"storage" yaml:"storage"`
	//Code       []byte                      `json:"code" yaml:"code"`
	//PrivateKey []byte                      `json:"private_key" yaml:"private_key"`
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

func (b *Balance) MarshalJSON() ([]byte, error) {
	return json.Marshal(*b)
}

func (b *Balance) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, b)
}

type Balances struct {
	Blockchain
}

func (bc *Balances) GetBalance(addr common.Address) (*Balance, error) {
	var b Balance

	key := append(balancesPrefix, addr.Bytes()...)
	if err := bc.Db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(key)
		if err != nil {
			return err
		}

		return item.Value(func(val []byte) error {
			return b.Deserialize(val)
		})
	}); err != nil {
		return nil, err
	}

	return &b, nil
}
