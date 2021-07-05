package node

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/rovergulf/rbn/core/types"
)

type StatusRes struct {
	LastHash   string                         `json:"block_hash,omitempty" yaml:"last_hash,omitempty"`
	Number     uint64                         `json:"chain_length,omitempty" yaml:"chain_length,omitempty"`
	KnownPeers map[string]PeerNode            `json:"peers_known,omitempty" yaml:"known_peers,omitempty"`
	PendingTXs map[common.Hash]types.SignedTx `json:"pending_txs,omitempty" yaml:"pending_t_xs,omitempty"`
	IsMining   bool                           `json:"is_mining" yaml:"is_mining"`
	DbSize     map[string]int64               `json:"db_size" yaml:"db_size"`
}

type TxAddReq struct {
	From    string `json:"from" yaml:"from"`
	FromPwd string `json:"from_pwd" yaml:"from_pwd"`
	To      string `json:"to" yaml:"to"`
	Value   uint64 `json:"value" yaml:"value"`
	Data    []byte `json:"data" yaml:"data"`
}
