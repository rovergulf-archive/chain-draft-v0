package types

import (
	"bytes"
	"encoding/gob"
	"github.com/ethereum/go-ethereum/common"
)

// Receipt is an result of confirmed transaction
type Receipt struct {
	Addr common.Address `json:"addr" yaml:"addr" yaml:"addr"`

	Balance         uint64         `json:"balance" yaml:"balance" yaml:"balance"`
	ContractAddress common.Address `json:"contract_address" yaml:"contract_address"` // TBD

	NetherUsed  uint64 `json:"nether_used" yaml:"nether_used"`
	NetherPrice uint64 `json:"nether_price" yaml:"nether_price"`

	BlockHash        common.Hash `json:"block_hash,omitempty" yaml:"block_hash"`
	BlockNumber      uint64      `json:"block_number,omitempty" yaml:"block_number"`
	TxHash           common.Hash `json:"tx_hash" yaml:"tx_hash"`
	TransactionIndex int         `json:"transaction_index" yaml:"transaction_index"`
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
