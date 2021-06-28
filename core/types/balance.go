package types

import (
	"github.com/ethereum/go-ethereum/common"
	"math/big"
)

type Balance struct {
	Address    common.Address              `json:"address" yaml:"address"`
	Code       []byte                      `json:"code" yaml:"code"`
	Balance    big.Int                     `json:"coins" yaml:"coins"`
	Nonce      uint64                      `json:"nonce" yaml:"nonce"`
	Storage    map[common.Hash]common.Hash `json:"storage" yaml:"storage"`
	PrivateKey []byte                      `json:"private_key" yaml:"private_key"`
}
