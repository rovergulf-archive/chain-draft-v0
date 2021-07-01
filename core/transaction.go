package core

import (
	"bytes"
	"crypto/elliptic"
	"crypto/sha256"
	"encoding/gob"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"time"
)

const (
	TxFee   = uint64(50)
	TxLimit = uint64(1 << 10)
)

var (
	txRewardData   = []byte("reward")
	txPrefix       = []byte("tx/")
	txPrefixLength = len(txPrefix)
)

// Transaction represents a Bitcoin transaction
type Transaction struct {
	Hash     common.Hash    `json:"hash" yaml:"hash"`
	From     common.Address `json:"from" yaml:"from"`
	To       common.Address `json:"to" yaml:"to"` // destination of contract, use empty address for contract creation
	Nonce    uint64         `json:"nonce" yaml:"nonce"`
	Value    uint64         `json:"value" yaml:"value"`
	Gas      uint64         `json:"gas" yaml:"gas"`
	GasPrice uint64         `json:"gas_price" yaml:"gas_price"`
	Data     []byte         `json:"data" yaml:"data"` // contract data
	Time     int64          `json:"time" yaml:"time"`
}

// SetHash sets a hash of the transaction
func (tx *Transaction) SetHash() error {
	txCopy := *tx

	serializedCopy, err := txCopy.Serialize()
	if err != nil {
		return err
	}

	tx.Hash = sha256.Sum256(serializedCopy)
	return nil
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

// NewTransaction creates a new transaction
func NewTransaction(from, to common.Address, amount uint64, nonce uint64, data []byte) (Transaction, error) {
	if from == to {
		return Transaction{}, fmt.Errorf("transaction cannot be sent to yourself")
	}

	return Transaction{
		From:     from,
		To:       to,
		Value:    amount,
		Nonce:    nonce,
		Gas:      0,
		GasPrice: 0,
		Data:     data,
		Time:     time.Now().Unix(),
	}, nil
}

func (tx *Transaction) IsReward() bool {
	return bytes.Compare(tx.Data, txRewardData) == 0
}

func (tx *Transaction) Cost() uint64 {
	return tx.Value + TxFee
}

type SignedTx struct {
	Transaction
	Sig []byte `json:"sig" yaml:"sig"`
}

func (t SignedTx) IsAuthentic() (bool, error) {
	recoveredPubKey, err := crypto.SigToPub(t.Transaction.Hash[:], t.Sig)
	if err != nil {
		return false, err
	}

	recoveredPubKeyBytes := elliptic.Marshal(crypto.S256(), recoveredPubKey.X, recoveredPubKey.Y)
	recoveredPubKeyBytesHash := crypto.Keccak256(recoveredPubKeyBytes[1:])
	recoveredAccount := common.BytesToAddress(recoveredPubKeyBytesHash[12:])

	fmt.Println(recoveredAccount.Hex(), t.Transaction.From.Hex())
	return recoveredAccount.Hex() == t.Transaction.From.Hex(), nil
}
