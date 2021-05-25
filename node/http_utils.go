package node

import (
	"github.com/rovergulf/rbn/core"
)

const (
	endpointSyncQueryKeyFromBlock = "fromBlock"
	endpointAddPeerQueryKeyIP     = "ip"
	endpointAddPeerQueryKeyPort   = "port"
	endpointAddPeerQueryKeyMiner  = "miner"
)

type StatusRes struct {
	LastHash   string              `json:"block_hash,omitempty" yaml:"last_hash,omitempty"`
	Number     int                 `json:"block_number,omitempty" yaml:"number,omitempty"`
	KnownPeers map[string]PeerNode `json:"peers_known,omitempty" yaml:"known_peers,omitempty"`
	PendingTXs []*core.Transaction `json:"pending_txs,omitempty" yaml:"pending_t_xs,omitempty"`
}

type SyncRes struct {
	Blocks []*core.Block `json:"blocks" yaml:"blocks"`
}

type AddPeerRes struct {
	Success bool   `json:"success" yaml:"success"`
	Error   string `json:"error,omitempty" yaml:"error,omitempty"`
}

type Addr struct {
	AddrList []string
}

type Block struct {
	AddrFrom string
	Block    []byte
}

type GetBlocks struct {
	AddrFrom string
}

type GetData struct {
	AddrFrom string
	Type     string
	ID       []byte
}

type Inv struct {
	AddrFrom string
	Type     string
	Items    [][]byte
}

type Tx struct {
	AddrFrom    string
	Transaction []byte
}

type Version struct {
	Version    int
	BestHeight int
	AddrFrom   string
}
