package core

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"encoding/json"
	"github.com/ethereum/go-ethereum/common"
)

const (
	TxFee   = uint64(50)
	TxLimit = uint64(1 << 10)
)

// Transaction represents a Bitcoin transaction
type Transaction struct {
	From     common.Address  `json:"from" yaml:"from"`
	To       *common.Address `json:"to" yaml:"to"` // destination of contract, use nil for contract creation
	Value    int64           `json:"value" yaml:"value"`
	Nonce    uint64          `json:"nonce" yaml:"nonce"`
	Gas      uint64          `json:"gas" yaml:"gas"`
	GasPrice uint64          `json:"gas_price" yaml:"gas_price"`
	Data     []byte          `json:"data" yaml:"data"` // contract data
	Time     uint64          `json:"time" yaml:"time"`
}

type SignedTx struct {
	Transaction
	Sig []byte `json:"sig" yaml:"sig"`
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

func (tx *Transaction) Serialize() ([]byte, error) {
	var encoded bytes.Buffer

	enc := gob.NewEncoder(&encoded)
	if err := enc.Encode(tx); err != nil {
		return nil, err
	}

	return encoded.Bytes(), nil
}

func DeserializeTransaction(data []byte) (*Transaction, error) {
	var transaction Transaction

	decoder := gob.NewDecoder(bytes.NewReader(data))
	if err := decoder.Decode(&transaction); err != nil {
		return nil, err
	}

	return &transaction, nil
}

func (tx Transaction) Encode() ([]byte, error) {
	return json.Marshal(tx)
}

func DecodeTransaction(data []byte) (*Transaction, error) {
	var transaction Transaction
	if err := json.Unmarshal(data, &transaction); err != nil {
		return nil, err
	}
	return &transaction, nil
}

// NewTransaction creates a new transaction
func NewTransaction(from, to common.Address, amount int64, nonce uint64, data []byte) (*Transaction, error) {
	return &Transaction{
		From:     from,
		To:       &to,
		Value:    amount,
		Nonce:    nonce,
		Gas:      0,
		GasPrice: 0,
		Data:     data,
		Time:     0,
	}, nil
}
