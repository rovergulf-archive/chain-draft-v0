package params

import "github.com/ethereum/go-ethereum/common"

const (
	OpenDevNetworkId = "dev_rbn"
	MainNetworkId    = "rbn"
)

const (
	OreLimit = 4800
)

type ChainConfig struct {
	ChainId string `json:"chain_id" yaml:"chain_id"`

	// RBN Treasurer genesis blocks
	MainBlock common.Hash `json:"main_block" yaml:"main_block"`
}
