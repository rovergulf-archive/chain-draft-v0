package node

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/rovergulf/rbn/core"
)

type StatusRes struct {
	LastHash   string                   `json:"block_hash,omitempty" yaml:"last_hash,omitempty"`
	Number     uint64                   `json:"chain_length,omitempty" yaml:"chain_length,omitempty"`
	KnownPeers map[string]PeerNode      `json:"peers_known,omitempty" yaml:"known_peers,omitempty"`
	PendingTXs map[string]core.SignedTx `json:"pending_txs,omitempty" yaml:"pending_t_xs,omitempty"`
	IsMining   bool                     `json:"is_mining" yaml:"is_mining"`
	DbSize     map[string]int64         `json:"db_size" yaml:"db_size"`
}

type SyncRes struct {
	Blocks []*core.Block `json:"blocks" yaml:"blocks"`
}

type AddPeerRes struct {
	Success bool   `json:"success" yaml:"success"`
	Error   string `json:"error,omitempty" yaml:"error,omitempty"`
}

type Addr struct {
	AddrList []string `json:"addr_list" yaml:"addr_list"`
}

type Block struct {
	AddrFrom common.Address `json:"addr_from" yaml:"addr_from"`
	Block    []byte         `json:"block" yaml:"block"`
}

type GetBlocks struct {
	AddrFrom common.Address `json:"addr_from" yaml:"addr_from"`
}

type GetData struct {
	AddrFrom common.Address `json:"addr_from" yaml:"addr_from"`
	Type     string         `json:"type" yaml:"type"`
	ID       []byte         `json:"id" yaml:"id"`
}

type Inv struct {
	AddrFrom common.Address `json:"addr_from" yaml:"addr_from"`
	Type     string         `json:"type" yaml:"type"`
	Items    [][]byte       `json:"items" yaml:"items"`
}

type Tx struct {
	AddrFrom    common.Address `json:"addr_from" yaml:"addr_from"`
	Transaction []byte         `json:"transaction" yaml:"transaction"`
}

type TxAddReq struct {
	From    string `json:"from" yaml:"from"`
	FromPwd string `json:"from_pwd" yaml:"from_pwd"`
	To      string `json:"to" yaml:"to"`
	Value   uint64 `json:"value" yaml:"value"`
	Data    []byte `json:"data" yaml:"data"`
}

type Version struct {
	Version    int64          `json:"version" yaml:"version"`
	BestHeight int64          `json:"best_height" yaml:"best_height"`
	AddrFrom   common.Address `json:"addr_from" yaml:"addr_from"`
}
