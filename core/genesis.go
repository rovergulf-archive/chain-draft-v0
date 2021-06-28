package core

import (
	"encoding/json"
	"github.com/ethereum/go-ethereum/common"
	"github.com/rovergulf/rbn/core/types"
	"io/ioutil"
	"time"
)

type Genesis struct {
	ChainId     string                           `json:"chain_id" yaml:"chain_id"`
	GenesisTime time.Time                        `json:"genesis_time" yaml:"genesis_time"`
	Difficulty  uint64                           `json:"difficulty" yaml:"difficulty"`
	GasLimit    uint64                           `json:"gas_limit" yaml:"gas_limit"`
	Coinbase    common.Address                   `json:"coinbase" yaml:"coinbase"`
	Symbol      string                           `json:"symbol" yaml:"symbol"`
	ParentHash  common.Hash                      `json:"parent_hash" yaml:"parent_hash"`
	ExtraData   []byte                           `json:"extra_data" yaml:"extra_data"`
	Alloc       map[common.Address]types.Balance `json:"alloc" yaml:"alloc"`
}

func (g *Genesis) MarshalJSON() ([]byte, error) {
	return json.Marshal(g)
}

func (g *Genesis) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, g)
}

func loadGenesisFromFile(filename string) (*Genesis, error) {
	var g *Genesis

	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(data, g); err != nil {
		return nil, err
	}

	return g, nil
}

func NewGenesisBlock(g *Genesis) (*Block, error) {
	var txs []*SignedTx

	for addr := range g.Alloc {
		coinbase := g.Alloc[addr]
		tx, err := NewTransaction(g.Coinbase, addr, coinbase.Balance.Int64(), coinbase.Nonce, g.ExtraData)
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
