package types

import (
	"github.com/ethereum/go-ethereum/common"
	"math/big"
)

type Receipt struct {
	Addr            common.Address `json:"addr" yaml:"addr"`
	Balance         big.Int        `json:"balance" yaml:"balance"`
	TxHash          common.Hash    `json:"tx_hash"`
	ContractAddress common.Address `json:"contract_address"`
	GasUsed         uint64         `json:"gas_used"`

	BlockHash        common.Hash `json:"blockHash,omitempty"`
	BlockNumber      *big.Int    `json:"block_number,omitempty"`
	TransactionIndex uint        `json:"transaction_index"`
}
