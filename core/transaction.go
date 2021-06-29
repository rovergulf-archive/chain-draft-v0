package core

import (
	"bytes"
	"crypto/elliptic"
	"crypto/sha256"
	"encoding/gob"
	"encoding/json"
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
	From     common.Address  `json:"from" yaml:"from"`
	To       *common.Address `json:"to" yaml:"to"` // destination of contract, use nil for contract creation
	Value    uint64          `json:"value" yaml:"value"`
	Nonce    uint64          `json:"nonce" yaml:"nonce"`
	Gas      uint64          `json:"gas" yaml:"gas"`
	GasPrice uint64          `json:"gas_price" yaml:"gas_price"`
	Data     []byte          `json:"data" yaml:"data"` // contract data
	Time     int64           `json:"time" yaml:"time"`
}

type SignedTx struct {
	Transaction
	Sig []byte `json:"sig" yaml:"sig"`
}

// Hash returns a hash of the transaction
func (tx *Transaction) Hash() (common.Hash, error) {
	txCopy := *tx

	serializedCopy, err := txCopy.Serialize()
	if err != nil {
		return common.Hash{}, err
	}

	return sha256.Sum256(serializedCopy), nil
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

func (tx Transaction) MarshalJSON() ([]byte, error) {
	var res bytes.Buffer
	encoder := json.NewEncoder(&res)
	if err := encoder.Encode(tx); err != nil {
		return nil, err
	}

	return res.Bytes(), nil
}

func (tx *Transaction) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, tx)
}

// NewTransaction creates a new transaction
func NewTransaction(from, to common.Address, amount uint64, nonce uint64, data []byte) (*Transaction, error) {
	if from == to {
		return nil, fmt.Errorf("transaction cannot be sent to yourself")
	}

	return &Transaction{
		From:     from,
		To:       &to,
		Value:    amount,
		Nonce:    nonce,
		Gas:      0,
		GasPrice: 0,
		Data:     data,
		Time:     time.Now().Unix(),
	}, nil
}

func (t SignedTx) IsAuthentic() (bool, error) {
	txHash, err := t.Transaction.Hash()
	if err != nil {
		return false, err
	}

	recoveredPubKey, err := crypto.SigToPub(txHash[:], t.Sig)
	if err != nil {
		return false, err
	}

	recoveredPubKeyBytes := elliptic.Marshal(crypto.S256(), recoveredPubKey.X, recoveredPubKey.Y)
	recoveredPubKeyBytesHash := crypto.Keccak256(recoveredPubKeyBytes[1:])
	recoveredAccount := common.BytesToAddress(recoveredPubKeyBytesHash[12:])

	return recoveredAccount.Hex() == t.From.Hex(), nil
}

func (tx *Transaction) IsReward() bool {
	return bytes.Compare(tx.Data, txRewardData) == 0
}

func (tx *Transaction) Cost() uint64 {
	return tx.Value + TxFee
}
