package types

import (
	"bytes"
	"encoding/gob"
	"github.com/ethereum/go-ethereum/common"
	"math/big"
)

// Receipt represents miner node
type Receipt struct {
	Addr common.Address `json:"addr" yaml:"addr"`

	Balance         uint64         `json:"balance" yaml:"balance"`
	TxHash          common.Hash    `json:"tx_hash"`
	ContractAddress common.Address `json:"contract_address"`
	GasUsed         uint64         `json:"gas_used"`

	BlockHash        common.Hash `json:"blockHash,omitempty"`
	BlockNumber      *big.Int    `json:"block_number,omitempty"`
	TransactionIndex uint        `json:"transaction_index"`
}

// Serialize serializes receipt
func (r Receipt) Serialize() ([]byte, error) {
	var buf bytes.Buffer
	encoder := gob.NewEncoder(&buf)
	if err := encoder.Encode(r); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// Deserialize deserializes a receipt from gob encoding
func (r *Receipt) Deserialize(d []byte) error {
	decoder := gob.NewDecoder(bytes.NewReader(d))
	return decoder.Decode(r)
}
