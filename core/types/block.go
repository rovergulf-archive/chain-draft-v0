package types

import (
	"bytes"
	"encoding/gob"
	"github.com/ethereum/go-ethereum/common"
)

// BlockHeader represents header part of chain block
type BlockHeader struct {
	Root      common.Hash `json:"root" yaml:"root"`
	PrevHash  common.Hash `json:"prev_hash" yaml:"prev_hash"`
	Hash      common.Hash `json:"hash" yaml:"hash"`
	Number    uint64      `json:"number" yaml:"number"`
	Timestamp int64       `json:"timestamp" yaml:"timestamp"`

	ReceiptHash common.Hash    `json:"receipts_hash" yaml:"receipts_hash"`
	TxHash      common.Hash    `json:"txs_hash" yaml:"txs_hash"`
	NetherUsed  uint64         `json:"nether_used" yaml:"nether_used"`
	Coinbase    common.Address `json:"coinbase" yaml:"coinbase"` // validator node address
}

// Serialize serializes block header with god encodig
func (bh BlockHeader) Serialize() ([]byte, error) {
	var buf bytes.Buffer
	encoder := gob.NewEncoder(&buf)
	if err := encoder.Encode(bh); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// Deserialize deserializes a block header from gob encoding
func (bh *BlockHeader) Deserialize(d []byte) error {
	decoder := gob.NewDecoder(bytes.NewReader(d))
	return decoder.Decode(bh)
}
