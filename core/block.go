package core

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
)

var (
	blocksPrefix       = []byte("blocks/")
	blocksPrefixLength = len(blocksPrefix)
)

type BlockHeader struct {
	PrevHash  common.Hash    `json:"prev_hash" yaml:"prev_hash"`
	Hash      common.Hash    `json:"hash" yaml:"hash"`
	Root      common.Hash    `json:"root" yaml:"root"`
	Number    uint64         `json:"number" yaml:"number"`
	Timestamp int64          `json:"timestamp" yaml:"timestamp"`
	Validator common.Address `json:"validator" yaml:"validator"`
}

func (bh *BlockHeader) Validate() error {
	if bytes.Compare(bh.Hash.Bytes(), common.Hash{}.Bytes()) == 0 {
		return fmt.Errorf("invalid block hash")
	}

	if bytes.Compare(bh.Validator.Bytes(), common.Address{}.Bytes()) == 0 {
		return fmt.Errorf("invalid validator address")
	}

	if bh.Timestamp == 0 {
		return fmt.Errorf("invalid timestamp")
	}

	return nil
}

// Block represents
type Block struct {
	BlockHeader
	Transactions []SignedTx `json:"transactions" yaml:"transactions"`
}

// NewBlock creates and returns Block
func NewBlock(header BlockHeader, txs []SignedTx) *Block {
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

// HashTransactions returns a hash of the transactions in the block
func (b *Block) HashTransactions() ([]byte, error) {
	var txHashes [][]byte
	var txHash [32]byte

	for _, tx := range b.Transactions {
		txHashes = append(txHashes, tx.Hash.Bytes())
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
