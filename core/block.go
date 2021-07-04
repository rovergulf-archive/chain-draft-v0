package core

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"github.com/rovergulf/rbn/core/types"
)

var (
	blocksPrefix       = []byte("blocks/")
	blocksPrefixLength = len(blocksPrefix)
)

// Block represents
type Block struct {
	types.BlockHeader
	Transactions []SignedTx `json:"transactions" yaml:"transactions"`

	size int64

	ReceivedAt int64 `json:"received_at" yaml:"received_at"`
}

// NewBlock creates and returns Block
func NewBlock(header types.BlockHeader, txs []SignedTx) *Block {
	return &Block{
		BlockHeader:  header,
		Transactions: txs,
	}
}

// SetHash sets a hash of the block
func (b *Block) SetHash() error {
	enc, err := b.Serialize()
	if err != nil {
		return err
	}

	hash := sha256.Sum256(enc)
	b.Hash = hash
	return nil
}

// Size returns encoded block value byte length
func (b *Block) Size() error {
	enc, err := b.Serialize()
	if err != nil {
		return err
	}

	hash := sha256.Sum256(enc)
	b.Hash = hash
	return nil
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

// Deserialize deserializes a block from gob encoding
func (b *Block) Deserialize(d []byte) error {
	decoder := gob.NewDecoder(bytes.NewReader(d))
	return decoder.Decode(b)
}

// DeserializeBlock deserializes a block from gob encoding
func DeserializeBlock(d []byte) (*Block, error) {
	var block Block

	decoder := gob.NewDecoder(bytes.NewReader(d))
	if err := decoder.Decode(&block); err != nil {
		return nil, err
	}

	return &block, nil
}
