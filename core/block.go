package core

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"encoding/json"
	"github.com/ethereum/go-ethereum/common"
	"gopkg.in/yaml.v2"
)

type BlockHeader struct {
}

// Block represents
type Block struct {
	PrevHash     common.Hash    `json:"prev_hash" yaml:"prev_hash"`
	Hash         common.Hash    `json:"hash" yaml:"hash"`
	Number       uint64         `json:"number" yaml:"number"`
	Nonce        uint64         `json:"nonce" yaml:"nonce"`
	Difficulty   uint64         `json:"difficulty" yaml:"difficulty"`
	Timestamp    int64          `json:"timestamp" yaml:"timestamp"`
	Miner        common.Address `json:"miner" yaml:"miner"`
	Transactions []*SignedTx    `json:"transactions" yaml:"transactions"`
}

// NewBlock creates and returns Block
func NewBlock(prev common.Hash, number uint64, nonce uint64, time int64, miner common.Address, txs []*SignedTx) *Block {
	return &Block{
		PrevHash:     prev,
		Number:       number,
		Nonce:        nonce,
		Timestamp:    time,
		Miner:        miner,
		Transactions: txs,
	}
}

// SetHash sets a hash of the block
func (b *Block) SetHash() error {
	enc, err := b.Encode()
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

// Encode serializes the block to json
func (b *Block) Encode() ([]byte, error) {
	jsonRaw, err := json.Marshal(b)
	if err != nil {
		return nil, err
	}

	return jsonRaw, nil
}

// EncodeYaml serializes the block to yaml
func (b *Block) EncodeYaml() ([]byte, error) {
	jsonRaw, err := yaml.Marshal(b)
	if err != nil {
		return nil, err
	}

	return jsonRaw, nil
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

// DecodeBlock deserializes a block from json
func DecodeBlock(jsonRaw []byte) (*Block, error) {
	var block Block

	if err := json.Unmarshal(jsonRaw, &block); err != nil {
		return nil, err
	}

	return &block, nil
}

// DecodeYAMLBlock deserializes a block from yaml
func DecodeYAMLBlock(jsonRaw []byte) (*Block, error) {
	var block Block
	if err := yaml.Unmarshal(jsonRaw, &block); err != nil {
		return nil, err
	}

	return &block, nil
}
