package core

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"path"
	"time"
)

var genesisSymbols = map[string]bool{
	"TBB": true, // test net
	"RBN": true, // main net
}

type Genesis struct {
	ChainId     string                            `json:"chain_id" yaml:"chain_id"`
	GenesisTime time.Time                         `json:"genesis_time" yaml:"genesis_time"`
	Difficulty  uint64                            `json:"difficulty" yaml:"difficulty"`
	GasLimit    uint64                            `json:"gas_limit" yaml:"gas_limit"`
	Coinbase    common.Address                    `json:"coinbase" yaml:"coinbase"`
	Symbol      string                            `json:"symbol" yaml:"symbol"`
	ParentHash  common.Hash                       `json:"parent_hash" yaml:"parent_hash"`
	Alloc       map[common.Address]GenesisAccount `json:"alloc" yaml:"alloc"`
	ExtraData   []byte                            `json:"extra_data,omitempty" yaml:"extra_data,omitempty"`
}

type GenesisAccount struct {
	Balance uint64 `json:"balance" yaml:"balance"`
	Nonce   uint64 `json:"nonce" yaml:"nonce"`
	Auth    string `json:"auth" yaml:"auth"`
}

func (g *Genesis) Validate() error {
	if len(g.ChainId) == 0 {
		return fmt.Errorf("bad chain_id: %s", g.ChainId)
	}

	if bytes.Compare(g.Coinbase.Bytes(), common.Address{}.Bytes()) == 0 {
		return fmt.Errorf("invalid coinbase address")
	}

	if g.GenesisTime.Unix() < 0 {
		return fmt.Errorf("bad genesis timestamp: %s", g.GenesisTime)
	}

	if _, ok := genesisSymbols[g.Symbol]; !ok {
		return fmt.Errorf("bad symbol: %s", g.Symbol)
	}

	if g.GasLimit == 0 {
		return fmt.Errorf("gas limit should be more than 0")
	}

	if g.Difficulty == 0 {
		return fmt.Errorf("difficulty should be more than 0")
	}

	for addr := range g.Alloc {
		if bytes.Compare(addr.Bytes(), common.Address{}.Bytes()) == 0 {
			return fmt.Errorf("invalid alloc address: %s", addr.Hex())
		}
	}

	return nil
}

func (g *Genesis) Encode() ([]byte, error) {
	var result bytes.Buffer
	encoder := gob.NewEncoder(&result)
	if err := encoder.Encode(*g); err != nil {
		return nil, err
	}

	return result.Bytes(), nil
}

func (g *Genesis) MarshalJSON() ([]byte, error) {
	var result bytes.Buffer
	encoder := json.NewEncoder(&result)

	if err := encoder.Encode(*g); err != nil {
		return nil, err
	}

	return result.Bytes(), nil
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
	var txs []SignedTx

	for addr := range g.Alloc {
		coinbase := g.Alloc[addr]
		tx, err := NewTransaction(g.Coinbase, addr, coinbase.Balance, coinbase.Nonce, g.ExtraData)
		if err != nil {
			return nil, err
		}

		txs = append(txs, SignedTx{Transaction: tx})
	}

	header := BlockHeader{
		PrevHash:  g.ParentHash,
		Number:    0,
		Timestamp: time.Now().Unix(),
		Validator: g.Coinbase,
	}

	b := NewBlock(header, txs)
	if err := b.SetHash(); err != nil {
		return nil, err
	}

	return b, nil
}
