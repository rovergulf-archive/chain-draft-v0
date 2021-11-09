package types

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/rovergulf/chain/params"
	"time"
)

var (
	TxRewardData = []byte("Treasurer reward")
)

type Transactions []Transaction

// NewTransaction creates a new transaction
func NewTransaction(from, to common.Address, amount uint64, nonce uint64, data []byte) (Transaction, error) {
	if from == to {
		return Transaction{}, fmt.Errorf("transaction cannot be sent to yourself")
	}

	percentile := params.Coin / (params.TxPrice * params.NetherPrice)
	nether := amount / percentile
	if nether < params.NetherLimit {
		nether = params.NetherLimit
	}

	return Transaction{
		From:        from,
		To:          to,
		Value:       amount,
		Nonce:       nonce,
		Nether:      nether,
		NetherPrice: params.NetherPrice,
		Data:        data,
		Time:        time.Now().Unix(),
	}, nil
}

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

// Serialize encodes Transaction to binary data
func (tx Transaction) Serialize() ([]byte, error) {
	var encoded bytes.Buffer

	enc := gob.NewEncoder(&encoded)
	if err := enc.Encode(tx); err != nil {
		return nil, err
	}

	return encoded.Bytes(), nil
}

// Deserialize decodes binary data and returns valid Transaction
func (tx *Transaction) Deserialize(data []byte) error {
	decoder := gob.NewDecoder(bytes.NewReader(data))
	return decoder.Decode(tx)
}

func (tx *Transaction) Cost() uint64 {
	return tx.Value + tx.Nether
}

func (tx *Transaction) AppendData(data []byte) {
	tx.Data = append(tx.Data, data...)
}

func (tx *Transaction) IsReward() bool {
	return bytes.Compare(tx.Data, TxRewardData) == 0
}
