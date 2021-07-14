package types

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"github.com/ethereum/go-ethereum/common"
)

// BlockHeader represents header part of chain block
// must be created with struct literal
type BlockHeader struct {
	Root        common.Hash    `json:"root" yaml:"root"`
	PrevHash    common.Hash    `json:"prev_hash" yaml:"prev_hash"`
	BlockHash   common.Hash    `json:"block_hash" yaml:"block_hash"`
	Number      uint64         `json:"number" yaml:"number"`
	Timestamp   int64          `json:"timestamp" yaml:"timestamp"`
	ReceiptHash common.Hash    `json:"receipts_hash" yaml:"receipts_hash"`
	TxHash      common.Hash    `json:"txs_hash" yaml:"txs_hash"`
	NetherUsed  uint64         `json:"nether_used" yaml:"nether_used"`
	Coinbase    common.Address `json:"coinbase" yaml:"coinbase"` // author node address
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

// Deserialize deserializes binary data to BlockHeader
func (bh *BlockHeader) Deserialize(d []byte) error {
	decoder := gob.NewDecoder(bytes.NewReader(d))
	return decoder.Decode(bh)
}

// Hash returns a hash of the block header
func (bh *BlockHeader) Hash() ([]byte, error) {
	enc, err := bh.Serialize()
	if err != nil {
		return nil, err
	}

	hash := sha256.Sum256(enc)
	return hash[:], nil
}

// NewBlock creates and returns Block
func NewBlock(header BlockHeader, txs []*SignedTx) *Block {
	return &Block{
		BlockHeader:  header,
		Transactions: txs,
	}
}

// Block represents BlockChain state change interface
type Block struct {
	BlockHeader
	Transactions []*SignedTx `json:"transactions" yaml:"transactions"`

	TxHashes []common.Hash `json:"tx_hashes" yaml:"tx_hashes"`

	ReceivedAt int64 `json:"received_at" yaml:"received_at"`
}

// Hash returns a hash of the block
func (b *Block) Hash() ([]byte, error) {
	enc, err := b.Serialize()
	if err != nil {
		return nil, err
	}

	hash := sha256.Sum256(enc)
	return hash[:], nil
}

// Size returns encoded block value byte length
func (b *Block) Size() (int, error) {
	enc, err := b.Serialize()
	if err != nil {
		return 0, err
	}

	return len(enc), nil
}

// HashTransactions returns a hash of the transactions in the block
func (b *Block) HashTransactions() ([]byte, error) {
	var txHashes [][]byte
	var txHash [32]byte

	for _, tx := range b.Transactions {
		hash, err := tx.Hash()
		if err != nil {
			return nil, err
		}
		txHashes = append(txHashes, hash)
	}
	txHash = sha256.Sum256(bytes.Join(txHashes, []byte{}))

	return txHash[:], nil
}

// Serialize serializes the block
func (b *Block) Serialize() ([]byte, error) {
	var result bytes.Buffer
	encoder := gob.NewEncoder(&result)
	if err := encoder.Encode(b); err != nil {
		return nil, err
	}

	return result.Bytes(), nil
}

// Deserialize deserializes binary data to block
func (b *Block) Deserialize(d []byte) error {
	decoder := gob.NewDecoder(bytes.NewReader(d))
	return decoder.Decode(b)
}
