package core

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"github.com/ethereum/go-ethereum/common"
	"github.com/rovergulf/rbn/core/types"
	"github.com/rovergulf/rbn/params"
	"time"
)

var genesisSymbols = map[string]bool{
	"RBN":  true, // dev unit
	"Gulf": true,
}

type Genesis struct {
	ChainId     string         `json:"chain_id" yaml:"chain_id"`
	GenesisTime int64          `json:"genesis_time" yaml:"genesis_time"`
	GasLimit    uint64         `json:"gas_limit" yaml:"gas_limit"`
	Nonce       uint64         `json:"nonce" yaml:"nonce"`
	Coinbase    common.Address `json:"coinbase" yaml:"coinbase"`
	Symbol      string         `json:"symbol" yaml:"symbol"`
	Units       string         `json:"units" yaml:"units"`
	ParentHash  common.Hash    `json:"parent_hash" yaml:"parent_hash"`
	Alloc       genesisAlloc   `json:"alloc" yaml:"alloc"`
	ExtraData   []byte         `json:"extra_data,omitempty" yaml:"extra_data,omitempty"`
}

// DevNetGenesis returns default Genesis for development and testing network
func DevNetGenesis() *Genesis {
	return &Genesis{
		ChainId:     params.OpenDevNetworkId,
		GenesisTime: 1625300000,
		GasLimit:    4800,
		Nonce:       0,
		Coinbase:    common.Address{},
		Symbol:      "DevCoin",
		Units:       "Wei", // in favor of Etherium native denomination
		ParentHash:  common.Hash{},
		Alloc:       developerNetAlloc(),
		ExtraData:   []byte("0x00000000000000000000000000000000000000"),
	}
}

// DefaultMainNetGenesis returns default Genesis for main Rovergulf Blockchain Network
func DefaultMainNetGenesis() *Genesis {
	return &Genesis{
		ChainId:     params.MainNetworkId,
		GenesisTime: 1625319335,
		GasLimit:    4800,
		Nonce:       0,
		Coinbase:    common.Address{},
		Symbol:      "Gulf", //
		Units:       "Ore",  //
		ParentHash:  common.Hash{},
		Alloc:       defaultMainNetAlloc(),
		ExtraData:   []byte("0x00000000000000000000000000000000000000"),
	}
}

// Serialize encodes genesis with gob encoding
func (g Genesis) Serialize() ([]byte, error) {
	var result bytes.Buffer
	encoder := gob.NewEncoder(&result)
	if err := encoder.Encode(g); err != nil {
		return nil, err
	}

	return result.Bytes(), nil
}

// Deserialize decodes gob value to genesis
func (g *Genesis) Deserialize(data []byte) error {
	decoder := gob.NewDecoder(bytes.NewReader(data))
	return decoder.Decode(g)
}

func (g *Genesis) MarshalJSON() ([]byte, error) {
	return json.Marshal(*g)
}

func (g *Genesis) ToBlock() (*Block, error) {
	var txs []SignedTx

	for addr := range g.Alloc {
		coinbase := g.Alloc[addr]
		tx, err := NewTransaction(g.Coinbase, addr, coinbase.Balance, 0, g.ExtraData)
		if err != nil {
			return nil, err
		}

		txs = append(txs, SignedTx{Transaction: tx})
	}

	header := types.BlockHeader{
		PrevHash:  g.ParentHash,
		Number:    g.Nonce,
		Timestamp: time.Now().Unix(),
		Coinbase:  g.Coinbase,
	}

	b := NewBlock(header, txs)
	if err := b.SetHash(); err != nil {
		return nil, err
	}

	return b, nil
}

func genesisByNetworkId(networkId string) *Genesis {
	switch networkId {
	case params.OpenDevNetworkId:
		return DevNetGenesis()
	case params.MainNetworkId:
		return DefaultMainNetGenesis()
	default:
		return DefaultMainNetGenesis()
	}
}
