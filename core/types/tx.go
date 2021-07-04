package types

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"github.com/ethereum/go-ethereum/common"
)

var (
	txRewardData = []byte("Treasurer reward")
)

type Transactions []Transaction

// Transaction represents a Bitcoin transaction
type Transaction struct {
	From        common.Address `json:"from" yaml:"from"`
	To          common.Address `json:"to" yaml:"to"` // destination of contract, use empty address for contract creation
	Nonce       uint64         `json:"nonce" yaml:"nonce"`
	Value       uint64         `json:"value" yaml:"value"`
	Nether      uint64         `json:"nether" yaml:"nether"`
	NetherPrice uint64         `json:"nether_price" yaml:"nether_price"`
	Data        []byte         `json:"data" yaml:"data"` // contract data
	Time        int64          `json:"time" yaml:"time"`
}

// Hash returns a hash of the transaction
func (tx *Transaction) Hash() ([]byte, error) {
	txCopy := *tx

	serializedCopy, err := txCopy.Serialize()
	if err != nil {
		return nil, err
	}

	hash := sha256.Sum256(serializedCopy)
	return hash[:], nil
}

// Serialize encodes Transaction with gob encoder
func (tx Transaction) Serialize() ([]byte, error) {
	var encoded bytes.Buffer

	enc := gob.NewEncoder(&encoded)
	if err := enc.Encode(tx); err != nil {
		return nil, err
	}

	return encoded.Bytes(), nil
}

// Deserialize decodes and returns valid Transaction
func (tx *Transaction) Deserialize(data []byte) error {
	decoder := gob.NewDecoder(bytes.NewReader(data))
	return decoder.Decode(tx)
}

func (tx *Transaction) IsReward() bool {
	return bytes.Compare(tx.Data, txRewardData) == 0
}

func (tx *Transaction) Cost() uint64 {
	netherFee := tx.Nether * tx.NetherPrice
	return tx.Value + netherFee
}
