package core

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"github.com/ethereum/go-ethereum/common"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"path"
	"time"
)

type Genesis struct {
	ChainId     string                            `json:"chain_id" yaml:"chain_id"`
	GenesisTime time.Time                         `json:"genesis_time" yaml:"genesis_time"`
	Difficulty  uint64                            `json:"difficulty" yaml:"difficulty"`
	GasLimit    uint64                            `json:"gas_limit" yaml:"gas_limit"`
	Coinbase    common.Address                    `json:"coinbase" yaml:"coinbase"`
	Symbol      string                            `json:"symbol" yaml:"symbol"`
	ParentHash  common.Hash                       `json:"parent_hash,omitempty" yaml:"parent_hash,omitempty"`
	Alloc       map[common.Address]GenesisAccount `json:"alloc" yaml:"alloc"`
	ExtraData   []byte                            `json:"extra_data,omitempty" yaml:"extra_data,omitempty"`
}

type GenesisAccount struct {
	Balance int64  `json:"balance" yaml:"balance"`
	Nonce   uint64 `json:"nonce" yaml:"nonce"`
}

func (g *Genesis) Encode() ([]byte, error) {
	var result bytes.Buffer
	encoder := gob.NewEncoder(&result)
	if err := encoder.Encode(*g); err != nil {
		return nil, err
	}

	return result.Bytes(), nil
}

func (g *Genesis) Decode(data []byte) error {
	decoder := gob.NewDecoder(bytes.NewReader(data))
	return decoder.Decode(g)
}

func (g *Genesis) MarshalJSON() ([]byte, error) {
	var result bytes.Buffer
	encoder := json.NewEncoder(&result)

	if err := encoder.Encode(*g); err != nil {
		return nil, err
	}

	return result.Bytes(), nil
}

func (g *Genesis) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, g)
}

func loadGenesisFromFile(filename string) (*Genesis, error) {
	var g Genesis

	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	ext := path.Ext(filename)
	if ext == "json" {
		if err := json.Unmarshal(data, &g); err != nil {
			return nil, err
		}
	} else {
		if err := yaml.Unmarshal(data, &g); err != nil {
			return nil, err
		}
	}

	return &g, nil
}

func NewGenesisBlock(g *Genesis) (*Block, error) {
	var txs []*SignedTx

	for addr := range g.Alloc {
		coinbase := g.Alloc[addr]
		tx, err := NewTransaction(g.Coinbase, addr, coinbase.Balance, coinbase.Nonce, g.ExtraData)
		if err != nil {
			return nil, err
		}

		stx := &SignedTx{Transaction: *tx}

		txs = append(txs, stx)
	}

	b := NewBlock(g.ParentHash, 0, 0, time.Now().Unix(), g.Coinbase, txs)

	if err := b.SetHash(); err != nil {
		return nil, err
	}

	return b, nil
}
