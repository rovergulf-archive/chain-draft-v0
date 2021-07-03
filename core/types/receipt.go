package types

import (
	"bytes"
	"encoding/gob"
	"github.com/ethereum/go-ethereum/common"
)

// Receipt represents an
type Receipt struct {
	Addr common.Address `json:"addr" yaml:"addr"`

	Balance         uint64         `json:"balance" yaml:"balance"`
	ContractAddress common.Address `json:"contract_address"`
	GasUsed         uint64         `json:"gas_used"`

	BlockHash        common.Hash `json:"blockHash,omitempty"`
	BlockNumber      uint64      `json:"block_number,omitempty"`
	TxHash           common.Hash `json:"tx_hash"`
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
