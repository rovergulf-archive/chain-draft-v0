package core

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"encoding/json"
	"time"
)

type Block struct {
	Timestamp    int64          `json:"timestamp" yaml:"timestamp"`
	Transactions []*Transaction `json:"transactions" yaml:"transactions"`
	Hash         []byte         `json:"hash" yaml:"hash"`
	PrevHash     []byte         `json:"prev_hash" yaml:"prev_hash"`
	Nonce        int            `json:"nonce" yaml:"nonce"`
	Height       int            `json:"height" yaml:"height"`
}

// NewBlock creates and returns Block
func NewBlock(txs []*Transaction, prevBlockHash []byte) *Block {
	block := &Block{
		Timestamp:    time.Now().Unix(),
		Transactions: txs,
		PrevHash:     prevBlockHash,
		Hash:         []byte{},
		Nonce:        0,
	}
	pow := NewProofOfWork(block)
	nonce, hash := pow.Run()

	block.Hash = hash[:]
	block.Nonce = nonce

	return block
}

// HashTransactions returns a hash of the transactions in the block
func (b *Block) HashTransactions() []byte {
	var txHashes [][]byte
	var txHash [32]byte

	for _, tx := range b.Transactions {
		txHashes = append(txHashes, tx.ID)
	}
	txHash = sha256.Sum256(bytes.Join(txHashes, []byte{}))

	return txHash[:]
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

// MarshalJSON serializes the block to json
func (b *Block) MarshalJSON() ([]byte, error) {
	var result bytes.Buffer
	encoder := json.NewEncoder(&result)
	if err := encoder.Encode(b); err != nil {
		return nil, err
	}

	return result.Bytes(), nil
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

// UnmarshalJSONBlock deserializes a block from json
func UnmarshalJSONBlock(jsonRaw []byte) (*Block, error) {
	var block Block

	decoder := json.NewDecoder(bytes.NewReader(jsonRaw))
	if err := decoder.Decode(&block); err != nil {
		return nil, err
	}

	return &block, nil
}
